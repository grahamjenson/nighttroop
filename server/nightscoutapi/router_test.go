package nightscoutapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"nighttroop.t1d.nz/utils/config"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
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
	require.Equal(t, float64(cfg.SiteGETPerMinuteBurst*config.SiteIdentitiesShare), cfg.RateLimiters.SiteGETLimiter(subdomain).Tokens())
}

// A device with a valid token is limited by its site and identity, not the budget an IP has before it authenticates,
// so a phone uploading readings is not refused after the few requests a bad token gets
func TestAuthenticatedRequestsDoNotSpendTheUnauthenticatedLimit(t *testing.T) {
	cfg := testConfig
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	subdomain := site.SiteDB.Subdomain
	identity, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: subdomain, Name: consts.FIRST_READ_ONLY_NAME})
	require.NoError(t, err)
	jwt, _, err := cfg.Security.GenerateUserJWTToken(subdomain, identity.Name, string(identity.Role))
	require.NoError(t, err)
	router := NewNightscoutRouter(controlDB, cfg)

	get := func(token string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/authorization/request/token", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, nil, nil, subdomain, "10.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req.WithContext(ctx))
		return rec.Code
	}

	// The site's GET burst of 3, more than the IP's unauthenticated burst of 2
	for range 3 {
		require.Equal(t, http.StatusOK, get(jwt))
	}
	require.Equal(t, float64(cfg.SiteUnauthPerMinuteBurst), cfg.RateLimiters.SiteUnauthLimiter(subdomain, "10.0.0.1").Tokens())

	// A bad token from the same IP still spends it, and once spent is refused before its token is checked
	require.Equal(t, http.StatusUnauthorized, get("not-a-jwt"))
	require.Equal(t, http.StatusUnauthorized, get("not-a-jwt"))
	require.Equal(t, http.StatusTooManyRequests, get("not-a-jwt"))
}
