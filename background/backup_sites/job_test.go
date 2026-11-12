package backup_sites_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// fileSite is a new site at id whose DB is a file, as on a server rather than in memory as in other tests, last
// written an hour ago. Its connections close as soon as they are idle, as they would well before the next round.
func fileSite(t *testing.T, controlDB *db.ControlDB, id string) *db.SiteDB {
	t.Helper()
	site, err := controlDB.CreateSiteWithID(t.Context(), id)
	require.NoError(t, err)
	cfg := *controlDB.ServerConfig
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	fileDB, err := sitedb.NewDB(t.Context(), id, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { fileDB.Close() })
	fileDB.SQLiteDB.SetMaxIdleConns(0)
	require.NoError(t, site.SiteDB.Close())
	site.SiteDB.DB = fileDB
	hourAgo := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(fileDB.Path, hourAgo, hourAgo))
	return site.SiteDB
}

// backupTime is when site's backup was last copied
func backupTime(t *testing.T, site *db.SiteDB) time.Time {
	t.Helper()
	info, err := os.Stat(site.BackupPath)
	require.NoError(t, err)
	return info.ModTime()
}

// uploadJobs is the backup jobs for controlDB, with the paths of each batch upload kept in uploads, failing while fail
// is set
func uploadJobs(t *testing.T, controlDB *db.ControlDB, uploads *[][]string, fail *bool) *backup_sites.BackupSiteJobManager {
	cfg := controlDB.ServerConfig
	jobs := backup_sites.SetupBackupSiteJob(t.Context(), controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	jobs.Upload = func(_ context.Context, _ string, _ s3utils.Retention, paths []string) error {
		if *fail {
			return errors.New("S3 is down")
		}
		*uploads = append(*uploads, paths)
		return nil
	}
	return jobs
}

func TestBackupSiteBatchCopiesAndUploadsOnlyWhatChanged(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	written, unwritten := fileSite(t, controlDB, "qaaaaaaa"), fileSite(t, controlDB, "qbbbbbbb")

	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)
	writtenAt, unwrittenAt := backupTime(t, written), backupTime(t, unwritten)

	// Nothing was written, so neither site is copied, opening its DB, and the batch is not uploaded again
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)
	require.Equal(t, writtenAt, backupTime(t, written))
	require.Equal(t, unwrittenAt, backupTime(t, unwritten))

	// Only the site written to is copied again, and the batch is uploaded with both sites
	_, err := written.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "new", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 2)
	require.ElementsMatch(t, []string{written.BackupPath, unwritten.BackupPath}, uploads[1])
	require.NotEqual(t, writtenAt, backupTime(t, written))
	require.Equal(t, unwrittenAt, backupTime(t, unwritten))
}

// A site only read since its backup is not copied again, though while its connection is open its write-ahead log is
// there, empty, and newer than the backup
func TestBackupSiteBatchSkipsASiteOnlyRead(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	read := fileSite(t, controlDB, "qaaaaaaa")
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	readAt := backupTime(t, read)

	// A read keeps its connection idle, with the empty write-ahead log it made dated now
	read.SQLiteDB.SetMaxIdleConns(1)
	_, err := read.Queries().SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{Limit: 1})
	require.NoError(t, err)
	wal, err := os.Stat(read.Path + "-wal")
	require.NoError(t, err)
	require.Zero(t, wal.Size())
	require.True(t, wal.ModTime().After(readAt))

	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Equal(t, readAt, backupTime(t, read))
	require.Len(t, uploads, 1)
}

func TestBackupSiteBatchUploadsAgainWhenItsSitesChange(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	first := fileSite(t, controlDB, "qaaaaaaa")
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)

	// A site added to the batch is uploaded with it
	added := fileSite(t, controlDB, "qbbbbbbb")
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 2)
	require.ElementsMatch(t, []string{first.BackupPath, added.BackupPath}, uploads[1])

	// A site taken out of the batch is too, though no other site is copied
	require.NoError(t, controlDB.DeleteSites(ctx, []string{added.Subdomain}))
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 3)
	require.Equal(t, []string{first.BackupPath}, uploads[2])
}

// A batch whose last site is deleted or archived is uploaded empty, so S3's latest backup of it no longer has the
// site's data, while a batch that never had a site is not uploaded
func TestBackupSiteBatchUploadsEmptyOnceItsLastSiteGoes(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Empty(t, uploads)

	site := fileSite(t, controlDB, "qaaaaaaa")
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)

	require.NoError(t, controlDB.DeleteSites(ctx, []string{site.Subdomain}))
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 2)
	require.Empty(t, uploads[1])

	// The empty batch is not uploaded again
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 2)
}

func TestBackupSiteBatchUploadsAgainAfterAFailedUpload(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, true
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	require.Error(t, jobs.BackupSiteBatch(ctx, "q"))
	copiedAt := backupTime(t, site)

	// The site's copy is current, but the batch was never uploaded, so it is uploaded now
	fail = false
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)
	require.Equal(t, copiedAt, backupTime(t, site))
}

func TestBackupSiteBatchRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, batch := t.Context(), site.SiteDB.Subdomain[:1]
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	var logs bytes.Buffer
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupSiteBatch(ctx, batch))

	// One log lists the sites backed up, and as nothing went wrong, nothing else is logged
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Len(t, lines, 1, logs.String())
	require.Contains(t, lines[0], `msg="Backing up site DBs" sites=[`+site.SiteDB.Subdomain+`]`)

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_batch_backed_up", events[0].Name)
	require.Nil(t, events[0].SiteID)
	require.Equal(t, batch, *events[0].Batch, "so each site in the batch can see it")
	var attributes map[string]any
	require.NoError(t, json.Unmarshal([]byte(*events[0].Attributes), &attributes))
	require.Equal(t, batch, attributes["batch"])
	require.Equal(t, float64(1), attributes["site_count"])
	// S3 is not connected in tests, so the batch was backed up but not uploaded
	require.Equal(t, false, attributes["uploaded"])
	// It started when the backup did, and finished when it was recorded
	require.GreaterOrEqual(t, events[0].StartedAt, started)
	require.NotNil(t, events[0].FinishedAt)
	require.GreaterOrEqual(t, *events[0].FinishedAt, events[0].StartedAt)
}

func TestDeleteUnusedSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	// The site has never been used, and counts as unused from the moment it was made
	jobs.DeleteAfterInactivity = -time.Minute

	require.NoError(t, jobs.DeleteUnusedSites(ctx, siteID[:1]))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_deleted", events[0].Name)
	require.Equal(t, `{"reason":"unused"}`, *events[0].Attributes)
}

// A batch is not uploaded without a live site that is not open, as S3's latest backup of it would then lack that site
func TestBackupSiteBatchNotUploadedWithoutALiveSiteNotOpen(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	fileSite(t, controlDB, "qaaaaaaa")
	_, err := controlDB.Queries.CreateSite(ctx, controldb.CreateSiteParams{ID: "qbbbbbbb", BackupKeyHash: "hash"})
	require.NoError(t, err)

	require.ErrorIs(t, jobs.BackupSiteBatch(ctx, "q"), backup_sites.ErrSiteNotOpen)
	require.Empty(t, uploads)
}
