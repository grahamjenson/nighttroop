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

// APISecretDigest is what an identity keeps to be found by its token's SHA-1 in hex, which is what Loop sends as its
// api-secret: the SHA-256 of that SHA-1, so the digest is not itself a credential
func APISecretDigest(tokenSHA1 string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(tokenSHA1)))
	return hex.EncodeToString(sum[:])
}

// NightscoutTokenDigest is the APISecretDigest of a token
func (cfg *SecurityConfig) NightscoutTokenDigest(token string) string {
	sum := sha1.Sum([]byte(token))
	return APISecretDigest(hex.EncodeToString(sum[:]))
}

// IsSHA1Hex is whether s is a SHA-1 in hex, which a token is not
func IsSHA1Hex(s string) bool {
	_, err := hex.DecodeString(s)
	return len(s) == 2*sha1.Size && err == nil
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
