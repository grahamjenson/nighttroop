package backup_control

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

type BackupControlJobManager struct {
	controlDB    *db.ControlDB
	ctx          context.Context
	S3Client     *s3utils.S3Client
	ServerConfig *config.ServerConfig

	ControlBackupIntervalInSeconds int
}

// BackupJob

// SetupBackupJob Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupControlJob(ctx context.Context, controlDB *db.ControlDB, config *config.ServerConfig, s3Client *s3utils.S3Client) *BackupControlJobManager {

	jobs := &BackupControlJobManager{
		controlDB:                      controlDB,
		ctx:                            ctx,
		ControlBackupIntervalInSeconds: config.ControlBackupIntervalInSeconds,
		S3Client:                       s3Client,
		ServerConfig:                   config,
	}

	return jobs
}

func (b *BackupControlJobManager) SetupCron(c cron.CronScheduler, queue *liteq.JobQueue) {

	backupControlCron := fmt.Sprintf("@every %ds", b.ControlBackupIntervalInSeconds)

	c.AddFunc("BackupControlJobManager", backupControlCron, func() error {
		return b.BackupToS3()
	})
}

// BackupToS3 backs up the control DB, recording the backup as an event in the logs DB and a failure as an error
func (b *BackupControlJobManager) BackupToS3() (err error) {
	started := time.Now()
	defer func() {
		if err != nil {
			b.controlDB.Logs.ControlBackupError(context.TODO(), err)
			return
		}
		b.controlDB.Logs.ControlBackedUpEvent(context.TODO(), started, b.ServerConfig.Status.S3Working)
	}()

	// Backup the Database
	err = b.controlDB.Backup(context.TODO())
	if err != nil {
		return fmt.Errorf("cannot backup control db: %w", err)
	}

	// Copy the control db to S3 path
	err = b.S3Client.UploadFilesToS3(controldb.GetS3Path(b.ServerConfig), s3utils.ControlRetention, []string{b.controlDB.BackupPath})
	if err != nil {
		return fmt.Errorf("cannot sync control db: %w", err)
	}
	return nil
}
