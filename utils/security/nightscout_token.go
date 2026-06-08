package security

import (
	"encoding/base64"
	"fmt"
	"strings"
)

///////
// Legacy Token
///////

func (cfg *SecurityConfig) CreateNightscoutTokenForIdentity(subdomain, identityName, role, password string) (string, error) {
	tokenStr := fmt.Sprintf("%s:%s:%s:%s", subdomain, identityName, role, password)
	encryptedToken, err := cfg.Encrypt(tokenStr)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}
	urlEncodedToken := base64.RawURLEncoding.EncodeToString([]byte(encryptedToken))
	return urlEncodedToken, nil
}

func (cfg *SecurityConfig) SplitNightscoutTokenForIdentity(tokenStr string) (string, string, string, string, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(tokenStr)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decode token: %w", err)
	}

	decryptedToken, err := cfg.Decrypt(string(decodedBytes))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	parts := strings.SplitN(decryptedToken, ":", 4)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid token format")
	}

	tokenSubdomain := parts[0]
	tokenName := parts[1]
	tokenRole := parts[2]
	tokenPassword := parts[3]
	return tokenSubdomain, tokenName, tokenRole, tokenPassword, nil
}
