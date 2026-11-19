package backup_sites_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// fileSite is a new site at id whose DB is a file, as on a server rather than in memory as in other tests
func fileSite(t *testing.T, controlDB *db.ControlDB, id string) *db.SiteDB {
	t.Helper()
	site, err := controlDB.CreateSiteWithID(t.Context(), id)
	require.NoError(t, err)
	cfg := *controlDB.ServerConfig
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	fileDB, err := sitedb.NewDB(t.Context(), id, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { fileDB.Close() })
	require.NoError(t, site.SiteDB.Close())
	site.SiteDB.DB = fileDB
	return site.SiteDB
}

// uploadJobs is the backup jobs for controlDB, with S3 on, with the key and path of each site's upload kept in
// uploads, failing while fail is set
func uploadJobs(t *testing.T, controlDB *db.ControlDB, uploads *[][]string, fail *bool) *backup_sites.BackupSiteJobManager {
	// A copy, as the config is shared by the package's tests
	cfg := *controlDB.ServerConfig
	cfg.Status.S3Working = true
	jobs := backup_sites.SetupBackupSiteJob(controlDB, &cfg, &s3utils.S3Client{ServerConfig: &cfg})
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

// written records the site as last written to at
func written(t *testing.T, controlDB *db.ControlDB, site *db.SiteDB, at time.Time) {
	t.Helper()
	ms := at.UnixMilli()
	require.NoError(t, controlDB.Queries.UpdateSiteDateFlags(t.Context(), controldb.UpdateSiteDateFlagsParams{LastWriteAt: &ms, ID: site.Subdomain}))
}

// stored is the site's row in the control DB
func stored(t *testing.T, controlDB *db.ControlDB, site *db.SiteDB) *controldb.Site {
	t.Helper()
	row, err := controlDB.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: site.Subdomain})
	require.NoError(t, err)
	return row
}

// Only a site written since its last backup, an interval ago, is backed up, each by itself: not one never written, nor
// one written but backed up less than an interval ago, nor one backed up long after its last write
func TestBackupWrittenSitesBacksUpOnlyDirtySites(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	dirty, unwritten, old := fileSite(t, controlDB, "qaaaaaaa"), fileSite(t, controlDB, "qbbbbbbb"), fileSite(t, controlDB, "qccccccc")
	written(t, controlDB, dirty, time.Now())
	written(t, controlDB, old, time.Now().Add(-time.Hour))

	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(-time.Hour)))
	require.ElementsMatch(t, [][]string{upload(dirty), upload(old)}, uploads)
	require.Nil(t, stored(t, controlDB, unwritten).BackedUpAt)

	// Backed up less than an interval ago, so not again, though written since
	written(t, controlDB, dirty, time.Now())
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(-time.Hour)))
	require.Len(t, uploads, 2)

	// An interval later only the site written since is backed up: the old one's last write was long before its backup
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(time.Hour)))
	require.Equal(t, [][]string{upload(dirty)}, uploads[2:])
}

// A site whose upload failed is uploaded again at the next tick, though it is not written since: by the sites not
// uploaded yet, or by the backup of written sites once it was uploaded before
func TestBackupWrittenSitesUploadsAgainAfterAFailedUpload(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, true
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	written(t, controlDB, site, time.Now().Add(-time.Hour))
	require.Error(t, jobs.BackupWrittenSites(ctx, time.Now()))
	require.Nil(t, stored(t, controlDB, site).BackupUploadedAt, "the upload failed")
	fail = false
	jobs.BackupSitesNotUploaded(ctx)
	require.Equal(t, [][]string{upload(site)}, uploads)

	written(t, controlDB, site, time.Now())
	fail = true
	time.Sleep(2 * time.Millisecond)
	require.Error(t, jobs.BackupWrittenSites(ctx, time.Now()))
	fail = false
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(-time.Hour)))
	require.Equal(t, [][]string{upload(site), upload(site)}, uploads)
	require.GreaterOrEqual(t, *stored(t, controlDB, site).BackupUploadedAt, *stored(t, controlDB, site).BackedUpAt)
}

// A new site is uploaded on the next tick, though not written, so a VPS restored from S3 has it
func TestBackupSitesNotUploadedUploadsOnlyNewSites(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	added := fileSite(t, controlDB, "raaaaaaa")
	jobs.BackupSitesNotUploaded(ctx)
	require.Equal(t, [][]string{upload(added)}, uploads)
	require.NotNil(t, stored(t, controlDB, added).BackupUploadedAt)

	// Every site is uploaded, so nothing is uploaded
	jobs.BackupSitesNotUploaded(ctx)
	require.Len(t, uploads, 1)
}

// A model change, such as a connector's sync, records the site as written to, at most every 10 minutes, as its
// requests do, so its backup finds it
func TestModelChangesRecordTheSiteWritten(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	jobs.ServerConfig.Events = events.NewSiteBus()
	controlDB.RecordModelWrites(jobs.ServerConfig.Events)
	site := fileSite(t, controlDB, "qaaaaaaa")
	require.Nil(t, stored(t, controlDB, site).LastWriteAt)

	jobs.ServerConfig.Events.Publish(site.Subdomain, events.ModelChange(site.Subdomain, events.Created, &sitedb.GlucoseReading{}))
	require.Eventually(t, func() bool { return stored(t, controlDB, site).LastWriteAt != nil }, 5*time.Second, 10*time.Millisecond)
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now()))
	require.Equal(t, [][]string{upload(site)}, uploads)
}

// scheduler keeps the jobs added to it, to run them
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name, _ string, job func(context.Context) error)                       { s[name] = job }
func (s scheduler) AddQueueFunc(string, string, *liteq.JobQueue, func(context.Context) error)     {}
func (s scheduler) AddFuncSoon(string, time.Duration, time.Duration, func(context.Context) error) {}

// A tick truncates each site written just before its backup, as a device report from long ago is deleted from it, and
// leaves a site not written as it is, as it does not grow
func TestBackupSiteJobTruncatesOnlyWrittenSites(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	ctx, cron := t.Context(), scheduler{}
	backup_sites.SetupBackupSiteJob(controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg}).SetupCron(cron)
	device, old := "phone", time.Now().AddDate(0, 0, -40).UnixMilli()
	writtenSite, notWritten := fileSite(t, controlDB, "taaaaaaa"), fileSite(t, controlDB, "uaaaaaaa")
	for _, site := range []*db.SiteDB{writtenSite, notWritten} {
		_, err := site.Queries().CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "old", OccurredAt: old, CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	written(t, controlDB, writtenSite, time.Now())

	require.NoError(t, cron["BackupSiteJob"](ctx))
	_, err := writtenSite.Queries().GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "old"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NotNil(t, stored(t, controlDB, writtenSite).BackedUpAt)
	_, err = notWritten.Queries().GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: "old"})
	require.NoError(t, err, "a site not written is not truncated")
}

// While S3 is off, a site uploaded before is copied once an interval like any other, not at each tick as its last copy
// is not uploaded; once S3 is on that copy uploads at the next tick
func TestBackupWrittenSitesWithS3OffCopiesAnUploadedSiteOnceAnInterval(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	written(t, controlDB, site, time.Now())
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now()))
	require.Len(t, uploads, 1)

	jobs.ServerConfig.Status.S3Working = false
	time.Sleep(2 * time.Millisecond)
	written(t, controlDB, site, time.Now())
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now()))
	copied := *stored(t, controlDB, site).BackedUpAt
	require.Greater(t, copied, *stored(t, controlDB, site).BackupUploadedAt)

	time.Sleep(2 * time.Millisecond)
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(-time.Hour)))
	require.Equal(t, copied, *stored(t, controlDB, site).BackedUpAt, "copied again within the interval")

	jobs.ServerConfig.Status.S3Working = true
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now().Add(-time.Hour)))
	require.Len(t, uploads, 2)
}

// A site records when its backup was copied and when it was uploaded: only once S3 has it, not while S3 is off
func TestBackupWrittenSitesRecordsTheBackupOnTheSite(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, uploads, fail := t.Context(), [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	jobs.ServerConfig.Status.S3Working = false
	site := fileSite(t, controlDB, "qaaaaaaa")
	written(t, controlDB, site, time.Now())

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now()))
	backedUp := stored(t, controlDB, site)
	require.NotNil(t, backedUp.BackedUpAt)
	require.GreaterOrEqual(t, *backedUp.BackedUpAt, started)
	require.NotNil(t, backedUp.BackupSizeKb)
	require.Nil(t, backedUp.BackupUploadedAt, "S3 is off")
	require.Empty(t, uploads)
}

func TestBackupWrittenSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx := t.Context()
	jobs := backup_sites.SetupBackupSiteJob(controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	written(t, controlDB, site.SiteDB, time.Now())
	var logs bytes.Buffer
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupWrittenSites(ctx, time.Now()))

	// One log lists the sites backed up, and as nothing went wrong, nothing else is logged
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Len(t, lines, 1, logs.String())
	require.Contains(t, lines[0], `msg="Backed up site DBs" sites=[`+site.SiteDB.Subdomain+`]`)

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sites_backed_up", events[0].Name)
	require.Nil(t, events[0].SiteID, "it is of many sites, so only the control admin sees it")
	var attributes map[string]any
	require.NoError(t, json.Unmarshal([]byte(*events[0].Attributes), &attributes))
	require.Equal(t, float64(1), attributes["site_count"])
	// S3 is not connected in tests, so the sites were backed up but not uploaded
	require.Equal(t, false, attributes["uploaded"])
	require.GreaterOrEqual(t, events[0].StartedAt, started)
	require.NotNil(t, events[0].FinishedAt)
}

func TestDeleteUnusedSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	jobs := backup_sites.SetupBackupSiteJob(controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	// The site has never been used, and counts as unused from the moment it was made
	jobs.DeleteAfterInactivity = -time.Minute

	require.NoError(t, jobs.DeleteUnusedSites(ctx))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_deleted", events[0].Name)
	require.Equal(t, `{"reason":"unused"}`, *events[0].Attributes)
}

// A copy that fails leaves the site's last backup: a site archived after the round listed it has its DB closed, and
// its backup is then its archive, the only one with S3 off
func TestBackupWrittenSitesKeepsTheLastBackupWhenACopyFails(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	uploads, fail := [][]string{}, false
	jobs := uploadJobs(t, controlDB, &uploads, &fail)
	site := fileSite(t, controlDB, "qaaaaaaa")
	written(t, controlDB, site, time.Now())
	require.NoError(t, os.MkdirAll(filepath.Dir(site.BackupPath), 0o700))
	require.NoError(t, os.WriteFile(site.BackupPath, []byte("the archive"), 0o600))
	require.NoError(t, site.Close())

	require.Error(t, jobs.BackupWrittenSites(t.Context(), time.Now()))
	require.FileExists(t, site.BackupPath)
	require.Empty(t, uploads)
}
