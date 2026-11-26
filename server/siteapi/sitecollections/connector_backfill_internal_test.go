package sitecollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestBackfillsShareTheSitesLimit checks a site's backfills, of all its connectors, share one limit of 1 a minute with
// a burst of 2. Each can hold a connector worker for 50 seconds and send the connector's host 200 requests, so a few
// sites starting them again and again against a slow host would stop every site's syncs, and use the server to load
// that host.
func TestBackfillsShareTheSitesLimit(t *testing.T) {
	testConfig, controlDB, site := testutils.NewControlDBWithSite(t)
	// The limit as a server has it, not the tests' far higher one
	cfg := *testConfig
	cfg.SiteBackfillPerMinuteLimit, cfg.SiteBackfillPerMinuteBurst = 1, 2
	cfg.RateLimiters = config.NewRateLimiters(cfg)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	_, siteModel, err := controlDB.GetSite(t.Context(), siteID)
	require.NoError(t, err)
	backfill := func(typ consts.ConnectorType) *httptest.ResponseRecorder {
		ctx := httputils.CreateSiteContext(context.Background(), &cfg, controlDB, site.SiteDB, siteModel, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "admin", Role: consts.IdentityRoleAdmin})
		route := chi.NewRouteContext()
		route.URLParams.Add("id", string(typ))
		w := httptest.NewRecorder()
		BackfillHandler(w, httptest.NewRequest(http.MethodPost, "/settings/connectors/"+string(typ)+"/backfill", nil).WithContext(context.WithValue(ctx, chi.RouteCtxKey, route)))
		return w
	}
	// finish records the connector's running backfill as finished, as its job does
	finish := func(typ consts.ConnectorType) {
		durationMs, records := time.Second.Milliseconds(), int64(0)
		require.NoError(t, q.FinishConnectorBackfill(t.Context(), controldb.FinishConnectorBackfillParams{DurationMs: &durationMs, Records: &records, Type: typ, SiteID: siteID}))
	}
	for _, typ := range []consts.ConnectorType{consts.ConnectorTypeNightscout, consts.ConnectorTypeVirtual} {
		require.NoError(t, q.CreateConnector(t.Context(), controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))
	}

	// The first starts at once. Another of it while it runs is refused as before, and does not spend the limit.
	require.Equal(t, http.StatusAccepted, backfill(consts.ConnectorTypeNightscout).Code)
	require.Equal(t, http.StatusConflict, backfill(consts.ConnectorTypeNightscout).Code)
	finish(consts.ConnectorTypeNightscout)

	// The burst's second is another connector's: the site's connectors share the limit
	require.Equal(t, http.StatusAccepted, backfill(consts.ConnectorTypeVirtual).Code)
	finish(consts.ConnectorTypeVirtual)

	// A third in the minute, of either connector, is refused with when to try again, and starts nothing
	for _, typ := range []consts.ConnectorType{consts.ConnectorTypeNightscout, consts.ConnectorTypeVirtual} {
		refused := backfill(typ)
		require.Equal(t, http.StatusTooManyRequests, refused.Code, refused.Body.String())
		source, err := q.GetConnector(t.Context(), controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		require.Nil(t, source.BackfillStartedAt, "a refused backfill is not started")
		require.Equal(t, int64(1), source.TotalBackfills)
	}
	require.Greater(t, cfg.RateLimiters.SiteBackfillLimiter(siteID).Tokens(), -0.1, "a refused backfill does not spend the limit")

	// Another site has its own limit
	require.True(t, cfg.RateLimiters.SiteBackfillLimiter("othersit").Allow())
}
