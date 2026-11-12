package security

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// IdentityIDClaim and CredentialsVersionClaim are the ID of the identity a control, user or pre-auth JWT was issued to,
// and the version of its credentials then. One not matching its identity is refused, so changing its password or TOTP
// signs out every JWT issued before, however close in time, and one for a deleted identity fails its namesake.
const IdentityIDClaim = "iid"
const CredentialsVersionClaim = "cv"

// Credentialed is a control or site identity, which a JWT is issued to
type Credentialed interface {
	GetID() int64
	GetCredentialsVersion() int64
}

// Credentials are the ID of the identity a JWT was issued to, and the version of its credentials then
type Credentials struct {
	IdentityID int64
	Version    int64
}

// CheckCurrent refuses credentials from a JWT that are not identity's current ones
func (c Credentials) CheckCurrent(identity Credentialed) error {
	if c.IdentityID != identity.GetID() || c.Version != identity.GetCredentialsVersion() {
		return fmt.Errorf("JWT is not for the identity's current credentials")
	}
	return nil
}

// credentialsClaims are the credentials a verified JWT was issued with
func credentialsClaims(verifiedJWT *jwt.VerifiedJWT) (Credentials, error) {
	identityID, err := verifiedJWT.NumberClaim(IdentityIDClaim)
	if err != nil {
		return Credentials{}, err
	}
	version, err := verifiedJWT.NumberClaim(CredentialsVersionClaim)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{IdentityID: int64(identityID), Version: int64(version)}, nil
}

// AUDIENCE STRINGS
var ControlAudience = "control"
var UserAudience = "user"
var PreAuthAudience = "preauth"

// //
// Creating the primitives
// //
func (config *SecurityConfig) createJWTPrimitives() error {

	mac, err := jwt.NewMAC(config.JWTKeySet)
	if err != nil {
		return fmt.Errorf("failed to create JWT MAC primitive: %w", err)
	}

	controlValidator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
		ExpectedAudience:      &ControlAudience,
		IgnoreIssuer:          true, // Control tokens can be issued by any subdomain
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT control validator: %w", err)
	}

	userValidator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
		ExpectedAudience:      &UserAudience,
		IgnoreIssuer:          true, // We validate the subdomain externally
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT user validator: %w", err)
	}

	preAuthValidator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
		ExpectedAudience:      &PreAuthAudience,
		IgnoreIssuer:          true, // VerifyPreAuthJWTToken checks the subdomain
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT pre-auth validator: %w", err)
	}

	config.jwtMac = mac
	config.jwtControlValidator = controlValidator
	config.jwtUserValidator = userValidator
	config.jwtPreAuthValidator = preAuthValidator

	return nil
}

// signJWT is a JWT of subject from issuer for audience, with claims and identity's credentials, and when it expires,
// after lifetime
func (config *SecurityConfig) signJWT(subject, issuer, audience string, lifetime time.Duration, identity Credentialed, claims map[string]any) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(lifetime)
	claims[IdentityIDClaim] = identity.GetID()
	claims[CredentialsVersionClaim] = identity.GetCredentialsVersion()
	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{Subject: &subject, Issuer: &issuer, Audience: &audience,
		ExpiresAt: &expiresAt, IssuedAt: &now, CustomClaims: claims})
	if err != nil {
		return "", time.Time{}, err
	}
	tokenString, err := config.jwtMac.ComputeMACAndEncode(rawJWT)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

////
// CONTROL JWT
////

var ControlIssuer = "control"

func (config *SecurityConfig) GenerateControlJWTToken(identity Credentialed, userName string, role string) (string, time.Time, error) {
	return config.signJWT(userName, ControlIssuer, ControlAudience, time.Hour, identity, map[string]any{"role": role})
}

// VerifyControlJWTToken is the control admin name, role and credentials a control JWT is for
func (config *SecurityConfig) VerifyControlJWTToken(tokenStr string) (string, string, Credentials, error) {
	// Verify and parse in one step
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtControlValidator)
	if err != nil {
		return "", "", Credentials{}, err
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", Credentials{}, err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", Credentials{}, err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", Credentials{}, err
	}

	if issuer != ControlIssuer {
		return "", "", Credentials{}, fmt.Errorf("invalid issuer: expected %s, got %s", ControlIssuer, issuer)
	}

	credentials, err := credentialsClaims(verifiedJWT)
	return subject, string(role), credentials, err
}

////
// USER JWT
////

func (config *SecurityConfig) GenerateUserJWTToken(identity Credentialed, subdomain string, identityName string, role string) (string, time.Time, error) {
	return config.signJWT(identityName, subdomain, UserAudience, time.Hour, identity, map[string]any{"role": role})
}

// VerifyUserJWTToken is the site, identity name, role and credentials a user JWT is for
func (config *SecurityConfig) VerifyUserJWTToken(tokenStr string) (string, string, string, Credentials, error) {

	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtUserValidator)
	if err != nil {
		slog.Info("Failed to VerifyMACAndDecode user JWT token ", "error", err)
		return "", "", "", Credentials{}, err
	}

	identityName, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", "", Credentials{}, err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", "", Credentials{}, err
	}

	subdomain, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", "", Credentials{}, err
	}

	credentials, err := credentialsClaims(verifiedJWT)
	return subdomain, identityName, string(role), credentials, err
}

///
// PREAUTH JWT (For TOTP flow)
///

func (config *SecurityConfig) GeneratePreAuthJWTToken(identity Credentialed, subdomain string, deviceName string) (string, time.Time, error) {
	// Issued by the site, which binds the password step to it
	return config.signJWT(deviceName, subdomain, PreAuthAudience, 5*time.Minute, identity, map[string]any{})
}

// VerifyPreAuthJWTToken is the identity name and credentials a pre-auth JWT for the site is for
func (config *SecurityConfig) VerifyPreAuthJWTToken(tokenStr string, subdomain string) (string, Credentials, error) {
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtPreAuthValidator)
	if err != nil {
		return "", Credentials{}, err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil || issuer != subdomain {
		return "", Credentials{}, fmt.Errorf("pre-auth token is for a different site")
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", Credentials{}, err
	}

	credentials, err := credentialsClaims(verifiedJWT)
	return subject, credentials, err
}
