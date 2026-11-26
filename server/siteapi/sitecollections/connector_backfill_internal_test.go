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
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestBackfillWaitsAfterTheLastOne checks a site cannot start one backfill of a connector just after another. Each
// can hold a connector worker for 50 seconds and send the connector's host 200 requests, so a few sites starting them
// again and again against a slow host would stop every site's syncs, and use the server to load that host.
func TestBackfillWaitsAfterTheLastOne(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	_, siteModel, err := controlDB.GetSite(t.Context(), siteID)
	require.NoError(t, err)
	backfill := func(typ consts.ConnectorType) *httptest.ResponseRecorder {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, siteModel, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "admin", Role: consts.IdentityRoleAdmin})
		route := chi.NewRouteContext()
		route.URLParams.Add("id", string(typ))
		w := httptest.NewRecorder()
		BackfillHandler(w, httptest.NewRequest(http.MethodPost, "/settings/connectors/"+string(typ)+"/backfill", nil).WithContext(context.WithValue(ctx, chi.RouteCtxKey, route)))
		return w
	}
	// finish records the connector's running backfill as finished now, having taken took, as its job does
	finish := func(typ consts.ConnectorType, took time.Duration) {
		durationMs, records := took.Milliseconds(), int64(0)
		require.NoError(t, q.FinishConnectorBackfill(t.Context(), controldb.FinishConnectorBackfillParams{DurationMs: &durationMs, Records: &records, Type: typ, SiteID: siteID}))
	}
	for _, typ := range []consts.ConnectorType{consts.ConnectorTypeNightscout, consts.ConnectorTypeVirtual} {
		require.NoError(t, q.CreateConnector(t.Context(), controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))
	}

	// The first starts at once, and while it runs another is refused as before
	require.Equal(t, http.StatusAccepted, backfill(consts.ConnectorTypeNightscout).Code)
	require.Equal(t, http.StatusConflict, backfill(consts.ConnectorTypeNightscout).Code)

	// It took all the time a backfill has, as from a slow host: the next waits 5 times that, and says for how long
	finish(consts.ConnectorTypeNightscout, 50*time.Second)
	refused := backfill(consts.ConnectorTypeNightscout)
	require.Equal(t, http.StatusTooManyRequests, refused.Code, refused.Body.String())
	require.Contains(t, refused.Body.String(), "wait 250 seconds")
	source, err := q.GetConnector(t.Context(), controldb.GetConnectorParams{SiteID: siteID, Type: consts.ConnectorTypeNightscout})
	require.NoError(t, err)
	require.Nil(t, source.BackfillStartedAt, "a refused backfill is not started")
	require.Equal(t, int64(1), source.TotalBackfills)

	// Deleting the connector and adding it again keeps the wait
	nightscout := controldb.DeleteConnectorParams{Type: consts.ConnectorTypeNightscout, SiteID: siteID}
	require.NoError(t, q.DeleteConnector(t.Context(), nightscout))
	require.NoError(t, q.UpdateConnector(t.Context(), controldb.UpdateConnectorParams{Enabled: true, Type: nightscout.Type, SiteID: siteID}))
	require.Equal(t, http.StatusTooManyRequests, backfill(consts.ConnectorTypeNightscout).Code)

	// The virtual connector asks no host, so it has no wait
	require.Equal(t, http.StatusAccepted, backfill(consts.ConnectorTypeVirtual).Code)
	finish(consts.ConnectorTypeVirtual, 50*time.Second)
	require.Equal(t, http.StatusAccepted, backfill(consts.ConnectorTypeVirtual).Code)
}
