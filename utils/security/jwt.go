package security

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// IssuedAtMsClaim is when a control, user or pre-auth JWT was issued, in milliseconds, as iat only has seconds and a JWT
// issued before its identity's credentials changed is refused
const IssuedAtMsClaim = "iat_ms"

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

// signJWT is a JWT of subject from issuer for audience, with claims and when it was issued in milliseconds, and when
// it expires, after lifetime
func (config *SecurityConfig) signJWT(subject, issuer, audience string, lifetime time.Duration, claims map[string]any) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(lifetime)
	claims[IssuedAtMsClaim] = now.UnixMilli()
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

func (config *SecurityConfig) GenerateControlJWTToken(userName string, role string) (string, time.Time, error) {
	return config.signJWT(userName, ControlIssuer, ControlAudience, time.Hour, map[string]any{"role": role})
}

// VerifyControlJWTToken is the control admin name and role a control JWT is for, and when it was issued in milliseconds
func (config *SecurityConfig) VerifyControlJWTToken(tokenStr string) (string, string, int64, error) {
	// Verify and parse in one step
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtControlValidator)
	if err != nil {
		return "", "", 0, err
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", 0, err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", 0, err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", 0, err
	}

	if issuer != ControlIssuer {
		return "", "", 0, fmt.Errorf("invalid issuer: expected %s, got %s", ControlIssuer, issuer)
	}

	issuedAtMs, err := verifiedJWT.NumberClaim(IssuedAtMsClaim)
	if err != nil {
		return "", "", 0, err
	}
	return subject, string(role), int64(issuedAtMs), nil
}

////
// USER JWT
////

func (config *SecurityConfig) GenerateUserJWTToken(subdomain string, identityName string, role string) (string, time.Time, error) {
	return config.signJWT(identityName, subdomain, UserAudience, time.Hour, map[string]any{"role": role})
}

// VerifyUserJWTToken is the site, identity name and role a user JWT is for, and when it was issued in milliseconds
func (config *SecurityConfig) VerifyUserJWTToken(tokenStr string) (string, string, string, int64, error) {

	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtUserValidator)
	if err != nil {
		slog.Info("Failed to VerifyMACAndDecode user JWT token ", "error", err)
		return "", "", "", 0, err
	}

	identityName, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", "", 0, err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", "", 0, err
	}

	subdomain, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", "", 0, err
	}

	issuedAtMs, err := verifiedJWT.NumberClaim(IssuedAtMsClaim)
	if err != nil {
		return "", "", "", 0, err
	}

	return subdomain, identityName, string(role), int64(issuedAtMs), nil
}

///
// PREAUTH JWT (For TOTP flow)
///

func (config *SecurityConfig) GeneratePreAuthJWTToken(subdomain string, deviceName string) (string, time.Time, error) {
	// Issued by the site, which binds the password step to it
	return config.signJWT(deviceName, subdomain, PreAuthAudience, 5*time.Minute, map[string]any{})
}

// VerifyPreAuthJWTToken is the identity name a pre-auth JWT for the site is for, and when it was issued in milliseconds
func (config *SecurityConfig) VerifyPreAuthJWTToken(tokenStr string, subdomain string) (string, int64, error) {
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtPreAuthValidator)
	if err != nil {
		return "", 0, err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil || issuer != subdomain {
		return "", 0, fmt.Errorf("pre-auth token is for a different site")
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", 0, err
	}

	issuedAtMs, err := verifiedJWT.NumberClaim(IssuedAtMsClaim)
	if err != nil {
		return "", 0, err
	}

	return subject, int64(issuedAtMs), nil
}
