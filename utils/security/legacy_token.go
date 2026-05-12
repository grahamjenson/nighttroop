package security

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/consts"
)

///////
// Legacy Token
///////

func (cfg *SecurityConfig) CreateLegacyTokenForDevice(subdomain, deviceName, password string) (string, error) {
	tokenStr := fmt.Sprintf("%s:%s:%s", subdomain, deviceName, password)
	encryptedToken, err := cfg.Encrypt(tokenStr)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}
	return encryptedToken, nil
}

type DeviceGetter func(ctx context.Context, name string) (deviceName string, password string, totpSecret string, role string, err error)

func (cfg *SecurityConfig) VerifyLegacyTokenForDevice(tokenStr string, deviceSubdomain string, deviceGetter DeviceGetter) (string, string, error) {
	decryptedToken, err := cfg.Decrypt(tokenStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	var tokenSubdomain, tokenName, tokenPassword string
	_, err = fmt.Sscanf(decryptedToken, "%s:%s:%s", &tokenSubdomain, &tokenName, &tokenPassword)
	if err != nil {
		return "", "", fmt.Errorf("invalid token format: %w", err)
	}

	deviceName, devicePassword, deviceTotpSecret, deviceRole, err := deviceGetter(context.TODO(), tokenName)

	if err != nil {
		return "", "", fmt.Errorf("device not found: %w", err)
	}

	if deviceSubdomain != tokenSubdomain {
		return "", "", fmt.Errorf("subdomain mismatch")
	}

	// Constraint: Not allowed to use a user with TOTP with legacy token
	if deviceTotpSecret != "" {
		return "", "", fmt.Errorf("legacy token cannot be used for devices with TOTP enabled")
	}

	// Constraint: Not allowed to be admin and use legacy token
	if deviceRole == consts.UserRoleAdmin.String() {
		return "", "", fmt.Errorf("legacy token cannot be used for admin devices")
	}

	if !VerifyPassword(tokenPassword, devicePassword) {
		return "", "", fmt.Errorf("invalid password")
	}

	return deviceName, deviceRole, nil
}
