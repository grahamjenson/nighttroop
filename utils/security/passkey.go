package security

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

// PASSKEY_SESSION_COOKIE names a passkey ceremony's session, kept on the server, between its start and finish
const PASSKEY_SESSION_COOKIE = "passkey_session"

// passkeySessionFor is how long a passkey ceremony has to finish
const passkeySessionFor = 5 * time.Minute

// passkeySessionsPerIP is how many ceremonies one IP keeps begun and not finished, more than the people behind one
// address begin at once; past it, the IP's oldest is dropped, never another IP's
const passkeySessionsPerIP = 5

// passkeyCeremony is one ceremony's session, by a random ID only the browser that began it has
type passkeyCeremony struct {
	id      string
	session []byte
	until   time.Time
}

// passkeySessions keeps passkey ceremonies' sessions, with their challenges, in memory only, for the 10,000 IPs that
// began one last, so many begun from one IP push out only its own. One dropped early is refused, never finished twice.
type passkeySessions struct {
	mu   sync.Mutex
	byIP *expirable.LRU[string, []passkeyCeremony]
}

func newPasskeySessions() *passkeySessions {
	return &passkeySessions{byIP: expirable.NewLRU[string, []passkeyCeremony](10_000, nil, passkeySessionFor)}
}

// SetPasskeySessionCookie keeps a passkey ceremony's session begun from ip, for the passkeySessionFor it has to finish
func (cfg *SecurityConfig) SetPasskeySessionCookie(w http.ResponseWriter, ip string, session []byte) {
	s, id, until := cfg.passkeySessions, SecureRandomString(32, API_KEY_CHAR_SET), time.Now().Add(passkeySessionFor)
	s.mu.Lock()
	ceremonies, _ := s.byIP.Peek(ip)
	ceremonies = append(ceremonies, passkeyCeremony{id: id, session: session, until: until})
	s.byIP.Add(ip, ceremonies[max(0, len(ceremonies)-passkeySessionsPerIP):])
	s.mu.Unlock()
	cfg.setCookie(w, PASSKEY_SESSION_COOKIE, id, until)
}

// PasskeySession is the session SetPasskeySessionCookie kept for ip, which is then removed, so each challenge is
// finished once only, whether or not the finish succeeds: sending the same cookie again, as a replay would, is refused
func (cfg *SecurityConfig) PasskeySession(w http.ResponseWriter, r *http.Request, ip string) ([]byte, error) {
	id, err := cookieValue(r, PASSKEY_SESSION_COOKIE)
	if err != nil {
		return nil, err
	}
	cfg.clearCookie(w, PASSKEY_SESSION_COOKIE)
	s := cfg.passkeySessions
	s.mu.Lock()
	defer s.mu.Unlock()
	ceremonies, _ := s.byIP.Peek(ip)
	for i, ceremony := range ceremonies {
		if subtle.ConstantTimeCompare([]byte(ceremony.id), []byte(id)) == 1 {
			s.byIP.Add(ip, slices.Delete(slices.Clone(ceremonies), i, i+1))
			if time.Now().Before(ceremony.until) {
				return ceremony.session, nil
			}
		}
	}
	return nil, fmt.Errorf("the passkey session has expired or was already finished")
}

// WELCOME_PASSKEY_COOKIE lets the browser a new site's credentials were shown to add its admin a passkey
const WELCOME_PASSKEY_COOKIE = "welcome_passkey"

// welcomePasskeyFor is how long after its credentials are shown a new site's admin can add a passkey from its welcome
const welcomePasskeyFor = 15 * time.Minute

// SetWelcomePasskeyCookie lets this browser add the new site's admin a passkey, for welcomePasskeyFor. The site and
// when it ends are encrypted, so neither can be changed, and the cookie is only sent to the site's own address.
func (cfg *SecurityConfig) SetWelcomePasskeyCookie(w http.ResponseWriter, subdomain string) error {
	until := time.Now().Add(welcomePasskeyFor)
	encrypted, err := cfg.EncryptFor(PurposeWelcomePasskey, "", subdomain+" "+strconv.FormatInt(until.UnixMilli(), 10))
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
	value, err := cfg.DecryptFor(PurposeWelcomePasskey, "", encrypted)
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
