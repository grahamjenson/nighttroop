package security

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/utils/consts"
)

///////
// Legacy Token
///////

func (cfg *SecurityConfig) CreateNightscoutTokenForIdentity(subdomain, identityName, password string) (string, error) {
	tokenStr := fmt.Sprintf("%s:%s:%s", subdomain, identityName, password)
	encryptedToken, err := cfg.Encrypt(tokenStr)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}
	urlEncodedToken := base64.RawURLEncoding.EncodeToString([]byte(encryptedToken))
	return urlEncodedToken, nil
}

type IdentityGetter func(ctx context.Context, name string) (identityName string, password string, totpSecret string, role consts.IdentityRole, err error)

func (cfg *SecurityConfig) VerifyNightscoutTokenForIdentity(tokenStr string, identitySubdomain string, identityGetter IdentityGetter) (string, consts.IdentityRole, error) {

	decodedBytes, err := base64.RawURLEncoding.DecodeString(tokenStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode token: %w", err)
	}

	decryptedToken, err := cfg.Decrypt(string(decodedBytes))
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	parts := strings.SplitN(decryptedToken, ":", 3)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid token format")
	}

	tokenSubdomain := parts[0]
	tokenName := parts[1]
	tokenPassword := parts[2]

	identityName, identityPassword, identityTotpSecret, identityRole, err := identityGetter(context.TODO(), tokenName)

	if err != nil {
		return "", "", fmt.Errorf("identity not found: %w", err)
	}

	if identitySubdomain != tokenSubdomain {
		return "", "", fmt.Errorf("subdomain mismatch")
	}

	// Constraint: Not allowed to use a user with TOTP with legacy token
	if identityTotpSecret != "" {
		return "", "", fmt.Errorf("legacy token cannot be used for identities with TOTP enabled")
	}

	// Constraint: Not allowed to be admin and use legacy token
	if identityRole == consts.IdentityRoleAdmin {
		return "", "", fmt.Errorf("legacy token cannot be used for admin identities")
	}

	if !VerifyPassword(tokenPassword, identityPassword) {
		return "", "", fmt.Errorf("invalid password")
	}

	return identityName, identityRole, nil
}
