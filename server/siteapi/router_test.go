package siteapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// testConfig allows 2 unauthenticated requests an IP on a site; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_GET_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_GET_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_POST_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_POST_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_BURST", "2")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

// runs makes each test's site new, so -count does not share their limits
var runs atomic.Int64

// Anyone who knows a subdomain can load its login page, try to log in, or send a bad cookie. Those spend a budget for
// their IP on the site, not the site's own limits, which its signed in identities and Nightscout devices share.
func TestUnauthenticatedRequestsDoNotSpendTheSiteLimit(t *testing.T) {
	cfg := testConfig
	subdomain := fmt.Sprint("site", runs.Add(1))
	router := NewSiteRouter(nil, cfg)

	send := func(ip string, method string, path string) int {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Accept", "application/json")
		req.AddCookie(&http.Cookie{Name: security.SITE_TOKEN_COOKIE, Value: "not-a-jwt"})
		// The main router sets the site context from the subdomain and the IP before it reaches this router
		ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, subdomain, ip, httputils.GetHTTPAction(req), httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req.WithContext(ctx))
		return rec.Code
	}

	// One IP's unauthenticated burst of 2 then runs out, for the login page, a login and a bad sign in alike
	require.Equal(t, http.StatusOK, send("10.0.0.1", http.MethodGet, "/login"))
	require.Equal(t, http.StatusUnauthorized, send("10.0.0.1", http.MethodGet, "/summary"))
	require.Equal(t, http.StatusTooManyRequests, send("10.0.0.1", http.MethodPost, "/login"))
	require.Equal(t, http.StatusTooManyRequests, send("10.0.0.1", http.MethodGet, "/summary"))

	// Another IP has its own budget on the site
	require.Equal(t, http.StatusOK, send("10.0.0.2", http.MethodGet, "/login"))
	require.Equal(t, http.StatusUnauthorized, send("10.0.0.2", http.MethodPost, "/login"))

	// None of those spent the site's limits, which a signed in identity's request is checked against
	require.Equal(t, float64(cfg.SitePOSTPerMinuteBurst*config.SiteIdentitiesShare), cfg.RateLimiters.SitePOSTLimiter(subdomain).Tokens())
	require.Equal(t, float64(cfg.SiteGETPerMinuteBurst*config.SiteIdentitiesShare), cfg.RateLimiters.SiteGETLimiter(subdomain).Tokens())
}
