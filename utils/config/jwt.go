package config

import (
	"fmt"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// AUDIENCE STRINGS
var ControlAudience = "control"
var UserAudience = "user"

// //
// Creating the primitives
// //
func (config *ServerConfig) createJWTPrimitives() error {

	mac, err := jwt.NewMAC(config.JWTKeySet)
	if err != nil {
		return fmt.Errorf("failed to create JWT MAC primitive: %w", err)
	}

	controlValidator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
		ExpectedAudience:      &ControlAudience,
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT control validator: %w", err)
	}

	userValidator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
		ExpectedAudience:      &UserAudience,
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT user validator: %w", err)
	}

	config.jwtMac = mac
	config.jwtControlValidator = controlValidator
	config.jwtUserValidator = userValidator

	return nil
}

////
// CONTROL JWT
////

func (config *ServerConfig) GenerateControlJWTToken(userName string) (string, time.Time, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 1)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &userName,
		Audience:  &ControlAudience,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
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

func (config *ServerConfig) VerifyControlJWTToken(tokenStr string) (string, error) {
	// Verify and parse in one step
	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtControlValidator)
	if err != nil {
		return "", err
	}

	return verifiedJWT.Subject()
}

////
// USER JWT
////

func (config *ServerConfig) GenerateUserJWTToken(deviceName string) (string, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 1)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &deviceName,
		Audience:  &UserAudience,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
	})
	if err != nil {
		return "", err
	}

	// 3. Sign and return
	return config.jwtMac.ComputeMACAndEncode(rawJWT)
}

func (config *ServerConfig) VerifyUserJWTToken(tokenStr string) (string, error) {

	verifiedJWT, err := config.jwtMac.VerifyMACAndDecode(tokenStr, config.jwtUserValidator)
	if err != nil {
		return "", err
	}
	return verifiedJWT.Subject()
}
