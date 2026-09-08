package backup_sites

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/acron"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

const BACKUP_SITE_JOB = "backup_site_job"

type BackupSiteJobParams struct {
	Char string
}

type BackupSiteJobManager struct {
	controlDB *db.ControlDB
	ctx       context.Context
	S3Client  *s3utils.S3Client

	DeleteAfterInactivity time.Duration

	BatchBackupIntervalInSeconds   int
	ControlBackupIntervalInSeconds int
	DeleteUnusedUsersInterval      int
}

// BackupSiteJob

// SetupBackupSiteJob Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupSiteJob(ctx context.Context, controlDB *db.ControlDB, config config.ServerConfig, s3Client *s3utils.S3Client) *BackupSiteJobManager {

	jobs := &BackupSiteJobManager{
		controlDB: controlDB,
		ctx:       ctx,

		DeleteAfterInactivity: time.Duration(config.DeleteAfterInactivityMins) * time.Minute,

		BatchBackupIntervalInSeconds: config.BatchBackupIntervalInSeconds,
		S3Client:                     s3Client,
	}

	return jobs
}

func (b *BackupSiteJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, BACKUP_SITE_JOB, 1, b.workerFunc())
}

func (b *BackupSiteJobManager) workerFunc() liteq.WorkerFunc {

	return func(ctx context.Context, j *jobs.Job) error {
		var params BackupSiteJobParams
		err := json.Unmarshal([]byte(j.Job), &params)
		if err != nil {
			return fmt.Errorf("error unmarshalling job params: %w", err)
		}
		batch := params.Char
		err = b.DeleteUnusedSites(batch)
		if err != nil {
			slog.Error("Error deleting unused sites", "error", err)
			// Continue even if deleting unused sites fails
		}

		return b.BackupSiteBatch(batch)

	}
}

func (b *BackupSiteJobManager) SetupCron(c acron.CronScheduler, queue *liteq.JobQueue) {
	backupBatchCron := fmt.Sprintf("@every %ds", b.BatchBackupIntervalInSeconds)
	c.AddFunc("BackupSiteJobManager", backupBatchCron, b.buildBackupBatchCronFunc(queue))
}

func (b *BackupSiteJobManager) buildBackupBatchCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		for index := range utils.SITE_CHAR_SET {
			char := string(utils.SITE_CHAR_SET[index])
			// Optimization to not queue jobs with no sites
			// If it errors just ignore the query all together
			res, err := HasSitesInBatch(b.controlDB, char)
			if err == nil && !res {
				// Skip spaming the jobs queue
				continue
			}

			jsonParams, err := json.Marshal(BackupSiteJobParams{
				Char: char,
			})
			if err != nil {
				slog.Error("Error marshalling job params", "error", err)
				continue
			}

			err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
				Queue:       BACKUP_SITE_JOB,
				Job:         string(jsonParams),
				DedupingKey: liteq.ReplaceDuplicate(string(jsonParams)),
			})
			if err != nil {
				slog.Error("Error queueing job", "error", err)
			}
		}
	}
}

func (b *BackupSiteJobManager) DeleteUnusedSites(batch string) error {
	log := false
	defer utils.TrackTime(time.Now(), "DeleteUnusedSites", &log, nil)

	deleteList, err := b.findSitesToDelete(batch)
	if err != nil {
		return fmt.Errorf("cant find sites to delete without archiving %w", err)
	}
	if len(deleteList) == 0 {
		return nil
	}
	log = true
	slog.Info("Delete site without archiving", "subdomains", deleteList)

	// Just delete these sites without an archive, as they have not been used in so many minutes
	err = b.controlDB.DeleteSites(deleteList)
	if err != nil {
		slog.Error("Failed to delete site without archiving", "error", err, "sites", deleteList)
	}

	return nil
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
	defer utils.TrackTime(time.Now(), "backup site pool "+char, &trackTrim, nil)
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

	for _, site := range sites {
		if site.Subdomain == "" {
			continue
		}
		_, backupPath := utils.SiteDBPaths(site.Subdomain)

		// Have to add it anyway because we need all paths
		err := utils.SqliteBackup(ctx, site.SQLiteDB, backupPath)
		if err != nil {
			return fmt.Errorf("cannot backup site db %s: %w", site.Subdomain, err)
		}

		paths = append(paths, backupPath)
	}

	trackTrim = len(paths) > 0

	err = b.S3Client.UploadFilesToS3(b.S3Client.SiteBatchS3Path(char), paths)
	if err != nil {
		return fmt.Errorf("cannot sync site pool %s: %w", string(char), err)
	}

	// Copy the control db to the backup location
	return nil
}
