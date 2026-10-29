package sync_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestSyncBacksOffOnFailureAndResetsOnSuccess(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, context.TODO(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, Enabled: true}))
	interval := func() int64 {
		s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s.SyncIntervalSeconds
	}
	ms := int64(10)

	for _, want := range []int64{120, 240, 480} {
		require.NoError(t, q.UpdateSyncSourceReportsFailure(ctx, controldb.UpdateSyncSourceReportsFailureParams{ErrorMessages: []string{"HTTP 500"}, ElapsedMs: &ms, SiteID: siteID, Type: typ}))
		require.Equal(t, want, interval())
	}

	for range 10 {
		require.NoError(t, q.UpdateSyncSourceReportsFailure(ctx, controldb.UpdateSyncSourceReportsFailureParams{ErrorMessages: []string{"HTTP 400"}, ElapsedMs: &ms, SiteID: siteID, Type: typ}))
	}
	require.Equal(t, int64(600), interval())

	require.NoError(t, q.UpdateSyncSourceReportsSuccess(ctx, controldb.UpdateSyncSourceReportsSuccessParams{ElapsedMs: &ms, SiteID: siteID, Type: typ}))
	require.Equal(t, int64(60), interval())
}
