package backup_sites

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
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
	// Upload uploads a site's backup, the S3 client's UploadFilesToS3, which tests can stand in for
	Upload s3utils.UploadFunc
	// Delete deletes a deleted site's backup, the S3 client's Delete, which tests can stand in for
	Delete func(ctx context.Context, keys []string) error

	DeleteAfterInactivity time.Duration

	BatchBackupIntervalInSeconds int
	DeleteUnusedUsersInterval    int
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
		Delete:                       s3Client.Delete,
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
		// Tests run each job once, so this backs up one batch, but they see every site truncated straight away
		if b.ServerConfig.IsTest() {
			if err := b.controlDB.TruncateSites(ctx, b.controlDB.AllSites()); err != nil {
				slog.ErrorContext(ctx, "Cannot truncate the sites", "error", err)
			}
		}

		b.BackupSitesNotUploaded(ctx)
		return b.BackupSiteBatch(ctx, batch)
	})
}

// BackupSitesNotUploaded backs up each live site not uploaded yet, such as a new or restored site, so it is in S3
// within a tick and not only on its batch's turn, up to an interval later. Until then, a new VPS restored from S3 has
// the site in its control DB but not its backup, and cannot start. It runs in the backup cron's function, so never at
// the same time as a batch's turn.
func (b *BackupSiteJobManager) BackupSitesNotUploaded(ctx context.Context) {
	if !b.ServerConfig.Status.S3Working {
		return
	}
	res, err := b.controlDB.Queries.GetSitesNotUploaded(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot get the sites not uploaded", "error", err)
		return
	}
	for _, site := range b.controlDB.GetSiteDBsFromModels(res) {
		if err := b.backupSite(aslog.WithSite(ctx, site.Subdomain), site); err != nil {
			slog.ErrorContext(ctx, "Error backing up a site not uploaded", "error", err, "site_id", site.Subdomain)
		}
	}
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
	keys := []string{}
	for _, siteID := range deleteList {
		b.controlDB.Logs.SiteDeletedEvent(aslog.WithSite(ctx, siteID), siteID, "unused")
		keys = append(keys, sitedb.GetS3SitePath(siteID, b.ServerConfig))
	}
	// Their backups go too, as each is kept by itself. S3 keeps the old versions for the retention period.
	if b.ServerConfig.Status.S3Working {
		return b.Delete(ctx, keys)
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

// BackupSiteBatch backs up the sites whose IDs start with char, each uploaded by itself, recording the round as an
// event in the logs DB and a failure as an error. It logs the sites it backs up once, then only what goes wrong. A
// site whose backup is current is not copied again, so its DB is not opened and left with an idle connection, and one
// whose copy is already uploaded is not uploaded again.
func (b *BackupSiteJobManager) BackupSiteBatch(ctx context.Context, char string) (err error) {
	ctx = aslog.With(ctx, slog.String("batch", char))
	started := time.Now()
	defer func() {
		if err != nil {
			b.controlDB.Logs.SiteBatchBackupError(ctx, char, err)
		}
	}()

	sites, err := GetSiteBatch(ctx, b.controlDB, char)
	if err != nil && !errors.Is(err, ErrSiteNotOpen) {
		return fmt.Errorf("cannot get site batch: %w", err)
	}
	// Truncated first, so the backup has none of what truncating clears, and a site it changed is copied now. One that
	// cannot be truncated is still backed up.
	if err := b.controlDB.TruncateSites(ctx, sites); err != nil {
		slog.ErrorContext(ctx, "Cannot truncate the batch's sites", "error", err)
	}

	backedUp := []string{}
	var failed error
	for _, backupSite := range sites {
		uploaded, err := b.backupSiteIfChanged(aslog.WithSite(ctx, backupSite.Subdomain), backupSite)
		if err != nil {
			failed = errors.Join(failed, fmt.Errorf("cannot backup site db %s: %w", backupSite.Subdomain, err))
		} else if uploaded {
			backedUp = append(backedUp, backupSite.Subdomain)
		}
	}
	if len(backedUp) > 0 {
		slog.InfoContext(ctx, "Backed up site DBs", "sites", backedUp)
		b.controlDB.Logs.SiteBatchBackedUpEvent(ctx, char, len(backedUp), started, b.ServerConfig.Status.S3Working)
	}
	// A live site that is not open is not backed up, but the rest of its batch is
	return errors.Join(err, failed)
}

// backupSiteIfChanged copies the site's DB to its backup when it changed since its last copy, and uploads the copy
// when it has not been yet, and is whether it uploaded it
func (b *BackupSiteJobManager) backupSiteIfChanged(ctx context.Context, site *db.SiteDB) (bool, error) {
	model, err := b.controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: site.Subdomain})
	if err != nil {
		return false, err
	}
	// The copy was uploaded after it was made, or S3 is off, so there is nothing to send
	uploaded := model.BackupUploadedAt != nil && model.BackedUpAt != nil && *model.BackupUploadedAt >= *model.BackedUpAt
	if backupIsCurrent(site) && (uploaded || !b.ServerConfig.Status.S3Working) {
		return false, nil
	}
	return true, b.backupSite(ctx, site)
}

// backupSite copies the site's DB to its backup, when it changed since, and uploads it by itself to S3 when S3 is on
func (b *BackupSiteJobManager) backupSite(ctx context.Context, site *db.SiteDB) error {
	if !backupIsCurrent(site) {
		if err := copyBackup(ctx, site); err != nil {
			return err
		}
		b.recordBackup(ctx, site)
	}
	if !b.ServerConfig.Status.S3Working {
		return nil
	}
	if err := b.Upload(ctx, sitedb.GetS3SitePath(site.Subdomain, b.ServerConfig), s3utils.SiteRetention, []string{site.BackupPath}); err != nil {
		return fmt.Errorf("cannot upload site %s: %w", site.Subdomain, err)
	}
	if err := b.controlDB.Queries.UpdateSitesBackupUploaded(ctx, controldb.UpdateSitesBackupUploadedParams{IDs: []string{site.Subdomain}}); err != nil {
		slog.ErrorContext(ctx, "Cannot record the site's upload", "error", err)
	}
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

// recordBackup records on the site that its backup was copied now, with the copy's size on disk. Not recording it
// does not stop the backup.
func (b *BackupSiteJobManager) recordBackup(ctx context.Context, site *db.SiteDB) {
	info, err := os.Stat(site.BackupPath)
	if err == nil {
		err = b.controlDB.Queries.UpdateSiteBackedUp(ctx, controldb.UpdateSiteBackedUpParams{ID: site.Subdomain, BackupSizeKb: info.Size() / 1024})
	}
	if err != nil {
		slog.ErrorContext(ctx, "Cannot record the site's backup", "error", err)
	}
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

// ErrSiteNotOpen is a live site left out of its batch's backups because its DB is not open
var ErrSiteNotOpen = errors.New("a live site is not open")

// GetSiteBatch is the open DBs of the live and disabled sites in batch char, with ErrSiteNotOpen if any is not open
func GetSiteBatch(ctx context.Context, controlDB *db.ControlDB, char string) ([]*db.SiteDB, error) {

	sites := []*db.SiteDB{}
	res, err := controlDB.Queries.GetBatchSitesToBackUp(ctx, controldb.GetBatchSitesToBackUpParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	open := controlDB.GetSiteDBsFromModels(res)
	if len(open) != len(res) {
		return open, fmt.Errorf("%w: %d of %d sites in batch %s", ErrSiteNotOpen, len(res)-len(open), len(res), char)
	}
	return open, nil
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
