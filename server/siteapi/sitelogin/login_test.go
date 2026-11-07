package sitelogin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// testConfig allows one login attempt an IP and name; a config can only be made once per process
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
		preAuth := func() string {
			token, _, err := testConfig.Security.GeneratePreAuthJWTToken(siteID, consts.FIRST_ADMIN_NAME)
			require.NoError(t, err)
			return token
		}

		token := ""
		if preAuthBeforeReset {
			token = preAuth()
			time.Sleep(2 * time.Millisecond) // so the reset is a millisecond after the token
		}
		admin, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
		require.NoError(t, err)
		_, err = controlDB.ResetSiteIdentityPassword(t.Context(), admin, "control admin")
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
