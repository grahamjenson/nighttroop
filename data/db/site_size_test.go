package db_test

import (
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
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	get := func() *controldb.Site {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return model
	}

	// Seen but not written to measures nothing
	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(t.Context(), get(), siteDBFiles(t, 4096, 0), &controldb.SiteIdentity{}, true, false))
	require.Nil(t, get().EstimatedDbSizeKb)

	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(t.Context(), get(), siteDBFiles(t, 4096, 0), &controldb.SiteIdentity{}, true, true))
	require.Equal(t, ptr(4), get().EstimatedDbSizeKb)

	// Written to again within 10 minutes, so not measured again
	require.NoError(t, controlDB.UpdateSiteAndIdentityDateFlags(t.Context(), get(), siteDBFiles(t, 8192, 0), &controldb.SiteIdentity{}, true, true))
	require.Equal(t, ptr(4), get().EstimatedDbSizeKb)
}

func TestSearchSitesOrders(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, q := t.Context(), controlDB.Queries
	now := time.Now().UnixMilli()

	// Each site's last write and DB size; nil when there is none yet
	sites := map[string]struct{ lastWriteAt, sizeKB *int64 }{
		"written_today_small":     {ptr(now - 60*60*1000), ptr(10)},
		"written_last_week_large": {ptr(now - 7*24*60*60*1000), ptr(2000)},
		"never_written":           {nil, nil},
	}
	ids := map[string]string{}
	for name, flags := range sites {
		site, err := controlDB.CreateSite(t.Context())
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

func TestSearchConnectorsFailedFirst(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, q := t.Context(), controlDB.Queries
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain

	// Nightscout failed first and Dexcom failed last; Virtual never fails, and was synced most recently
	for _, sourceType := range []consts.ConnectorType{consts.ConnectorTypeVirtual, consts.ConnectorTypeNightscout, consts.ConnectorTypeDexcom} {
		require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: sourceType, Enabled: true}))
	}
	for _, sourceType := range []consts.ConnectorType{consts.ConnectorTypeNightscout, consts.ConnectorTypeDexcom} {
		require.NoError(t, q.UpdateConnectorReportsFailure(ctx, controldb.UpdateConnectorReportsFailureParams{SiteID: siteID, Type: sourceType, ErrorMessages: errors.ErrorList{"failed"}}))
		time.Sleep(2 * time.Millisecond)
	}
	require.NoError(t, q.UpdateConnectorReportsSuccess(ctx, controldb.UpdateConnectorReportsSuccessParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual}))

	sources, err := q.SearchConnectors(ctx, controldb.SearchConnectorsParams{Limit: 10})
	require.NoError(t, err)
	types := []consts.ConnectorType{}
	for _, source := range sources {
		types = append(types, source.Type)
	}
	require.Equal(t, []consts.ConnectorType{consts.ConnectorTypeDexcom, consts.ConnectorTypeNightscout, consts.ConnectorTypeVirtual}, types)
}

// TestRecordSiteSizes checks the resource check keeps each site's DB size, whatever wrote to it, and keeps the last
// size of a DB in memory, which has none to measure
func TestRecordSiteSizes(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	get := func() *int64 {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return model.EstimatedDbSizeKb
	}
	onDisk := func(fileBytes int, walBytes int) *db.SiteDB {
		siteDB := siteDBFiles(t, fileBytes, walBytes)
		siteDB.Subdomain = siteID
		return siteDB
	}

	require.NoError(t, controlDB.RecordSiteSizes(ctx, []*db.SiteDB{onDisk(4096, 0)}))
	require.Equal(t, ptr(4), get())

	// Grown by a sync, with no request to the site
	require.NoError(t, controlDB.RecordSiteSizes(ctx, []*db.SiteDB{onDisk(8192, 2048)}))
	require.Equal(t, ptr(10), get())

	require.NoError(t, controlDB.RecordSiteSizes(ctx, []*db.SiteDB{{DB: &sitedb.DB{InMemory: true}, Subdomain: siteID}}))
	require.Equal(t, ptr(10), get())
}

// A site's size is measured when asked, as its page in control does, keeping the size of its DB on disk; a site whose
// DB is in memory, or not open, keeps the size it has
func TestRecordSiteSize(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	get := func() *int64 {
		model, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return model.EstimatedDbSizeKb
	}

	require.NoError(t, controlDB.RecordSiteSize(ctx, siteID))
	require.Nil(t, get(), "a DB in memory has no size")

	onDisk := siteDBFiles(t, 8192, 2048)
	onDisk.Subdomain = siteID
	controlDB.OpenSite(onDisk)
	require.NoError(t, controlDB.RecordSiteSize(ctx, siteID))
	require.Equal(t, ptr(10), get())

	require.NoError(t, controlDB.RecordSiteSize(ctx, "notopen"))
}

// The resource check scores each site for abuse from its DB's size now and its connectors' highest failure rate, which
// moves a fifth of the way to 1 on each failed sync and to 0 on each sync that works
func TestRecordAbuseScores(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, q := t.Context(), controlDB.Queries
	newSite := func(sizeKB int64) *db.SiteDB {
		site, err := controlDB.CreateSite(ctx)
		require.NoError(t, err)
		site.SiteDB.SetSizeKB(func() *int64 { return &sizeKB })
		return site.SiteDB
	}
	score := func(site *db.SiteDB) float64 {
		model, err := q.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.Subdomain})
		require.NoError(t, err)
		return model.AbuseScore
	}
	large, failing, notScored := newSite(45056), newSite(100), newSite(1<<20)
	for _, typ := range []consts.ConnectorType{consts.ConnectorTypeNightscout, consts.ConnectorTypeDexcom} {
		require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: failing.Subdomain, Type: typ, Enabled: true}))
	}
	fail := controldb.UpdateConnectorReportsFailureParams{SiteID: failing.Subdomain, Type: consts.ConnectorTypeNightscout, ErrorMessages: errors.ErrorList{"failed"}}
	for range 2 {
		require.NoError(t, q.UpdateConnectorReportsFailure(ctx, fail))
	}
	require.NoError(t, q.UpdateConnectorReportsSuccess(ctx, controldb.UpdateConnectorReportsSuccessParams{SiteID: failing.Subdomain, Type: consts.ConnectorTypeNightscout}))
	connector, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: failing.Subdomain, Type: consts.ConnectorTypeNightscout})
	require.NoError(t, err)
	require.InDelta(t, 0.8*0.36, connector.FailureRateEma, 0.0001)

	require.NoError(t, controlDB.RecordAbuseScores(ctx, []*db.SiteDB{large, failing}))
	require.InDelta(t, 0.48, score(large), 0.001) // 2.2 times its 20 MB limit
	require.Equal(t, controldb.AbuseOverUse, controldb.AbuseLevelOf(score(large)))
	require.InDelta(t, 0.288, score(failing), 0.001) // its highest connector's failure rate
	require.Zero(t, score(notScored), "a site not given is not scored")

	// Its connectors syncing again lower its score at the next check
	for range 20 {
		require.NoError(t, q.UpdateConnectorReportsSuccess(ctx, controldb.UpdateConnectorReportsSuccessParams{SiteID: failing.Subdomain, Type: consts.ConnectorTypeNightscout}))
	}
	require.NoError(t, controlDB.RecordAbuseScores(ctx, []*db.SiteDB{failing}))
	require.Less(t, score(failing), 0.01)
}
