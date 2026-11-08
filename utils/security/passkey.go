package security

import (
	"net/http"
	"time"
)

// PASSKEY_SESSION_COOKIE keeps a passkey ceremony's challenge, encrypted, between its start and finish
const PASSKEY_SESSION_COOKIE = "passkey_session"

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
