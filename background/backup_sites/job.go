package backup_sites

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

const BACKUP_SITE_JOB = "backup_site_job"

type BackupSiteJobParams struct {
	Char string
}

type BackupSiteJobManager struct {
	controlDB    *db.ControlDB
	ctx          context.Context
	ServerConfig *config.ServerConfig
	// Upload uploads a batch's backups, the S3 client's UploadFilesToS3, which tests can stand in for
	Upload s3utils.UploadFunc

	// uploaded is each batch's sorted, comma joined site IDs, kept once its upload succeeds and dropped as soon as one
	// of its sites' backups is copied again, so while it is kept S3 has the batch as its backups are now
	uploaded sync.Map

	DeleteAfterInactivity time.Duration

	BatchBackupIntervalInSeconds   int
	ControlBackupIntervalInSeconds int
	DeleteUnusedUsersInterval      int
}

// BackupSiteJob

// SetupBackupSiteJob Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupSiteJob(ctx context.Context, controlDB *db.ControlDB, config *config.ServerConfig, s3Client *s3utils.S3Client) *BackupSiteJobManager {

	jobs := &BackupSiteJobManager{
		controlDB: controlDB,
		ctx:       ctx,

		DeleteAfterInactivity: time.Duration(config.DeleteAfterInactivityMins) * time.Minute,

		BatchBackupIntervalInSeconds: config.BatchBackupIntervalInSeconds,
		Upload:                       s3Client.UploadFilesToS3,
		ServerConfig:                 config,
	}

	return jobs
}

func (b *BackupSiteJobManager) SetupCron(c cron.CronScheduler) {
	count := len(utils.SITE_CHAR_SET)

	tick := max(1, b.BatchBackupIntervalInSeconds/count)

	var index atomic.Int64

	// The goal here is to smooth out resource usage, if you looped over all baches at once, we would use a ton of CPU and that will slow down responses
	// This nice thing about this method of backing up is that because the cron task cannot run concurrently for the same batch,
	// we avoid using too many resources, instead it will just delay the backup by tick time
	c.AddFunc("BackupSiteJob", fmt.Sprintf("@every %ds", tick), func(ctx context.Context) error {
		i := int(index.Add(1)-1) % count
		batch := string(utils.SITE_CHAR_SET[i])
		if err := b.DeleteUnusedSites(ctx, batch); err != nil {
			slog.ErrorContext(ctx, "Error deleting unused sites", "error", err, "batch", batch)
		}

		return b.BackupSiteBatch(ctx, batch)
	})
}

func (b *BackupSiteJobManager) DeleteUnusedSites(ctx context.Context, batch string) error {
	log := false
	defer aslog.TrackTime(ctx, time.Now(), "DeleteUnusedSites", &log, nil)

	deleteList, err := b.findSitesToDelete(ctx, batch)
	if err != nil {
		return fmt.Errorf("cant find sites to delete without archiving %w", err)
	}
	if len(deleteList) == 0 {
		return nil
	}
	log = true
	slog.InfoContext(ctx, "Deleting site without archiving", "subdomains", deleteList)

	// Just delete these sites without an archive, as they have not been used in so many minutes
	if err := b.controlDB.DeleteSites(ctx, deleteList); err != nil {
		return err
	}
	for _, siteID := range deleteList {
		b.controlDB.Logs.SiteDeletedEvent(aslog.WithSite(ctx, siteID), siteID, "unused")
	}
	return nil
}

// The sites we delete without archiving are sites that
func (b *BackupSiteJobManager) findSitesToDelete(ctx context.Context, batch string) ([]string, error) {
	sites, err := GetNeverUsedSites(ctx, b.controlDB, batch, b.DeleteAfterInactivity)
	if err != nil {
		return nil, fmt.Errorf("cant get never seen sites %w", err)
	}

	if len(sites) == 0 {
		return nil, nil
	}

	deleteSites := []string{}
	for _, site := range sites {
		if site.Subdomain == "" {
			continue
		}
		deleteSites = append(deleteSites, site.Subdomain)
	}

	return deleteSites, nil
}

// BackupToS3 This is the last method called to backup both control and site dbs
func (b *BackupSiteJobManager) BackupToS3(ctx context.Context) {
	for _, char := range utils.SITE_CHAR_SET {
		err := b.BackupSiteBatch(ctx, string(char))

		if err != nil {
			slog.ErrorContext(ctx, "Error backing up site batch", "error", err, "char", string(char))
			continue
		}
	}
}

// //
// Control Backup
// //

// ///
// Site Backup
// ///

// BackupSiteBatch backs up the sites whose IDs start with char, recording the backup as an event in the logs DB and
// a failure as an error. It logs the sites it backs up once, then only what goes wrong. A site whose backup is
// current is not copied again, so its DB is not opened and left with an idle connection, and a batch with the same
// sites, none copied again, since this process last uploaded it is not uploaded or recorded again.
func (b *BackupSiteJobManager) BackupSiteBatch(ctx context.Context, char string) (err error) {
	ctx = aslog.With(ctx, slog.String("batch", char))
	started := time.Now()
	defer func() {
		if err != nil {
			b.controlDB.Logs.SiteBatchBackupError(ctx, char, err)
		}
	}()

	sites, err := GetSiteBatch(ctx, b.controlDB, char)

	if err != nil {
		return fmt.Errorf("cannot get site batch: %w", err)
	}
	if len(sites) == 0 {
		return nil
	}

	siteIDs, paths := []string{}, []string{}
	for _, backupSite := range sites {
		if backupSite.Subdomain == "" {
			continue
		}
		// Have to add it anyway because we need all paths
		siteIDs = append(siteIDs, backupSite.Subdomain)
		paths = append(paths, backupSite.BackupPath)
		if backupIsCurrent(backupSite) {
			continue
		}
		// Dropped before copying, so a round that fails after this still uploads the batch the next time
		b.uploaded.Delete(char)
		if err := copyBackup(aslog.WithSite(ctx, backupSite.Subdomain), backupSite); err != nil {
			return fmt.Errorf("cannot backup site db %s: %w", backupSite.Subdomain, err)
		}
	}

	batchSites := strings.Join(slices.Sorted(slices.Values(siteIDs)), ",")
	if last, ok := b.uploaded.Load(char); ok && last == batchSites {
		return nil
	}
	slog.InfoContext(ctx, "Backing up site DBs", "sites", siteIDs)

	err = b.Upload(ctx, sitedb.GetS3BatchPath(char, b.ServerConfig), s3utils.SiteRetention, paths)
	if err != nil {
		return fmt.Errorf("cannot sync site pool %s: %w", string(char), err)
	}
	b.uploaded.Store(char, batchSites)
	b.controlDB.Logs.SiteBatchBackedUpEvent(ctx, char, len(paths), started, b.ServerConfig.Status.S3Working)

	// Copy the control db to the backup location
	return nil
}

// copySlack is how long before a copy starts its backup is dated, covering file systems that keep times to the second
// and the kernel's coarse clock, which sets a file's time, lagging behind time.Now
const copySlack = 5 * time.Second

// copyBackup copies site's DB to its backup, dated a little before the copy started rather than when it finished. A
// write the copy may have missed is made after the copy started, and making it sets the DB file's or its write-ahead
// log's time to then, newer than the backup, so the next round copies the site again: no write is left out for good.
func copyBackup(ctx context.Context, site *db.SiteDB) error {
	dated := time.Now().Add(-copySlack)
	err := site.Backup(ctx)
	if err == nil {
		err = os.Chtimes(site.BackupPath, dated, dated)
	}
	if err != nil {
		// A copy that failed part way, or is dated when it finished, could pass for a current one, so it goes
		_ = utils.DeleteFile(site.BackupPath)
	}
	return err
}

// backupIsCurrent is whether site's backup is dated after its DB file and write-ahead log were last written, so a
// copy would change nothing. A time equal to the backup's could be a write it missed, so is copied again. It reads
// only the files' times, opening no connection to the DB.
func backupIsCurrent(site *db.SiteDB) bool {
	backup, err := os.Stat(site.BackupPath)
	if site.InMemory || err != nil {
		return false
	}
	for _, path := range []string{site.Path, site.Path + "-wal"} {
		info, err := os.Stat(path)
		// The write-ahead log is removed when the DB's last connection closes, once what it held is in the DB file, and is
		// empty, though dated when it was opened, while a connection that has written nothing is open, as one that read
		if path != site.Path && (errors.Is(err, fs.ErrNotExist) || (err == nil && info.Size() == 0)) {
			continue
		}
		if err != nil || !info.ModTime().Before(backup.ModTime()) {
			return false
		}
	}
	return true
}

// ///
// DB queries
// ////

func GetSiteBatch(ctx context.Context, controlDB *db.ControlDB, char string) ([]*db.SiteDB, error) {

	sites := []*db.SiteDB{}
	res, err := controlDB.Queries.GetBatchSitesToBackUp(ctx, controldb.GetBatchSitesToBackUpParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res), nil
}

// GetNeverUsedSites retrieves sites that have never been logged into
func GetNeverUsedSites(ctx context.Context, controlDB *db.ControlDB, char string, lastSeenMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetSitesToDeleteInBatch(ctx, controldb.GetSitesToDeleteInBatchParams{
		Batch:      char,
		LastSeenAt: lastSeenMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res), nil
}
