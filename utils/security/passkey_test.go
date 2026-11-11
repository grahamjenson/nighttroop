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
		encrypted, err := cfg.Encrypt(value)
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
		w := httptest.NewRecorder()
		cfg.SetPasskeySessionCookie(w, []byte(session))
		return w.Result().Cookies()[0]
	}
	finish := func(cookie *http.Cookie) (string, error) {
		r := httptest.NewRequest(http.MethodPost, "/login/passkey/finish", nil)
		r.AddCookie(cookie)
		session, err := cfg.PasskeySession(httptest.NewRecorder(), r)
		return string(session), err
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
