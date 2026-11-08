package security

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// PasskeyAudience is a link from creating a site to adding the admin's first passkey on the new site
var PasskeyAudience = "passkey"

// passkeyFor is how long a link to add a passkey is good for, and PASSKEY_SESSION_COOKIE keeps a passkey ceremony's
// challenge, encrypted, between its start and finish
const passkeyFor = 15 * time.Minute
const PASSKEY_SESSION_COOKIE = "passkey_session"

// GeneratePasskeyJWTToken is a link token for adding a passkey to the site's identity, good for 15 minutes
func (config *SecurityConfig) GeneratePasskeyJWTToken(subdomain string, identityName string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(passkeyFor)
	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{Subject: &identityName, Issuer: &subdomain, Audience: &PasskeyAudience,
		ExpiresAt: &expiresAt, IssuedAt: &now})
	if err != nil {
		return "", err
	}
	return config.jwtMac.ComputeMACAndEncode(rawJWT)
}

// VerifyPasskeyJWTToken is the identity name a link token to add a passkey is for, refusing one for another site or
// expired
func (config *SecurityConfig) VerifyPasskeyJWTToken(tokenStr string, subdomain string) (string, error) {
	validator, err := jwt.NewValidator(&jwt.ValidatorOpts{ExpectedAudience: &PasskeyAudience, ExpectIssuedInThePast: true, IgnoreIssuer: true})
	if err != nil {
		return "", err
	}
	verified, err := config.jwtMac.VerifyMACAndDecode(tokenStr, validator)
	if err != nil {
		return "", err
	}
	if issuer, err := verified.Issuer(); err != nil || issuer != subdomain {
		return "", fmt.Errorf("the link is for a different site")
	}
	return verified.Subject()
}

// SetPasskeySessionCookie keeps a passkey ceremony's session, encrypted, for the 5 minutes it has to finish
func (cfg *SecurityConfig) SetPasskeySessionCookie(w http.ResponseWriter, session []byte) error {
	encrypted, err := cfg.Encrypt(string(session))
	if err != nil {
		return err
	}
	cfg.setCookie(w, PASSKEY_SESSION_COOKIE, encrypted, time.Now().Add(5*time.Minute))
	return nil
}

// PasskeySession is the session SetPasskeySessionCookie kept, which is then dropped, so each is finished once
func (cfg *SecurityConfig) PasskeySession(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	encrypted, err := cookieValue(r, PASSKEY_SESSION_COOKIE)
	if err != nil {
		return nil, err
	}
	cfg.clearCookie(w, PASSKEY_SESSION_COOKIE)
	session, err := cfg.Decrypt(encrypted)
	return []byte(session), err
}
