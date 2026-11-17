package backup_sites_test

import (
	"bytes"
	"context"
	"database/sql"
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

// uploadJobs is the backup jobs for controlDB, with S3 on, with the key and path of each site's upload kept in
// uploads, failing while fail is set
func uploadJobs(t *testing.T, controlDB *db.ControlDB, uploads *[][]string, fail *bool) *backup_sites.BackupSiteJobManager {
	// A copy, as the config is shared by the package's tests
	cfg := *controlDB.ServerConfig
	cfg.Status.S3Working = true
	jobs := backup_sites.SetupBackupSiteJob(t.Context(), controlDB, &cfg, &s3utils.S3Client{ServerConfig: &cfg})
	jobs.Upload = func(_ context.Context, key string, _ s3utils.Retention, paths []string) error {
		if *fail {
			return errors.New("S3 is down")
		}
		*uploads = append(*uploads, append([]string{key}, paths...))
		return nil
	}
	return jobs
}

// upload is the key and path of the site's upload
func upload(site *db.SiteDB) []string {
	return []string{sitedb.GetS3SitePath(site.Subdomain, site.ServerConfig), site.BackupPath}
}

// Each site of a batch is uploaded by itself, and only once it changed
func TestBackupSiteBatchCopiesAndUploadsOnlyWhatChanged(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	written, unwritten := fileSite(t, controlDB, "qaaaaaaa"), fileSite(t, controlDB, "qbbbbbbb")

	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.ElementsMatch(t, [][]string{upload(written), upload(unwritten)}, uploads)
	writtenAt, unwrittenAt := backupTime(t, written), backupTime(t, unwritten)

	// Nothing was written, so neither site is copied, opening its DB, or uploaded again
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 2)
	require.Equal(t, writtenAt, backupTime(t, written))
	require.Equal(t, unwrittenAt, backupTime(t, unwritten))

	// Only the site written to is copied and uploaded again
	_, err := written.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "new", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 3)
	require.Equal(t, upload(written), uploads[2])
	require.NotEqual(t, writtenAt, backupTime(t, written))
	require.Equal(t, unwrittenAt, backupTime(t, unwritten))
}

// A backup truncates its batch's sites first, so it has none of what truncating clears
func TestBackupSiteBatchTruncatesItsSitesFirst(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site, device := fileSite(t, controlDB, "taaaaaaa"), "phone"
	old := time.Now().AddDate(0, 0, -40).UnixMilli()
	_, err := site.Queries().CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "old", OccurredAt: old, CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	require.NoError(t, jobs.BackupSiteBatch(ctx, "t"))
	require.Len(t, uploads, 1)
	_, err = site.Queries().GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "old"})
	require.ErrorIs(t, err, sql.ErrNoRows)
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

// A new site is uploaded on the next tick, not only on its batch's turn, so a VPS restored from S3 has it
func TestBackupSitesNotUploadedUploadsOnlyNewSites(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	uploaded := fileSite(t, controlDB, "qaaaaaaa")
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)

	added := fileSite(t, controlDB, "raaaaaaa")
	jobs.BackupSitesNotUploaded(ctx)
	require.Equal(t, [][]string{upload(uploaded), upload(added)}, uploads)
	stored, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: added.Subdomain})
	require.NoError(t, err)
	require.NotNil(t, stored.BackupUploadedAt)

	// Every site is uploaded, so nothing is uploaded
	jobs.BackupSitesNotUploaded(ctx)
	require.Len(t, uploads, 2)
}

func TestBackupSiteBatchUploadsAgainAfterAFailedUpload(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, true
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	require.Error(t, jobs.BackupSiteBatch(ctx, "q"))
	copiedAt := backupTime(t, site)

	// The site's copy is current, but it was never uploaded, so it is uploaded now
	fail = false
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Len(t, uploads, 1)
	require.Equal(t, copiedAt, backupTime(t, site))
}

// A site records when its backup was copied and the copy's size, and when it was uploaded: only once S3 has it,
// not while S3 is off, as the upload then sends nothing, nor when the upload fails
func TestBackupSiteBatchRecordsTheBackupOnTheSite(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	cfg := jobs.ServerConfig
	cfg.Status.S3Working = false
	site := fileSite(t, controlDB, "qaaaaaaa")
	stored := func() *controldb.Site {
		stored, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.Subdomain})
		require.NoError(t, err)
		return stored
	}
	require.Nil(t, stored().BackedUpAt)

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	backedUp := stored()
	require.NotNil(t, backedUp.BackedUpAt)
	require.GreaterOrEqual(t, *backedUp.BackedUpAt, started)
	info, err := os.Stat(site.BackupPath)
	require.NoError(t, err)
	require.NotNil(t, backedUp.BackupSizeKb)
	require.Equal(t, info.Size()/1024, *backedUp.BackupSizeKb)
	require.Nil(t, backedUp.BackupUploadedAt, "S3 is off")

	// Written to, so it is copied and uploaded again, but the upload fails
	cfg.Status.S3Working, fail = true, true
	_, err = site.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "new", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	require.Error(t, jobs.BackupSiteBatch(ctx, "q"))
	require.Greater(t, *stored().BackedUpAt, *backedUp.BackedUpAt)
	require.Nil(t, stored().BackupUploadedAt, "the upload failed")

	fail = false
	require.NoError(t, jobs.BackupSiteBatch(ctx, "q"))
	require.NotNil(t, stored().BackupUploadedAt)
	require.GreaterOrEqual(t, *stored().BackupUploadedAt, *stored().BackedUpAt)
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
	require.Contains(t, lines[0], `msg="Backed up site DBs" sites=[`+site.SiteDB.Subdomain+`]`)

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

// A live site that is not open is left out of its batch's backups, while the rest of the batch is backed up
func TestBackupSiteBatchBacksUpTheRestWithoutALiveSiteNotOpen(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	_, err := controlDB.Queries.CreateSite(ctx, controldb.CreateSiteParams{ID: "qbbbbbbb", BackupKeyHash: "hash"})
	require.NoError(t, err)

	require.ErrorIs(t, jobs.BackupSiteBatch(ctx, "q"), backup_sites.ErrSiteNotOpen)
	require.Equal(t, [][]string{upload(site)}, uploads)
}
