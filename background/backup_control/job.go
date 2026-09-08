package backup_control

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/acron"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

const BACKUP_CONTROL_JOB = "backup_control_job"

type BackupControlJobManager struct {
	controlDB *db.ControlDB
	ctx       context.Context
	S3Client  *s3utils.S3Client

	ControlBackupIntervalInSeconds int
}

// BackupJob

// SetupBackupJob Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupControlJob(ctx context.Context, controlDB *db.ControlDB, config config.ServerConfig, s3Client *s3utils.S3Client) *BackupControlJobManager {

	jobs := &BackupControlJobManager{
		controlDB:                      controlDB,
		ctx:                            ctx,
		ControlBackupIntervalInSeconds: config.ControlBackupIntervalInSeconds,
		S3Client:                       s3Client,
	}

	return jobs
}

func (b *BackupControlJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, BACKUP_CONTROL_JOB, 1, b.workerFunc())
}

func (b *BackupControlJobManager) workerFunc() liteq.WorkerFunc {

	return func(ctx context.Context, j *jobs.Job) error {
		return b.BackupControl()
	}
}

func (b *BackupControlJobManager) SetupCron(c acron.CronScheduler, queue *liteq.JobQueue) {

	backupControlCron := fmt.Sprintf("@every %ds", b.ControlBackupIntervalInSeconds)

	c.AddFunc("BackupControlJobManager", backupControlCron, buildBackupControlCronFunc(queue))

}

func buildBackupControlCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		err := queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue:       BACKUP_CONTROL_JOB,
			Job:         "",
			DedupingKey: liteq.ReplaceDuplicate("backup_control"),
		})
		if err != nil {
			slog.Error("Error queueing job", "error", err)
		}
	}
}

func (b *BackupControlJobManager) BackupToS3() {
	err := b.BackupControl()
	if err != nil {
		slog.Error("Backup error", "error", err)
	}
}

func (b *BackupControlJobManager) BackupControl() error {

	// Backup the Database
	err := utils.SqliteBackup(b.ctx, b.controlDB.SQLiteDB, utils.CONTROL_DB_BACKUP_PATH)
	if err != nil {
		return fmt.Errorf("cannot backup control db: %w", err)
	}

	// Copy the control db to S3 path
	err = b.S3Client.UploadFilesToS3(b.S3Client.ControlS3Path(), []string{utils.CONTROL_DB_BACKUP_PATH})
	if err != nil {
		return fmt.Errorf("cannot sync control db: %w", err)
	}

	return nil
}
