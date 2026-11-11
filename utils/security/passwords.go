package security

import (
	"fmt"
	"net/url"
	"regexp"

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

func (cfg *SecurityConfig) TOTPGenerate(accountName string) (totpSecret, totpEncSecret string, err error) {
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

	encTOTPSecret, err := cfg.Encrypt(totpSecret)
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

// totpCode is a TOTP code as an authenticator app shows it, six digits
var totpCode = regexp.MustCompile(`^[0-9]{6}$`)

// TOTPValidate is whether token is the code for secret now. A token that is not six digits is refused like a wrong
// code, before the secret is used.
func (cfg *SecurityConfig) TOTPValidate(token, secret string) bool {
	if !totpCode.MatchString(token) {
		return false
	}
	// The dev secret also takes the code it makes now, so a script that makes codes from the secret it is shown runs
	// on a dev or test server as it does on staging
	if cfg.IsDevOrTest() && secret == *cfg.DevelopmentDangerousHardcodedAdminTOTPSecret && token == *cfg.DevelopmentDangerousHardcodedAdminTOTPCode {
		return true
	}
	return totp.Validate(token, secret)
}
