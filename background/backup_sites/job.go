package backup_sites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
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

// ticksEachInterval is how many times the site job runs in an interval, so a site is backed up soon after it is due
const ticksEachInterval = 26

// SetupCron runs the site job every tick, the interval / 26. A tick deletes the sites never used, backs up the sites
// not uploaded yet, then truncates and backs up those written since their last backup an interval ago. Each runs after
// the one before, so a site is never truncated while its backup is copied.
func (b *BackupSiteJobManager) SetupCron(c cron.CronScheduler) {
	c.AddFunc("BackupSiteJob", fmt.Sprintf("@every %ds", max(1, int(b.Interval.Seconds())/ticksEachInterval)), func(ctx context.Context) error {
		if err := b.DeleteUnusedSites(ctx); err != nil {
			slog.ErrorContext(ctx, "Error deleting unused sites", "error", err)
		}
		b.BackupSitesNotUploaded(ctx)
		return b.BackupWrittenSites(ctx, time.Now().Add(-b.Interval))
	})
}

// BackupSitesNotUploaded backs up each live site not uploaded yet, such as a new or restored site, so it is in S3
// within a tick, not once it is written to. Until then, a new VPS restored from S3 has the site in its control DB but
// not its backup, and cannot start.
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

// DeleteUnusedSites deletes, without archiving, the sites never used in DeleteAfterInactivity
func (b *BackupSiteJobManager) DeleteUnusedSites(ctx context.Context) error {
	log := false
	defer aslog.TrackTime(ctx, time.Now(), "DeleteUnusedSites", &log, nil)

	deleteList, err := b.findSitesToDelete(ctx)
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

// findSitesToDelete are the sites never used in DeleteAfterInactivity, deleted without archiving
func (b *BackupSiteJobManager) findSitesToDelete(ctx context.Context) ([]string, error) {
	sites, err := GetNeverUsedSites(ctx, b.controlDB, b.DeleteAfterInactivity)
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

// BackupWrittenSites backs up the sites written since their last backup, copied before backedUpBefore, and, while S3
// is on, those whose last copy is not uploaded, the longest waiting first, recording the round as an event in the logs DB and a failure as
// an error. It logs the sites it backs up once, then only what goes wrong. A site grows only while written to, so it is
// truncated just before its backup, looking back to its last backup, when it was last truncated.
func (b *BackupSiteJobManager) BackupWrittenSites(ctx context.Context, backedUpBefore time.Time) (err error) {
	started := time.Now()
	defer func() {
		if err != nil {
			b.controlDB.Logs.SiteBackupError(ctx, err)
		}
	}()
	res, err := b.controlDB.Queries.GetSitesToBackUp(ctx, controldb.GetSitesToBackUpParams{
		BackedUpBefore: backedUpBefore.UnixMilli(), Uploading: b.ServerConfig.Status.S3Working, Limit: maxSitesEachTick,
	})
	if err != nil {
		return fmt.Errorf("cannot get the sites to back up: %w", err)
	}
	truncatedAt := map[string]int64{}
	for _, row := range res {
		truncatedAt[row.ID] = 0
		if row.BackedUpAt != nil {
			truncatedAt[row.ID] = *row.BackedUpAt
		}
	}
	sites := b.controlDB.GetSiteDBsFromModels(res)
	if failed := b.controlDB.TruncateSites(ctx, sites, truncatedAt); failed != nil {
		slog.ErrorContext(ctx, "Cannot truncate the sites to back up", "error", failed)
	}
	backedUp := []string{}
	for _, site := range sites {
		if failed := b.backupSite(aslog.WithSite(ctx, site.Subdomain), site); failed != nil {
			err = errors.Join(err, fmt.Errorf("cannot backup site db %s: %w", site.Subdomain, failed))
		} else {
			backedUp = append(backedUp, site.Subdomain)
		}
	}
	if len(backedUp) > 0 {
		slog.InfoContext(ctx, "Backed up site DBs", "sites", backedUp)
		b.controlDB.Logs.SitesBackedUpEvent(ctx, len(backedUp), started, b.ServerConfig.Status.S3Working)
	}
	return err
}

// backupSite copies the site's DB to its backup, and uploads it by itself to S3 when S3 is on
func (b *BackupSiteJobManager) backupSite(ctx context.Context, site *db.SiteDB) error {
	// A copy that fails leaves the last one, as it is written beside it and renamed: for a site archived since this
	// round listed it, that is its archive
	if err := site.Backup(ctx); err != nil {
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
func GetNeverUsedSites(ctx context.Context, controlDB *db.ControlDB, lastSeenMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetSitesToDelete(ctx, controldb.GetSitesToDeleteParams{LastSeenAt: lastSeenMinutes.Minutes()})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res), nil
}
