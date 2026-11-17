package backup_sites

import (
	"context"
	"errors"
	"fmt"
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

type BackupSiteJobManager struct {
	controlDB    *db.ControlDB
	ServerConfig *config.ServerConfig
	// Upload uploads a site's backup, the S3 client's UploadFilesToS3, which tests can stand in for
	Upload s3utils.UploadFunc

	DeleteAfterInactivity time.Duration
	// Interval is how old a written site's backup is before it is backed up again
	Interval time.Duration
}

// SetupBackupSiteJob backs up each site written to, once its last backup is an interval old, to S3
func SetupBackupSiteJob(controlDB *db.ControlDB, config *config.ServerConfig, s3Client *s3utils.S3Client) *BackupSiteJobManager {
	return &BackupSiteJobManager{
		controlDB:             controlDB,
		DeleteAfterInactivity: time.Duration(config.DeleteAfterInactivityMins) * time.Minute,
		Interval:              time.Duration(config.SiteBackupIntervalInSeconds) * time.Second,
		Upload:                s3Client.UploadFilesToS3,
		ServerConfig:          config,
	}
}

// SetupCron runs the site job every tick, the interval / 26. A tick first looks after one batch, the sites with the same
// first letter, so each site is seen once an interval: it deletes the batch's sites never used and truncates the rest.
// Then it backs up the sites not uploaded yet, and those written since their last backup an interval ago. The two run
// one after the other, so a site is never truncated while its backup is copied.
func (b *BackupSiteJobManager) SetupCron(c cron.CronScheduler) {
	count := len(utils.SITE_CHAR_SET)
	var index atomic.Int64
	c.AddFunc("BackupSiteJob", fmt.Sprintf("@every %ds", max(1, int(b.Interval.Seconds())/count)), func(ctx context.Context) error {
		b.upkeepBatch(ctx, string(utils.SITE_CHAR_SET[int(index.Add(1)-1)%count]))
		b.BackupSitesNotUploaded(ctx)
		return b.BackupWrittenSites(ctx, time.Now().Add(-b.Interval))
	})
}

// upkeepBatch deletes the batch's sites never used, then truncates its open live and disabled sites. Tests run each job
// once, so there it truncates every site.
func (b *BackupSiteJobManager) upkeepBatch(ctx context.Context, batch string) {
	if err := b.DeleteUnusedSites(ctx, batch); err != nil {
		slog.ErrorContext(ctx, "Error deleting unused sites", "error", err, "batch", batch)
	}
	sites := b.controlDB.AllSites()
	if !b.ServerConfig.IsTest() {
		res, err := b.controlDB.Queries.GetBatchSitesToTruncate(ctx, controldb.GetBatchSitesToTruncateParams{Batch: batch})
		if err != nil {
			slog.ErrorContext(ctx, "Cannot get the batch's sites", "error", err, "batch", batch)
			return
		}
		sites = b.controlDB.GetSiteDBsFromModels(res)
	}
	if err := b.controlDB.TruncateSites(ctx, sites); err != nil {
		slog.ErrorContext(ctx, "Cannot truncate the batch's sites", "error", err, "batch", batch)
	}
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
	for _, siteID := range deleteList {
		b.controlDB.Logs.SiteDeletedEvent(aslog.WithSite(ctx, siteID), siteID, "unused")
	}
	return nil
}

// findSitesToDelete are the batch's sites never used in DeleteAfterInactivity, deleted without archiving
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

// BackupToS3 backs up every site written since its last backup, as the server stops
func (b *BackupSiteJobManager) BackupToS3(ctx context.Context) {
	if err := b.BackupWrittenSites(ctx, time.Now()); err != nil {
		slog.ErrorContext(ctx, "Error backing up the sites", "error", err)
	}
}

// maxSitesEachTick is the most sites one tick backs up, so a burst of writes to many sites is backed up over a few
// ticks rather than copying thousands of DBs at once
const maxSitesEachTick = 500

// BackupWrittenSites backs up the sites written since their last backup, copied before backedUpBefore, and those whose
// last copy is not uploaded, the longest waiting first, recording the round as an event in the logs DB and a failure as
// an error. It logs the sites it backs up once, then only what goes wrong.
func (b *BackupSiteJobManager) BackupWrittenSites(ctx context.Context, backedUpBefore time.Time) (err error) {
	started := time.Now()
	defer func() {
		if err != nil {
			b.controlDB.Logs.SiteBatchBackupError(ctx, "", err)
		}
	}()
	res, err := b.controlDB.Queries.GetSitesToBackUp(ctx, controldb.GetSitesToBackUpParams{BackedUpBefore: backedUpBefore.UnixMilli(), Limit: maxSitesEachTick})
	if err != nil {
		return fmt.Errorf("cannot get the sites to back up: %w", err)
	}
	backedUp := []string{}
	for _, site := range b.controlDB.GetSiteDBsFromModels(res) {
		if failed := b.backupSite(aslog.WithSite(ctx, site.Subdomain), site); failed != nil {
			err = errors.Join(err, fmt.Errorf("cannot backup site db %s: %w", site.Subdomain, failed))
		} else {
			backedUp = append(backedUp, site.Subdomain)
		}
	}
	if len(backedUp) > 0 {
		slog.InfoContext(ctx, "Backed up site DBs", "sites", backedUp)
		b.controlDB.Logs.SiteBatchBackedUpEvent(ctx, "", len(backedUp), started, b.ServerConfig.Status.S3Working)
	}
	return err
}

// backupSite copies the site's DB to its backup, and uploads it by itself to S3 when S3 is on
func (b *BackupSiteJobManager) backupSite(ctx context.Context, site *db.SiteDB) error {
	if err := copyBackup(ctx, site); err != nil {
		return err
	}
	b.recordBackup(ctx, site)
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

// copyBackup copies site's DB to its backup. A copy that failed part way goes.
func copyBackup(ctx context.Context, site *db.SiteDB) error {
	err := site.Backup(ctx)
	if err != nil {
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
