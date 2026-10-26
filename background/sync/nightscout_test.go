package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestEntriesChanged(t *testing.T) {
	read := int64(1000)
	tests := []struct {
		name     string
		lastRead *int64
		modified int64
		changed  bool
	}{
		{"never read", nil, 1000, true},
		{"unchanged", &read, 1000, false},
		{"modified since", &read, 2000, true},
		{"modified earlier, as after a restore", &read, 500, true},
		{"Nightscout does not say", &read, 0, true},
	}
	for _, test := range tests {
		require.Equal(t, test.changed, entriesChanged(test.lastRead, test.modified), test.name)
	}
}

func TestNightscoutSyncRemembersWhenEntriesWereModified(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	q, ctx, siteID, typ := controlDB.Queries, context.TODO(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	lastModified := func() *int64 {
		s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s.EntriesLastModified
	}
	runSync := func() error { return jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}) }

	// A sync reads the entries and remembers when Nightscout says they were last modified
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	require.Nil(t, lastModified())
	require.NoError(t, runSync())
	require.Equal(t, nightscout_client.TestEntriesLastModified, *lastModified())

	// With the entries unchanged, the next sync still works and keeps the same time
	require.NoError(t, runSync())
	require.Equal(t, nightscout_client.TestEntriesLastModified, *lastModified())

	// New settings may be another Nightscout, so its entries are read again
	require.NoError(t, q.UpdateSyncSource(ctx, controldb.UpdateSyncSourceParams{SiteID: siteID, Type: typ, Enabled: true}))
	require.Nil(t, lastModified())
	require.NoError(t, runSync())
	require.NotNil(t, lastModified())

	// Deleting the source forgets it too
	require.NoError(t, q.DeleteSyncSource(ctx, controldb.DeleteSyncSourceParams{SiteID: siteID, Type: typ}))
	require.Nil(t, lastModified())
}

func TestNightscoutClientIsKeptBetweenSyncs(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	q, ctx, siteID, typ := controlDB.Queries, context.TODO(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	runSync := func() *nightscoutClient {
		require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
		return jobs.nightscoutClients[siteID]
	}

	// The second sync uses the first's client, and so its JWT
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	first := runSync()
	require.NotNil(t, first)
	require.Same(t, first, runSync())

	// A new token gets a new client, which replaces the old one
	require.NoError(t, q.UpdateSyncSource(ctx, controldb.UpdateSyncSourceParams{SiteID: siteID, Type: typ, PasswordEnc: enc("new-token"), Enabled: true}))
	second := runSync()
	require.NotSame(t, first, second)
	require.Equal(t, "new-token", second.token)
	require.Len(t, jobs.nightscoutClients, 1)
}
