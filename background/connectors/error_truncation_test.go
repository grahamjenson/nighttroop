package connectors

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// A 10KB error (an upstream body, say) must not break the row's CHECK: it is stored truncated
// and the failure still counts, so the source backs off instead of being retried every tick
func TestOversizedSyncErrorIsTruncatedAndStillBacksOff(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	jobs := SetupConnectorJobManager(cfg, controlDB)
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))

	long := "HTTP 502: " + strings.Repeat("<html>", 2000)
	require.Greater(t, len(long), 10*1024)
	jobs.recordSyncFailure(&ConnectorJobInfo{ctx: ctx, SiteID: siteID, Type: typ, firstSync: true}, fmt.Errorf("%s", long), 10)

	s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Len(t, s.PreviousSyncErrors, 1)
	stored := s.PreviousSyncErrors[0]
	require.Contains(t, stored, "The first sync failed, so syncing has been turned off")
	require.Contains(t, stored, "HTTP 502: <html>")
	require.True(t, strings.HasSuffix(stored, "…"), "the cut is marked")
	require.Less(t, utf8.RuneCountInString(stored), errors.MaxErrorLength+120)

	// The failure was recorded, so the source is disabled (first sync) and backed off
	require.NotNil(t, s.ConsecutiveFailedSyncs)
	require.Equal(t, int64(1), *s.ConsecutiveFailedSyncs)
	require.Equal(t, int64(120), s.SyncIntervalSeconds)
	require.NotNil(t, s.LastAttemptedSync)
	require.False(t, s.Enabled)
}
