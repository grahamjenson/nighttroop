package security

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// AUDIENCE STRINGS
var ControlAudience = "control"
var UserAudience = "user"
var PreAuthAudience = "preauth"

type JWTStuff struct {
	jwtMac              jwt.MAC
	jwtControlValidator *jwt.Validator
	jwtUserValidator    *jwt.Validator
	jwtPreAuthValidator *jwt.Validator
}

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

////
// CONTROL JWT
////

var ControlIssuer = "control"

func (config *SecurityConfig) GenerateControlJWTToken(userName string, role string) (string, time.Time, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 1)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &userName,
		Audience:  &ControlAudience,
		ExpiresAt: &expiresAt,
		Issuer:    &ControlIssuer,
		IssuedAt:  &now,
		CustomClaims: map[string]any{
			"role": role,
		},
	})
	if err != nil {
		return "", time.Time{}, err
	}

	// 3. Sign and return
	tokenString, err := config.jwtMac.ComputeMACAndEncode(rawJWT)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (config *SecurityConfig) VerifyControlJWTToken(tokenStr string) (string, string, error) {
	// Verify and parse in one step
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtControlValidator)
	if err != nil {
		return "", "", err
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", err
	}

	if issuer != ControlIssuer {
		return "", "", fmt.Errorf("invalid issuer: expected %s, got %s", ControlIssuer, issuer)
	}
	return subject, string(role), nil
}

////
// USER JWT
////

func (config *SecurityConfig) GenerateUserJWTToken(subdomain string, identityName string, role string) (string, time.Time, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 1)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &identityName,
		Audience:  &UserAudience,
		Issuer:    &subdomain,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
		CustomClaims: map[string]any{
			"role": role,
		},
	})
	if err != nil {
		return "", time.Time{}, err
	}

	// 3. Sign and return
	tokenString, err := config.jwtMac.ComputeMACAndEncode(rawJWT)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (config *SecurityConfig) VerifyUserJWTToken(tokenStr string) (string, string, string, error) {

	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtUserValidator)
	if err != nil {
		slog.Error("Failed to VerifyMACAndDecode user JWT token ", "error", err)
		return "", "", "", err
	}

	identityName, err := verifiedJWT.Subject()
	if err != nil {
		return "", "", "", err
	}

	role, err := verifiedJWT.StringClaim("role")
	if err != nil {
		return "", "", "", err
	}

	subdomain, err := verifiedJWT.Issuer()
	if err != nil {
		return "", "", "", err
	}

	return subdomain, identityName, string(role), nil
}

///
// PREAUTH JWT (For TOTP flow)
///

func (config *SecurityConfig) GeneratePreAuthJWTToken(subdomain string, deviceName string) (string, time.Time, error) {

	now := time.Now()

	expiresAt := now.Add(time.Minute * 5)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &deviceName,
		Issuer:    &subdomain, // binds the password step to this site
		Audience:  &PreAuthAudience,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
	})
	if err != nil {
		return "", time.Time{}, err
	}

	tokenString, err := config.jwtMac.ComputeMACAndEncode(rawJWT)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (config *SecurityConfig) VerifyPreAuthJWTToken(tokenStr string, subdomain string) (string, error) {
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtPreAuthValidator)
	if err != nil {
		return "", err
	}

	issuer, err := verifiedJWT.Issuer()
	if err != nil || issuer != subdomain {
		return "", fmt.Errorf("pre-auth token is for a different site")
	}

	subject, err := verifiedJWT.Subject()
	if err != nil {
		return "", err
	}

	return subject, nil
}
