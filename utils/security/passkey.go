package security

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

// PASSKEY_SESSION_COOKIE names a passkey ceremony's session, kept on the server, between its start and finish
const PASSKEY_SESSION_COOKIE = "passkey_session"

// passkeySessionFor is how long a passkey ceremony has to finish
const passkeySessionFor = 5 * time.Minute

// newPasskeySessions keeps passkey ceremonies' sessions, with their challenges, in memory only, by a random ID only the
// browser that began each has. At most 10,000 are kept, many more than are begun in passkeySessionFor; one dropped
// early is refused, never finished twice.
func newPasskeySessions() *expirable.LRU[string, []byte] {
	return expirable.NewLRU[string, []byte](10_000, nil, passkeySessionFor)
}

// SetPasskeySessionCookie keeps a passkey ceremony's session, for the passkeySessionFor it has to finish
func (cfg *SecurityConfig) SetPasskeySessionCookie(w http.ResponseWriter, session []byte) {
	id := SecureRandomString(32, API_KEY_CHAR_SET)
	cfg.passkeySessions.Add(id, session)
	cfg.setCookie(w, PASSKEY_SESSION_COOKIE, id, time.Now().Add(passkeySessionFor))
}

// PasskeySession is the session SetPasskeySessionCookie kept, which is then removed, so each challenge is finished once
// only, whether or not the finish succeeds: sending the same cookie again, as a replay would, is refused
func (cfg *SecurityConfig) PasskeySession(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	id, err := cookieValue(r, PASSKEY_SESSION_COOKIE)
	if err != nil {
		return nil, err
	}
	cfg.clearCookie(w, PASSKEY_SESSION_COOKIE)
	session, ok := cfg.passkeySessions.Peek(id)
	if !ok || !cfg.passkeySessions.Remove(id) {
		return nil, fmt.Errorf("the passkey session has expired or was already finished")
	}
	return session, nil
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
