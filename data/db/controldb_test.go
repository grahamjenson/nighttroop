package db_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// Sites created at the same time must each get their own subdomain and site DB
func TestCreateSiteConcurrently(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)

	const count = 10
	subdomains := make([]string, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := controlDB.CreateSite()
			errs[i] = err
			if err == nil {
				subdomains[i] = result.SiteDB.Subdomain
			}
		}()
	}
	wg.Wait()

	seen := map[string]bool{}
	for i := range count {
		require.NoError(t, errs[i])
		require.False(t, seen[subdomains[i]], "subdomain %s created twice", subdomains[i])
		seen[subdomains[i]] = true

		siteDB, _, err := controlDB.GetSite(subdomains[i])
		require.NoError(t, err, "site %s was not created", subdomains[i])
		_, err = siteDB.Queries.GetMostRecentProfileStore(context.TODO())
		require.NoError(t, err, "site %s DB is not usable", subdomains[i])
	}
}

// A site DB that cannot be created is an error, not a panic
func TestCreateSiteWhenSiteDBFails(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)

	// A file where the data folder should be, so the site DB cannot be written
	badCfg := *cfg
	badCfg.Mode = consts.Development
	badCfg.DataDir = filepath.Join(t.TempDir(), "not-a-folder")
	require.NoError(t, os.WriteFile(badCfg.DataDir, nil, 0o600))
	controlDB.ServerConfig = &badCfg

	_, err := controlDB.CreateSite()
	require.ErrorContains(t, err, "Site DB Creation Error")
}

func TestArchiveSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := context.TODO(), site.SiteDB.Subdomain
	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)

	// An upload that fails keeps the site
	require.Error(t, controlDB.ArchiveSite(ctx, siteID, func(string, s3utils.Retention, []string) error { return errors.New("S3 is down") }))
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "disabled", row.State)

	// The site's backup is uploaded to its archive path, then it is archived and closed here
	var uploadedKey string
	var uploadedRetention s3utils.Retention
	var uploadedPaths []string
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(key string, retention s3utils.Retention, paths []string) error {
		uploadedKey, uploadedRetention, uploadedPaths = key, retention, paths
		return nil
	}))
	require.Equal(t, sitedb.GetS3ArchivePath(siteID, cfg), uploadedKey)
	require.Equal(t, s3utils.SiteRetention, uploadedRetention)
	require.Equal(t, []string{site.SiteDB.BackupPath}, uploadedPaths)
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)
	require.NotNil(t, row.ArchivedAt)
	require.NoFileExists(t, site.SiteDB.BackupPath)

	// It is no longer open, so it cannot be archived again
	require.ErrorContains(t, controlDB.ArchiveSite(ctx, siteID, func(string, s3utils.Retention, []string) error { return nil }), "is not open")
}

// An archived site hydrates from its archive with its data, live again and open, and only an archived one does
func TestHydrateSite(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	// Site DBs in files, so the archive holds the site's data
	fileCfg := *cfg
	fileCfg.Mode = consts.Development
	fileCfg.DataDir = t.TempDir()
	controlDB.ServerConfig = &fileCfg
	ctx := context.TODO()
	site, err := controlDB.CreateSite()
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain
	_, err = site.SiteDB.Queries.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	// A stand in for S3, holding the archive's file
	archive := filepath.Join(t.TempDir(), siteID+".sqlite")
	download := func(key string, outpath string) (int64, bool, error) {
		require.Equal(t, sitedb.GetS3ArchivePath(siteID, &fileCfg), key)
		data, err := os.ReadFile(archive)
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		require.NoError(t, err)
		return int64(len(data)), true, os.WriteFile(filepath.Join(outpath, siteID+".sqlite"), data, 0o600)
	}

	// A live site, open here, is not hydrated
	require.ErrorContains(t, controlDB.HydrateSite(ctx, siteID, download), "already open")

	_, err = controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(_ string, _ s3utils.Retention, paths []string) error {
		data, err := os.ReadFile(paths[0])
		require.NoError(t, err)
		return os.WriteFile(archive, data, 0o600)
	}))
	require.NoFileExists(t, sitedb.GetPath(siteID, &fileCfg))

	// A download that fails, or finds no archive, leaves it archived
	require.Error(t, controlDB.HydrateSite(ctx, siteID, func(string, string) (int64, bool, error) { return 0, false, errors.New("S3 is down") }))
	require.ErrorContains(t, controlDB.HydrateSite(ctx, siteID, func(string, string) (int64, bool, error) { return 0, false, nil }), "has no archive")
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)

	// Hydrated, it is live and open with its data, as if just used so it is not disabled again straight away
	require.NoError(t, controlDB.HydrateSite(ctx, siteID, download))
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "live", row.State)
	require.Nil(t, row.ArchivedAt)
	require.Nil(t, row.DisabledAt)
	require.NotNil(t, row.LastSeenAt)
	require.InDelta(t, time.Now().UnixMilli(), *row.LastSeenAt, 60_000)
	siteDB, _, err := controlDB.GetSite(siteID)
	require.NoError(t, err)
	reading, err := siteDB.Queries.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, 120.0, reading.Mgdl)
}

func TestTruncateSite(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := context.TODO(), site.SiteDB.Queries
	raw := `{"sgv":120,"device":"phone"}`
	reading, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "reading", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, RawJsonData: &raw, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NotNil(t, reading.RawJsonData)
	recent, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "recent", Mgdl: 130, ReadingType: "sgv", OccurredAt: time.Now().AddDate(0, 0, -controlDB.ServerConfig.TruncateRecordsAfterDays+1).UnixMilli(), RawJsonData: &raw, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	// The old reading's raw JSON is cleared and the DB shrunk; an in-memory DB has no size to record
	sizeKB, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.Nil(t, sizeKB)
	reading, err = q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: "reading"})
	require.NoError(t, err)
	require.Nil(t, reading.RawJsonData)
	require.Equal(t, float64(120), reading.Mgdl, "the reading itself is kept")
	recent, err = q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: recent.ID})
	require.NoError(t, err)
	require.Equal(t, raw, *recent.RawJsonData, "a reading newer than TruncateRecordsAfterDays keeps its raw JSON")

	_, err = controlDB.TruncateSite(ctx, "not-a-site")
	require.ErrorContains(t, err, "is not open")
}
