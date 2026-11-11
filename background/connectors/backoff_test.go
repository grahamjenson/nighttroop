package connectors_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestSyncBacksOffOnFailureAndResetsOnSuccess(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))
	interval := func() int64 {
		s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s.SyncIntervalSeconds
	}
	ms := int64(10)

	for _, want := range []int64{120, 240, 480} {
		require.NoError(t, q.UpdateConnectorReportsFailure(ctx, controldb.UpdateConnectorReportsFailureParams{ErrorMessages: []string{"HTTP 500"}, ElapsedMs: &ms, SiteID: siteID, Type: typ}))
		require.Equal(t, want, interval())
	}

	for range 10 {
		require.NoError(t, q.UpdateConnectorReportsFailure(ctx, controldb.UpdateConnectorReportsFailureParams{ErrorMessages: []string{"HTTP 400"}, ElapsedMs: &ms, SiteID: siteID, Type: typ}))
	}
	require.Equal(t, int64(600), interval())

	require.NoError(t, q.UpdateConnectorReportsSuccess(ctx, controldb.UpdateConnectorReportsSuccessParams{ElapsedMs: &ms, SiteID: siteID, Type: typ}))
	require.Equal(t, int64(60), interval())
}
