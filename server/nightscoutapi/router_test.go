package nightscoutapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/httputils"
)

// Anyone who knows a subdomain can send it requests that fail auth, so those spend a budget for their IP on the
// site, not the site's own GET limit that its real devices need once they authenticate
func TestUnauthenticatedRequestsDoNotSpendTheSiteLimit(t *testing.T) {
	cfg := testConfig
	run := runs.Add(1)
	subdomain := fmt.Sprint("site", run)
	router := NewNightscoutRouter(nil, cfg)

	get := func(ip string, path string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		// The main router sets the site context from the subdomain and the IP before it reaches this router
		ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, subdomain, ip, httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req.WithContext(ctx))
		return rec.Code
	}

	// One IP's unauthenticated burst of 2 then runs out, including for the Socket.IO handshake
	require.Equal(t, http.StatusUnauthorized, get("10.0.0.1", "/api/v1/entries.json"))
	require.Equal(t, http.StatusUnauthorized, get("10.0.0.1", "/api/v1/entries.json"))
	require.Equal(t, http.StatusTooManyRequests, get("10.0.0.1", "/api/v1/entries.json"))
	require.Equal(t, http.StatusTooManyRequests, get("10.0.0.1", "/socket.io/?EIO=4&transport=polling"))

	// Another IP has its own budget on the site
	require.Equal(t, http.StatusUnauthorized, get("10.0.0.2", "/api/v1/entries.json"))

	// None of those spent the site's GET limit, which an authenticated device's request would be checked against
	require.Equal(t, float64(cfg.SiteGETPerMinuteBurst), cfg.RateLimiters.SiteGETLimiter(subdomain).Tokens())
}
