package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// siteDBFiles is a site DB on disk whose file and write-ahead log have the given sizes in bytes, 0 for no file
func siteDBFiles(t *testing.T, fileBytes int, walBytes int) *db.SiteDB {
	path := filepath.Join(t.TempDir(), "site.sqlite")
	for file, size := range map[string]int{path: fileBytes, path + "-wal": walBytes} {
		if size > 0 {
			require.NoError(t, os.WriteFile(file, make([]byte, size), 0o600))
		}
	}
	return &db.SiteDB{DB: &sitedb.DB{Path: path}}
}

func TestEstimatedSizeKB(t *testing.T) {
	tests := []struct {
		name   string
		siteDB *db.SiteDB
		sizeKB *int64
	}{
		{"in memory has no size", &db.SiteDB{DB: &sitedb.DB{InMemory: true}}, nil},
		{"no files", siteDBFiles(t, 0, 0), ptr(0)},
		{"file only", siteDBFiles(t, 4096, 0), ptr(4)},
		{"file and write-ahead log", siteDBFiles(t, 4096, 2048), ptr(6)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.sizeKB, test.siteDB.EstimatedSizeKB())
		})
	}
}

// TestDateFlagsRecordSize checks a site's DB size is recorded with its last write, at most every 10 minutes
func TestDateFlagsRecordSize(t *testing.T) {
	_, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	ctx, siteID := context.TODO(), site.SiteDB.Subdomain
	get := func() *controldb.Site {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return model
	}

	// Seen but not written to measures nothing
	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(get(), siteDBFiles(t, 4096, 0), &controldb.SiteIdentity{}, true, false))
	require.Nil(t, get().EstimatedDbSizeKb)

	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(get(), siteDBFiles(t, 4096, 0), &controldb.SiteIdentity{}, true, true))
	require.Equal(t, ptr(4), get().EstimatedDbSizeKb)

	// Written to again within 10 minutes, so not measured again
	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(get(), siteDBFiles(t, 8192, 0), &controldb.SiteIdentity{}, true, true))
	require.Equal(t, ptr(4), get().EstimatedDbSizeKb)
}

func TestSearchSitesOrders(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, q := context.TODO(), controlDB.Queries
	now := time.Now().UnixMilli()

	// Each site's last write and DB size; nil when there is none yet
	sites := map[string]struct{ lastWriteAt, sizeKB *int64 }{
		"written_today_small":     {ptr(now - 60*60*1000), ptr(10)},
		"written_last_week_large": {ptr(now - 7*24*60*60*1000), ptr(2000)},
		"never_written":           {nil, nil},
	}
	ids := map[string]string{}
	for name, flags := range sites {
		site, err := controlDB.CreateSite()
		require.NoError(t, err)
		ids[site.SiteDB.Subdomain] = name
		require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastWriteAt: flags.lastWriteAt, EstimatedDbSizeKb: flags.sizeKB, ID: site.SiteDB.Subdomain}))
	}
	names := func(rows []*controldb.Site, err error) []string {
		require.NoError(t, err)
		named := []string{}
		for _, row := range rows {
			named = append(named, ids[row.ID])
		}
		return named
	}

	// A site never written to counts from when it was made, which is now
	require.Equal(t, []string{"written_last_week_large", "written_today_small", "never_written"}, names(q.SearchSitesByLastWrite(ctx, controldb.SearchSitesByLastWriteParams{Limit: 10})))
	require.Equal(t, []string{"written_last_week_large", "written_today_small", "never_written"}, names(q.SearchSitesByDBSize(ctx, controldb.SearchSitesByDBSizeParams{Limit: 10})))
}

func ptr(v int64) *int64 { return &v }

func TestSearchSyncSourcesFailedFirst(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, q := context.TODO(), controlDB.Queries
	site, err := controlDB.CreateSite()
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain

	// Nightscout failed first and Dexcom failed last; Virtual never fails, and was synced most recently
	for _, sourceType := range []consts.SyncSourceType{consts.SyncSourceTypeVirtual, consts.SyncSourceTypeNightscout, consts.SyncSourceTypeDexcom} {
		require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: sourceType, Enabled: true}))
	}
	for _, sourceType := range []consts.SyncSourceType{consts.SyncSourceTypeNightscout, consts.SyncSourceTypeDexcom} {
		require.NoError(t, q.UpdateSyncSourceReportsFailure(ctx, controldb.UpdateSyncSourceReportsFailureParams{SiteID: siteID, Type: sourceType, ErrorMessages: errors.ErrorList{"failed"}}))
		time.Sleep(2 * time.Millisecond)
	}
	require.NoError(t, q.UpdateSyncSourceReportsSuccess(ctx, controldb.UpdateSyncSourceReportsSuccessParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual}))

	sources, err := q.SearchSyncSources(ctx, controldb.SearchSyncSourcesParams{Limit: 10})
	require.NoError(t, err)
	types := []consts.SyncSourceType{}
	for _, source := range sources {
		types = append(types, source.Type)
	}
	require.Equal(t, []consts.SyncSourceType{consts.SyncSourceTypeDexcom, consts.SyncSourceTypeNightscout, consts.SyncSourceTypeVirtual}, types)
}
