package sync_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestFailedFirstSyncDisablesTheSource(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
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

	// A new source with bad details: the first sync fails, turns syncing off and says why
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc("bad-user"), PasswordEnc: enc("pass"), Enabled: true}))
	require.Error(t, runSync())
	s := get()
	require.NotNil(t, s.FirstSyncTime)
	require.False(t, s.Enabled)
	require.Contains(t, s.PreviousSyncErrors[0], "The first sync failed, so syncing has been turned off")

	// Saving working details clears the error, and the next sync is a first sync again
	require.NoError(t, q.UpdateSyncSource(ctx, controldb.UpdateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(dexcom_client.TestValidUsername), Enabled: true}))
	s = get()
	require.Nil(t, s.FirstSyncTime)
	require.Empty(t, s.PreviousSyncErrors)
	require.NoError(t, runSync())
	s = get()
	require.NotNil(t, s.FirstSyncTime)
	require.True(t, s.Enabled)

	// After the first sync has worked, a failure only backs off, it does not turn syncing off
	_, err = controlDB.SQLiteDB.ExecContext(ctx, "UPDATE site_sync_sources SET url_or_username_enc = ? WHERE site_id = ? AND type = ?", *enc("bad-user"), siteID, typ)
	require.NoError(t, err)
	require.Error(t, runSync())
	s = get()
	require.True(t, s.Enabled)
	require.NotContains(t, s.PreviousSyncErrors[0], "first sync")

	// Deleting clears it, so a deleted source does not keep its first sync time
	require.NoError(t, q.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{SiteID: siteID, Type: typ}))
	require.Nil(t, get().FirstSyncTime)
}
