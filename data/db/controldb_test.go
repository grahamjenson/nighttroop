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
	require.Error(t, controlDB.ArchiveSite(ctx, siteID, func(string, []string) error { return errors.New("S3 is down") }))
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "disabled", row.State)

	// The site's backup is uploaded to its archive path, then it is archived and closed here
	var uploadedKey string
	var uploadedPaths []string
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(key string, paths []string) error {
		uploadedKey, uploadedPaths = key, paths
		return nil
	}))
	require.Equal(t, sitedb.GetS3ArchivePath(siteID, cfg), uploadedKey)
	require.Equal(t, []string{site.SiteDB.BackupPath}, uploadedPaths)
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)
	require.NotNil(t, row.ArchivedAt)
	require.NoFileExists(t, site.SiteDB.BackupPath)

	// It is no longer open, so it cannot be archived again
	require.ErrorContains(t, controlDB.ArchiveSite(ctx, siteID, func(string, []string) error { return nil }), "is not open")
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
