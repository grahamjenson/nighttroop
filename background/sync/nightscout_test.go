package sync

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestChangedSince(t *testing.T) {
	tests := []struct {
		name     string
		lastRead int64
		modified int64
		changed  bool
	}{
		{"never read", 0, 1000, true},
		{"unchanged", 1000, 1000, false},
		{"modified since", 1000, 2000, true},
		{"modified earlier, as after a restore", 1000, 500, true},
		{"Nightscout does not say", 1000, 0, true},
	}
	for _, test := range tests {
		require.Equal(t, test.changed, changedSince(test.lastRead, test.modified), test.name)
	}
}

func TestNightscoutSyncRemembersWhenCollectionsWereModified(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	all := func(modified int64) map[nsutils.Collection]int64 {
		return map[nsutils.Collection]int64{nsutils.Entries: modified, nsutils.Treatments: modified, nsutils.Devicestatus: modified, nsutils.Profile: modified}
	}
	client := func() *nightscoutClient {
		c, _ := jobs.nightscoutClients.Get(siteID)
		return c
	}
	runSync := func() map[nsutils.Collection]int64 {
		require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
		return client().lastModified
	}

	// A sync reads entries, treatments, device statuses and profiles and its client remembers when Nightscout says each was last modified
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	require.Equal(t, all(nightscout_client.TestLastModified), runSync())

	// With them unchanged, the next sync still works and keeps the same times
	require.Equal(t, all(nightscout_client.TestLastModified), runSync())

	// New details may be another Nightscout, so their new client reads everything again
	require.NoError(t, q.UpdateSyncSource(ctx, controldb.UpdateSyncSourceParams{SiteID: siteID, Type: typ, PasswordEnc: enc("new-token"), Enabled: true}))
	first := client()
	require.Equal(t, all(nightscout_client.TestLastModified), runSync())
	require.NotSame(t, first, client())
}

func TestNightscoutClientIsKeptBetweenSyncs(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
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
	cfg, controlDB, _ := testutils.NewControlDBWithSite(t)
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
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
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
	readings, err := site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	require.Equal(t, 120.0, readings[0].Mgdl)
	doses, err := site.SiteDB.Queries().AllInsulinDoses(ctx, sitedb.AllInsulinDosesParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, doses, 1)
	require.Equal(t, 4.0, doses[0].Dose)
	food, err := site.SiteDB.Queries().AllFoodIntakes(ctx, sitedb.AllFoodIntakesParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, food, 1)
	reports, err := site.SiteDB.Queries().AllDeviceReports(ctx, sitedb.AllDeviceReportsParams{Limit: 10})
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

// A profile store is used long after it starts, so the newest is copied whenever the profiles change, however old, as
// Nightscout's profile editor saves it through v1, without a date. An edit is copied though it was saved by another
// subject on that Nightscout than the sync that made the site's copy.
func TestNightscoutSyncCopiesTheCurrentProfile(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	startDate := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	nightscout_client.TestDocuments = map[string][]any{"profile": {map[string]any{
		"_id": "6553f1b2c4a1e2d3f4a5b6c7", "defaultProfile": "Default", "startDate": startDate, "created_at": startDate,
		"mills": float64(time.Now().Add(-30 * 24 * time.Hour).UnixMilli()), "units": "mg/dl", "srvModified": float64(time.Now().UnixMilli()),
		// Who saved it on that Nightscout, which the site's copy does not keep, as the sync made it
		"subject": "admin",
		"store": map[string]any{"Default": map[string]any{
			"dia": 5, "units": "mg/dl", "timezone": "Pacific/Auckland",
			"basal":       []any{map[string]any{"time": "00:00", "value": 0.8, "timeAsSeconds": 0}},
			"carbratio":   []any{map[string]any{"time": "00:00", "value": 10, "timeAsSeconds": 0}},
			"sens":        []any{map[string]any{"time": "00:00", "value": 50, "timeAsSeconds": 0}},
			"target_low":  []any{map[string]any{"time": "00:00", "value": 90, "timeAsSeconds": 0}},
			"target_high": []any{map[string]any{"time": "00:00", "value": 120, "timeAsSeconds": 0}},
		}},
	}}}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })
	profileStores := func() []*sitedb.ProfilePreset {
		stores, err := site.SiteDB.Queries().GetMostRecentProfileStore(ctx)
		require.NoError(t, err)
		return stores
	}

	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
	stores := profileStores()
	require.Len(t, stores, 1)
	require.Equal(t, "Default", stores[0].Name)
	require.True(t, stores[0].IsDefault)

	// Read again, as when the Nightscout's lastModified moves for another profile change, it is not copied twice
	c, _ := jobs.nightscoutClients.Get(siteID)
	c.lastModified = map[nsutils.Collection]int64{}
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
	require.Len(t, profileStores(), 1)
	require.Equal(t, stores[0].ProfileStoreID, profileStores()[0].ProfileStoreID)

	// Nightscout's profile editor saves an edit to the same document, which replaces the copy of it
	doc := nightscout_client.TestDocuments["profile"][0].(map[string]any)
	doc["store"].(map[string]any)["Default"].(map[string]any)["basal"] = []any{map[string]any{"time": "00:00", "value": 0.9, "timeAsSeconds": 0}}
	c.lastModified = map[nsutils.Collection]int64{}
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ}))
	edited := profileStores()
	require.Len(t, edited, 1)
	require.Equal(t, stores[0].ProfileStoreID, edited[0].ProfileStoreID)
	require.JSONEq(t, `[{"time":"00:00","timeAsSeconds":0,"value":0.9}]`, *edited[0].BasalRatesJson)
}
