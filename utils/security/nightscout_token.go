package security

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

// NightscoutTokenSHA1 is a token's SHA1 in hex, which xDrip+ sends as its api-secret, as for a Nightscout API secret
func NightscoutTokenSHA1(token string) string {
	sum := sha1.Sum([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NightscoutTokenDigest is kept to find an identity from its token's SHA1: the SHA256 of the SHA1, so a copy of the
// database does not have what a client sends. The SHA1 is of a long random token, so it needs no salt or stretching.
func NightscoutTokenDigest(tokenSHA1 string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(tokenSHA1)))
	return hex.EncodeToString(sum[:])
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
