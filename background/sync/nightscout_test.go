package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestChangedSince(t *testing.T) {
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
		require.Equal(t, test.changed, changedSince(test.lastRead, test.modified), test.name)
	}
}

func TestNightscoutSyncRemembersWhenCollectionsWereModified(t *testing.T) {
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
		require.Equal(t, s.EntriesLastModified, s.TreatmentsLastModified, "the fake Nightscout modified each at the same time")
		require.Equal(t, s.EntriesLastModified, s.DevicestatusLastModified, "the fake Nightscout modified each at the same time")
		return s.EntriesLastModified
	}
	runSync := func() error { return jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}) }

	// A sync reads entries, treatments and device statuses and remembers when Nightscout says each was last modified
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	require.Nil(t, lastModified())
	require.NoError(t, runSync())
	require.Equal(t, nightscout_client.TestLastModified, *lastModified())

	// With them unchanged, the next sync still works and keeps the same times
	require.NoError(t, runSync())
	require.Equal(t, nightscout_client.TestLastModified, *lastModified())

	// New settings may be another Nightscout, so everything is read again
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
		client, _ := jobs.nightscoutClients.Get(siteID)
		return client
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
	require.Equal(t, 1, jobs.nightscoutClients.Len())
}

func TestNightscoutClientsAreBounded(t *testing.T) {
	cfg, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	jobs := SetupSyncJobManager(cfg, controlDB)
	clientFor := func(i int) *nightscoutClient {
		return jobs.nightscoutClientFor(&SyncJobInfo{SiteID: fmt.Sprintf("site-%d", i), urlOrUsername: nightscout_client.TestValidURL, password: "token"})
	}

	// Once there are too many, the least recently used site's client is dropped
	first := clientFor(0)
	for i := 1; i < maxNightscoutClients; i++ {
		clientFor(i)
	}
	require.Same(t, first, clientFor(0), "using it makes site-1 the least recently used")
	clientFor(maxNightscoutClients)
	require.Equal(t, maxNightscoutClients, jobs.nightscoutClients.Len())
	require.True(t, jobs.nightscoutClients.Contains("site-0"))
	require.False(t, jobs.nightscoutClients.Contains("site-1"))
}

func TestNightscoutSyncCopiesEntriesTreatmentsAndDeviceStatuses(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	q, ctx, siteID, typ := controlDB.Queries, context.TODO(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	now := float64(time.Now().UnixMilli())
	nightscout_client.TestDocuments = map[string][]any{
		"entries": {map[string]any{"type": "sgv", "sgv": 120, "date": now, "srvModified": now}},
		"treatments": {
			map[string]any{"eventType": "Meal Bolus", "carbs": 40, "insulin": 4, "date": now, "srvCreated": now},
			map[string]any{"eventType": "Something Custom", "date": now},
		},
		"devicestatus": {map[string]any{"device": "openaps://phone", "date": now, "uploaderBattery": 80}},
	}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })

	// The reading, the meal and the status are copied, and the treatment the site does not know is skipped, not failing the sync
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
	readings, err := site.SiteDB.Queries.AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	require.Equal(t, 120.0, readings[0].Mgdl)
	doses, err := site.SiteDB.Queries.AllInsulinDoses(ctx, sitedb.AllInsulinDosesParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, doses, 1)
	require.Equal(t, 4.0, doses[0].Dose)
	food, err := site.SiteDB.Queries.AllFoodIntakes(ctx, sitedb.AllFoodIntakesParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, food, 1)
	reports, err := site.SiteDB.Queries.AllDeviceReports(ctx, sitedb.AllDeviceReportsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, reports, 1)
	require.Equal(t, "openaps://phone", *reports[0].CreatedByDeviceName)

	// The skipped treatment is in the site's logs
	logs, err := controlDB.Logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "skipped Nightscout documents the site refused", logs[0].Message)
	require.Contains(t, *logs[0].Attributes, `"collection":"treatments"`)
}
