package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestFailedFirstSyncDisablesTheSource(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, context.TODO(), site.SiteDB.Subdomain, consts.SyncSourceTypeDexcom
	jobs := sync.SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	get := func() *controldb.SiteSyncSource {
		s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s
	}
	runSync := func() error { return jobs.SyncSource(ctx, sync.SyncJobParams{SiteID: siteID, Type: typ}) }
	// Only sites seen in the last day are synced
	now := time.Now().UnixMilli()
	require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID}))
	due := func() bool {
		sources, err := q.GetSyncReportsToRun(ctx)
		require.NoError(t, err)
		for _, s := range sources {
			if s.SiteID == siteID && s.Type == typ {
				return true
			}
		}
		return false
	}

	// A new source with bad details: the first sync fails, turns syncing off and says why
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc("bad-user"), PasswordEnc: enc("pass"), Enabled: true}))
	require.True(t, due(), "a new source syncs straight away")
	require.Error(t, runSync())
	s := get()
	require.Nil(t, s.FirstSuccessfulSyncTime)
	require.False(t, s.Enabled)
	require.Contains(t, s.PreviousSyncErrors[0], "The first sync failed, so syncing has been turned off")
	require.False(t, due())

	// The failure is in the logs DB's errors, and turning syncing off is an event, both for the site
	syncErrors, err := controlDB.Logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, syncErrors, 1)
	require.Equal(t, `{"kind":"sync","sync_type":"dexcom"}`, *syncErrors[0].Attributes)
	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sync_turned_off", events[0].Name)

	// Saving working details clears the error, and the next sync is a first sync again
	require.NoError(t, q.UpdateSyncSource(ctx, controldb.UpdateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(dexcom_client.TestValidUsername), Enabled: true}))
	s = get()

	// It has failed before, so the retry waits out the doubled backoff rather than running at once
	require.Equal(t, int64(120), s.SyncIntervalSeconds)
	require.False(t, due())
	require.Nil(t, s.FirstSuccessfulSyncTime)
	require.Empty(t, s.PreviousSyncErrors)
	require.NoError(t, runSync())
	s = get()
	require.NotNil(t, s.FirstSuccessfulSyncTime)
	require.True(t, s.Enabled)

	// After the first sync has worked, a failure only backs off, it does not turn syncing off
	require.NoError(t, q.UpdateSyncSourceReportsFailure(ctx, controldb.UpdateSyncSourceReportsFailureParams{SiteID: siteID, Type: typ, ErrorMessages: errors.ErrorList{"remote is down"}}))
	s = get()
	require.True(t, s.Enabled)
	require.Equal(t, int64(120), s.SyncIntervalSeconds)

	// Deleting clears it, so a deleted source does not keep its first successful sync time
	require.NoError(t, q.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{SiteID: siteID, Type: typ}))
	require.Nil(t, get().FirstSuccessfulSyncTime)
}
