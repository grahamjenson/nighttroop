package security

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// WELCOME_PASSKEY_COOKIE lets the browser a new site's credentials were shown to add its admin a passkey
const WELCOME_PASSKEY_COOKIE = "welcome_passkey"

// welcomePasskeyFor is how long after its credentials are shown a new site's admin can add a passkey from its welcome
const welcomePasskeyFor = 15 * time.Minute

// SetWelcomePasskeyCookie lets this browser add the new site's admin a passkey, for welcomePasskeyFor. The site and
// when it ends are encrypted, so neither can be changed, and the cookie is only sent to the site's own address.
func (cfg *SecurityConfig) SetWelcomePasskeyCookie(w http.ResponseWriter, subdomain string) error {
	until := time.Now().Add(welcomePasskeyFor)
	encrypted, err := cfg.Encrypt(subdomain + " " + strconv.FormatInt(until.UnixMilli(), 10))
	if err != nil {
		return err
	}
	cfg.setCookie(w, WELCOME_PASSKEY_COOKIE, encrypted, until)
	return nil
}

// WelcomePasskeySite is the site SetWelcomePasskeyCookie let this browser add the admin a passkey to, while it can
func (cfg *SecurityConfig) WelcomePasskeySite(r *http.Request) (string, error) {
	encrypted, err := cookieValue(r, WELCOME_PASSKEY_COOKIE)
	if err != nil {
		return "", err
	}
	value, err := cfg.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	subdomain, untilMs, _ := strings.Cut(value, " ")
	if until, err := strconv.ParseInt(untilMs, 10, 64); err != nil || time.Now().UnixMilli() > until {
		return "", fmt.Errorf("the welcome's passkey cookie has expired")
	}
	return subdomain, nil
}

// ClearWelcomePasskeyCookie drops the cookie once the admin's passkey is added, so it adds only one
func (cfg *SecurityConfig) ClearWelcomePasskeyCookie(w http.ResponseWriter) {
	cfg.clearCookie(w, WELCOME_PASSKEY_COOKIE)
}
