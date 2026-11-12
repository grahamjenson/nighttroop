package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site using up the global rate limit cannot stop a control admin signing in to fix it, or one signed in using the
// control site, while the rest of the control site has a limit of its own
func TestGlobalRateLimitLeavesTheControlSite(t *testing.T) {
	testConfig, controlDB, site := testutils.NewControlDBWithSite(t)
	config := *testConfig
	limiters := *config.RateLimiters
	limiters.GlobalRateLimiter = rate.NewLimiter(0, 0)
	limiters.ControlRateLimiter = rate.NewLimiter(0, 1)
	config.RateLimiters = &limiters
	handler := GetMainHandler(controlDB, &config)
	token := testutils.ControlJWT(t, &config, controlDB, consts.FIRST_ADMIN_NAME)

	send := func(url string, token string) int {
		r := httptest.NewRequest("GET", url, nil)
		r.RemoteAddr = "127.0.0.1:1234"
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}

	// The global limit is used up, so the sites and their Nightscout API are refused
	require.Equal(t, http.StatusTooManyRequests, send("http://"+site.SiteDB.Subdomain+".localhost/login", ""))
	require.Equal(t, http.StatusTooManyRequests, send("http://"+site.SiteDB.Subdomain+".ns.localhost/api/v1/status.json", ""))

	// The control login is not, until the control site's own limit is used up
	require.Equal(t, http.StatusOK, send("http://control.localhost/login", ""))
	require.Equal(t, http.StatusTooManyRequests, send("http://control.localhost/login", ""))

	// A control admin signed in is never limited by either
	for range 3 {
		require.Equal(t, http.StatusOK, send("http://control.localhost/usage", token))
	}
	// But a token that is not a control admin's is limited as any other request
	require.Equal(t, http.StatusTooManyRequests, send("http://control.localhost/usage", strings.ToUpper(token)))
}
