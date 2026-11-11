package security

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// TOTPIssuer names Nighttroop in an authenticator app
const TOTPIssuer = "Nighttroop"

// TOTPURI is the otpauth:// URI a QR code holds for an authenticator app to add secret, as an account named name at
// siteURL's host, so each server's and site's logins are told apart. An app cannot add one from the bare secret.
func TOTPURI(name, siteURL, secret string) string {
	account := name
	if u, err := url.Parse(siteURL); err == nil && u.Host != "" {
		account += "@" + u.Host
	}
	query := url.Values{"secret": {secret}, "issuer": {TOTPIssuer}}
	return (&url.URL{Scheme: "otpauth", Host: "totp", Path: "/" + TOTPIssuer + ":" + account, RawQuery: query.Encode()}).String()
}

// TOTPGenerate makes a TOTP secret for an identity of the site, "" for a control admin, encrypted for that site
func (cfg *SecurityConfig) TOTPGenerate(siteID, accountName string) (totpSecret, totpEncSecret string, err error) {
	totpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TOTPIssuer,
		AccountName: fmt.Sprintf("%s@%s", accountName, cfg.HostName),
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot generate TOTP key: %w", err)
	}

	totpSecret = totpKey.Secret()

	if cfg.IsDevOrTest() {
		totpSecret = *cfg.DevelopmentDangerousHardcodedAdminTOTPSecret
	}

	encTOTPSecret, err := cfg.EncryptFor(PurposeTOTP, siteID, totpSecret)
	if err != nil {
		return "", "", fmt.Errorf("cannot encrypt TOTP secret: %w", err)
	}

	return totpSecret, encTOTPSecret, nil
}

func (cfg *SecurityConfig) GeneratePassword() (password, hashedPassword string, err error) {
	if cfg.IsDevOrTest() {
		// We are going to override this stuff in dev, just to make it easier
		password = *cfg.DevelopmentDangerousHardcodedAdminPassword
		hashedPassword = HashPassword(password, bcrypt.MinCost)
		return password, hashedPassword, nil
	}
	password, hashedPassword, err = SecureCredentials(bcrypt.DefaultCost)

	if err != nil {
		return "", "", fmt.Errorf("cannot generate secure credentials: %w", err)
	}
	return password, hashedPassword, nil
}

// totpPeriod is the seconds each TOTP code is for, as authenticator apps make them
const totpPeriod = 30

// totpCode is a TOTP code as an authenticator app shows it, six digits
var totpCode = regexp.MustCompile(`^[0-9]{6}$`)

// TOTPValidate is whether token is the code for secret now. A token that is not six digits is refused like a wrong
// code, before the secret is used. It keeps no record of the code, so it is only for checking a secret was saved, not
// for signing in, which is TOTPSignIn.
func (cfg *SecurityConfig) TOTPValidate(token, secret string) bool {
	return cfg.devTOTPCode(token, secret) || len(totpSteps(token, secret)) > 0
}

// TOTPSignIn is whether token is the code for secret now, for identity (one key per site or control identity) to
// sign in with, and not of a time step identity has signed in with already, or before. So a code seen as it was typed
// cannot sign in again in the 90 seconds it is valid for, even after its owner has used it.
func (cfg *SecurityConfig) TOTPSignIn(identity, token, secret string) bool {
	// The dev code is not one time: every test signs in with it
	if cfg.devTOTPCode(token, secret) {
		return true
	}
	cfg.totpMu.Lock()
	defer cfg.totpMu.Unlock()
	last, used := cfg.totpUsedSteps.Get(identity)
	for _, step := range totpSteps(token, secret) {
		if !used || step > last {
			cfg.totpUsedSteps.Add(identity, step)
			return true
		}
	}
	return false
}

// devTOTPCode is whether token is the fixed code a dev or test server takes for the dev secret. The dev secret also
// takes the codes it makes, so a script that makes codes from the secret it is shown runs there as on staging
func (cfg *SecurityConfig) devTOTPCode(token, secret string) bool {
	return cfg.IsDevOrTest() && secret == *cfg.DevelopmentDangerousHardcodedAdminTOTPSecret && token == *cfg.DevelopmentDangerousHardcodedAdminTOTPCode
}

// totpSteps are the 30 second time steps, the one now and each next to it as totp.Validate allows for a clock's
// drift, whose code for secret is token. A token that is not six digits has none.
func totpSteps(token, secret string) []uint64 {
	if !totpCode.MatchString(token) {
		return nil
	}
	now := uint64(time.Now().Unix() / totpPeriod)
	var steps []uint64
	for _, step := range []uint64{now - 1, now, now + 1} {
		if hotp.Validate(token, step, secret) {
			steps = append(steps, step)
		}
	}
	return steps
}
