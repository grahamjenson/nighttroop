package sitelogin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// testConfig allows one login attempt an IP; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_LOGIN_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "1")
	testConfig = testutils.GetTestServerConfig()
	os.Exit(m.Run())
}

// A name longer than any identity's is refused before it reaches the rate limiter, which keys on it
func TestLoginRejectsLongNameBeforeRateLimit(t *testing.T) {
	longName := strings.Repeat("a", 33)
	tests := []struct {
		name        string
		ip          string
		contentType string
		accept      string
		body        string
	}{
		{"json", "203.0.113.1", "application/json", "application/json", `{"name":"` + longName + `","password":"pw"}`},
		{"form", "203.0.113.2", "application/x-www-form-urlencoded", "text/htmx", "name=" + longName + "&password=pw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("Accept", tt.accept)
			ctx := httputils.CreateSiteContext(httputils.WithClientIP(r), testConfig, nil, nil, nil, "site"+tt.name, tt.ip, httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, http.StatusUnauthorized, w.Code)

			// The IP still has its one attempt, so the refused login never reached the limiter
			allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(testConfig, r, "site"+tt.name, "phone")
			require.True(t, allowed)
		})
	}
}

// The TOTP step refuses a pre-auth token from before the identity's password was reset, as it proved the old password,
// and takes one from after it
func TestLoginTOTPRefusesPreAuthFromBeforeAReset(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	// Each from its own IP and site, as login attempts are limited to one for each IP and for each site and name
	for i, preAuthBeforeReset := range []bool{true, false} {
		ip := fmt.Sprintf("203.0.113.%d", 10+i)
		site, err := controlDB.CreateSite(t.Context())
		require.NoError(t, err)
		siteID := site.SiteDB.Subdomain
		getAdmin := func() *controldb.SiteIdentity {
			admin, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
			require.NoError(t, err)
			return admin
		}
		// As the password step makes it, for the identity as it is then
		preAuth := func() string {
			token, _, err := testConfig.Security.GeneratePreAuthJWTToken(getAdmin(), siteID, consts.FIRST_ADMIN_NAME)
			require.NoError(t, err)
			return token
		}

		token := ""
		if preAuthBeforeReset {
			token = preAuth()
		}
		_, err = controlDB.ResetSiteIdentityPassword(t.Context(), getAdmin(), "control admin")
		require.NoError(t, err)
		if !preAuthBeforeReset {
			token = preAuth()
		}

		r := httptest.NewRequest("POST", "/login/totp", strings.NewReader("totp=123123"))
		r.RemoteAddr = ip + ":1234"
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Accept", "text/htmx")
		r.AddCookie(&http.Cookie{Name: security.PRE_AUTH_TOKEN_COOKIE, Value: token})
		ctx := httputils.CreateSiteContext(httputils.WithClientIP(r), testConfig, controlDB, site.SiteDB, nil, siteID, ip, httputils.HTTPActionWrite, httputils.AcceptTypeHTMX, httputils.ContentTypeForm)
		w := httptest.NewRecorder()
		POSTLoginTOTP(controlDB, testConfig)(w, r.WithContext(ctx))

		signedIn := w.Header().Get("HX-Redirect") == "/"
		require.Equal(t, !preAuthBeforeReset, signedIn, "pre-auth before the reset: %v, response %d", preAuthBeforeReset, w.Code)
	}
}

// A login refused by the rate limiter says to wait a minute and try again, on the page and to the API
func TestLoginRateLimitedSaysToWait(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		contentType string
		accept      string
		body        string
	}{
		{"json", "198.51.100.1", "application/json", "application/json", `{"name":"phone","password":"pw"}`},
		{"form", "198.51.100.2", "application/x-www-form-urlencoded", "text/htmx", "name=phone&password=pw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("Accept", tt.accept)
			ctx := httputils.CreateSiteContext(httputils.WithClientIP(r), testConfig, nil, nil, nil, "limited"+tt.name, tt.ip, httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			r = r.WithContext(ctx)
			// The one attempt the test config allows is used up
			allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(testConfig, r, "limited"+tt.name, "phone")
			require.True(t, allowed)

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, http.StatusTooManyRequests, w.Code)
			require.Contains(t, w.Body.String(), "Too many login attempts. Wait a minute, then try again.")
		})
	}
}

// A name with no identity is refused as a wrong password is, with a password checked, so neither its answer nor how
// long it takes tells which names exist
func TestLoginMissingNameIsLikeWrongPassword(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	tests := []struct {
		name        string
		contentType string
		accept      string
		body        func(name string) string
	}{
		{"api", "application/json", "application/json", func(name string) string { return `{"name":"` + name + `","password":"wrong"}` }},
		{"page", "application/x-www-form-urlencoded", "text/htmx", func(name string) string { return "name=" + name + "&password=wrong" }},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site, err := controlDB.CreateSite(t.Context())
			require.NoError(t, err)
			siteID := site.SiteDB.Subdomain
			// Each from its own IP, as login attempts are limited to one for each IP
			login := func(ip string, name string) (int, string) {
				r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body(name)))
				r.RemoteAddr = ip + ":1234"
				r.Header.Set("Content-Type", tt.contentType)
				r.Header.Set("Accept", tt.accept)
				ctx := httputils.CreateSiteContext(httputils.WithClientIP(r), testConfig, controlDB, site.SiteDB, nil, siteID, ip, httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
				before := testConfig.Security.PasswordsVerified()
				w := httptest.NewRecorder()
				POSTLogin(controlDB, testConfig)(w, r.WithContext(ctx))
				require.Equal(t, int64(1), testConfig.Security.PasswordsVerified()-before, "passwords checked logging in as %s", name)
				// A toast has an ID of its own each time
				return w.Code, regexp.MustCompile(`id="id-[A-Z0-9]+"`).ReplaceAllString(w.Body.String(), "")
			}

			wrongStatus, wrongBody := login(fmt.Sprintf("192.0.2.%d", 10+2*i), consts.FIRST_ADMIN_NAME)
			missingStatus, missingBody := login(fmt.Sprintf("192.0.2.%d", 11+2*i), "nosuchname")
			require.Equal(t, http.StatusUnauthorized, wrongStatus)
			require.Equal(t, wrongStatus, missingStatus)
			require.Equal(t, wrongBody, missingBody)
		})
	}
}
