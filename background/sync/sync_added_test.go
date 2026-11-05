package sync_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync that adds records records a sync_added event with how many, keeps them as its source's last sync's records,
// and adds them to its total; one that adds none records no event
func TestSyncThatAddsRecordsRecordsAnEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID := controlDB.Queries, t.Context(), site.SiteDB.Subdomain
	jobs := sync.SetupSyncJobManager(cfg, controlDB)
	now := time.Now().UnixMilli()
	require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID}))
	username, err := cfg.Security.Encrypt(dexcom_client.TestValidUsername)
	require.NoError(t, err)
	password, err := cfg.Security.Encrypt("pass")
	require.NoError(t, err)
	added := func() []string {
		events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Search: new("sync_added"), Limit: 10})
		require.NoError(t, err)
		attributes := []string{}
		for _, event := range events {
			require.Equal(t, "sync_added", event.Name)
			attributes = append(attributes, *event.Attributes)
		}
		return attributes
	}

	// The test Dexcom has no readings, so its sync adds nothing
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: consts.SyncSourceTypeDexcom, UrlOrUsernameEnc: &username, PasswordEnc: &password, Enabled: true}))
	require.NoError(t, jobs.SyncSource(ctx, sync.SyncJobParams{SiteID: siteID, Type: consts.SyncSourceTypeDexcom}))
	require.Empty(t, added())
	source := func(typ consts.SyncSourceType) *controldb.SiteSyncSource {
		source, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return source
	}
	require.Zero(t, source(consts.SyncSourceTypeDexcom).TotalRecords)
	require.Equal(t, int64(0), *source(consts.SyncSourceTypeDexcom).LastSyncRecords)

	// The virtual sync's first sync fills the last 30 minutes, so it adds records
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual, Enabled: true}))
	require.NoError(t, jobs.SyncSource(ctx, sync.SyncJobParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual}))
	events := added()
	require.Len(t, events, 1)
	var attributes struct {
		SyncType string `json:"sync_type"`
		Records  int    `json:"records"`
	}
	require.NoError(t, json.Unmarshal([]byte(events[0]), &attributes))
	require.Equal(t, "virtual", attributes.SyncType)
	require.Positive(t, attributes.Records)
	virtual := source(consts.SyncSourceTypeVirtual)
	require.Equal(t, int64(attributes.Records), virtual.TotalRecords)
	require.Equal(t, int64(attributes.Records), *virtual.LastSyncRecords)

	// The next sync's records are its last, and are added to the total
	require.NoError(t, jobs.SyncSource(ctx, sync.SyncJobParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual}))
	next := source(consts.SyncSourceTypeVirtual)
	require.Equal(t, virtual.TotalRecords+*next.LastSyncRecords, next.TotalRecords)
}
