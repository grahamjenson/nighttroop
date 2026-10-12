package backup_sites

import (
	"context"
	"fmt"
	"log/slog"
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
	S3Client     *s3utils.S3Client

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
		S3Client:                     s3Client,
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
	c.AddFunc("BackupSiteJob", fmt.Sprintf("@every %ds", tick), func() error {
		i := int(index.Add(1)-1) % count
		batch := string(utils.SITE_CHAR_SET[i])
		if err := b.DeleteUnusedSites(batch); err != nil {
			slog.Error(
				"Error deleting unused sites",
				"error", err,
				"batch", batch,
			)
		}

		return b.BackupSiteBatch(batch)
	})
}

func (b *BackupSiteJobManager) DeleteUnusedSites(batch string) error {
	log := false
	defer aslog.TrackTime(time.Now(), "DeleteUnusedSites", &log, nil)

	deleteList, err := b.findSitesToDelete(batch)
	if err != nil {
		return fmt.Errorf("cant find sites to delete without archiving %w", err)
	}
	if len(deleteList) == 0 {
		return nil
	}
	log = true
	slog.Info("Deleting site without archiving", "subdomains", deleteList)

	// Just delete these sites without an archive, as they have not been used in so many minutes
	return b.controlDB.DeleteSites(deleteList)
}

// The sites we delete without archiving are sites that
func (b *BackupSiteJobManager) findSitesToDelete(batch string) ([]string, error) {
	sites, err := GetNeverUsedSites(b.controlDB, batch, b.DeleteAfterInactivity)
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
func (b *BackupSiteJobManager) BackupToS3() {
	for _, char := range utils.SITE_CHAR_SET {
		err := b.BackupSiteBatch(string(char))

		if err != nil {
			slog.Error("Error backing up site batch", "error", err, "char", char)
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

func (b *BackupSiteJobManager) BackupSiteBatch(char string) error {
	trackTrim := false
	defer aslog.TrackTime(time.Now(), "backup site pool "+char, &trackTrim, nil)
	ctx := context.TODO()

	sites, err := GetSiteBatch(b.controlDB, char)

	if err != nil {
		return fmt.Errorf("cannot get site batch: %w", err)
	}
	if len(sites) == 0 {
		return nil
	}

	slog.Info("Backing up site batch", "char", char, "site_count", len(sites))

	paths := []string{}

	for _, backupSite := range sites {
		if backupSite.Subdomain == "" {
			continue
		}

		// Have to add it anyway because we need all paths
		err := backupSite.Backup(ctx)
		if err != nil {
			return fmt.Errorf("cannot backup site db %s: %w", backupSite.Subdomain, err)
		}

		paths = append(paths, backupSite.BackupPath)
	}

	trackTrim = len(paths) > 0

	err = b.S3Client.UploadFilesToS3(sitedb.GetS3BatchPath(char, b.ServerConfig), paths)
	if err != nil {
		return fmt.Errorf("cannot sync site pool %s: %w", string(char), err)
	}

	// Copy the control db to the backup location
	return nil
}

// ///
// DB queries
// ////

func GetSiteBatch(controlDB *db.ControlDB, char string) ([]*db.SiteDB, error) {

	sites := []*db.SiteDB{}
	res, err := controlDB.Queries.GetBatchSitesToBackUp(context.TODO(), controldb.GetBatchSitesToBackUpParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}

// GetNeverUsedSites retrieves sites that have never been logged into
func GetNeverUsedSites(controlDB *db.ControlDB, char string, lastSeenMinutes time.Duration) ([]*db.SiteDB, error) {
	res, err := controlDB.Queries.GetSitesToDeleteInBatch(context.TODO(), controldb.GetSitesToDeleteInBatchParams{
		Batch:      char,
		LastSeenAt: lastSeenMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return controlDB.GetSiteDBsFromModels(res)
}
