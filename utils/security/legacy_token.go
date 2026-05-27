package security

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/utils/consts"
)

///////
// Legacy Token
///////

func (cfg *SecurityConfig) CreateLegacyTokenForIdentity(subdomain, identityName, password string) (string, error) {
	tokenStr := fmt.Sprintf("%s:%s:%s", subdomain, identityName, password)
	encryptedToken, err := cfg.Encrypt(tokenStr)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}
	return encryptedToken, nil
}

type IdentityGetter func(ctx context.Context, name string) (identityName string, password string, totpSecret string, role consts.IdentityRole, err error)

func (cfg *SecurityConfig) VerifyLegacyTokenForIdentity(tokenStr string, identitySubdomain string, identityGetter IdentityGetter) (string, consts.IdentityRole, error) {
	decryptedToken, err := cfg.Decrypt(tokenStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	var tokenSubdomain, tokenName, tokenPassword string
	_, err = fmt.Sscanf(decryptedToken, "%s:%s:%s", &tokenSubdomain, &tokenName, &tokenPassword)
	if err != nil {
		return "", "", fmt.Errorf("invalid token format: %w", err)
	}

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
