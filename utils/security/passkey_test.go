package security_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// TestWelcomePasskeyCookie checks a new site's welcome lets its browser add the admin a passkey for that site only, and
// only while it can: a cookie expired, changed or missing is refused
func TestWelcomePasskeyCookie(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	encrypt := func(value string) string {
		encrypted, err := cfg.EncryptFor(security.PurposeWelcomePasskey, "", value)
		require.NoError(t, err)
		return encrypted
	}
	set := httptest.NewRecorder()
	require.NoError(t, cfg.SetWelcomePasskeyCookie(set, "abcdefgh"))
	expired := "abcdefgh " + strconv.FormatInt(time.Now().Add(-time.Minute).UnixMilli(), 10)

	tests := []struct {
		name   string
		cookie *http.Cookie
		site   string
	}{
		{"set", set.Result().Cookies()[0], "abcdefgh"},
		{"expired", &http.Cookie{Name: security.WELCOME_PASSKEY_COOKIE, Value: encrypt(expired)}, ""},
		{"changed", &http.Cookie{Name: security.WELCOME_PASSKEY_COOKIE, Value: "abcdefgh 99999999999999"}, ""},
		{"missing", nil, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/welcome/passkey/begin", nil)
			if test.cookie != nil {
				r.AddCookie(test.cookie)
			}
			site, err := cfg.WelcomePasskeySite(r)
			require.Equal(t, test.site, site)
			require.Equal(t, test.site == "", err != nil)
		})
	}
}

// TestPasskeySessionFinishedOnce checks a passkey ceremony's session is finished once only: sending its cookie again,
// as a replayed sign in or registration would, is refused, while another ceremony's session still finishes
func TestPasskeySessionFinishedOnce(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	begin := func(session string) *http.Cookie {
		return beginPasskey(cfg, "192.0.2.1", session)
	}
	finish := func(cookie *http.Cookie) (string, error) {
		return finishPasskey(cfg, "192.0.2.1", cookie)
	}
	first, second := begin(`{"challenge":"one"}`), begin(`{"challenge":"two"}`)

	session, err := finish(first)
	require.NoError(t, err)
	require.Equal(t, `{"challenge":"one"}`, session)
	_, err = finish(first)
	require.Error(t, err)

	session, err = finish(second)
	require.NoError(t, err)
	require.Equal(t, `{"challenge":"two"}`, session)
	_, err = finish(&http.Cookie{Name: security.PASSKEY_SESSION_COOKIE, Value: "made up"})
	require.Error(t, err)
}

// beginPasskey begins a passkey ceremony from ip, with its session, giving the browser's cookie
func beginPasskey(cfg *security.SecurityConfig, ip string, session string) *http.Cookie {
	w := httptest.NewRecorder()
	cfg.SetPasskeySessionCookie(w, ip, []byte(session))
	return w.Result().Cookies()[0]
}

// finishPasskey is the session of the ceremony the cookie is for, finished from ip
func finishPasskey(cfg *security.SecurityConfig, ip string, cookie *http.Cookie) (string, error) {
	r := httptest.NewRequest(http.MethodPost, "/login/passkey/finish", nil)
	r.AddCookie(cookie)
	session, err := cfg.PasskeySession(httptest.NewRecorder(), r, ip)
	return string(session), err
}

// TestPasskeySessionsKeptPerIP checks many passkey ceremonies begun from one IP, as a flood would, push out only that
// IP's oldest: another IP's ceremony still finishes, and only from the IP that began it
func TestPasskeySessionsKeptPerIP(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	person, flooder := "192.0.2.1", "198.51.100.1"
	personal := beginPasskey(cfg, person, `{"challenge":"person"}`)
	first := beginPasskey(cfg, flooder, `{"challenge":"first"}`)
	for range 10_000 {
		beginPasskey(cfg, flooder, `{"challenge":"flood"}`)
	}
	last := beginPasskey(cfg, flooder, `{"challenge":"last"}`)

	_, err := finishPasskey(cfg, flooder, first)
	require.Error(t, err)
	_, err = finishPasskey(cfg, flooder, personal)
	require.Error(t, err)

	session, err := finishPasskey(cfg, person, personal)
	require.NoError(t, err)
	require.Equal(t, `{"challenge":"person"}`, session)
	session, err = finishPasskey(cfg, flooder, last)
	require.NoError(t, err)
	require.Equal(t, `{"challenge":"last"}`, session)
}
