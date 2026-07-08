package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

const (
	ControlBackupJob = "control_backup_job"
	BatchBackupJob   = "batch_backup_job"
	ArchiveBackupJob = "archive_backup_job"
)

const BACKUP_JOB = "backup_batch_job"

type BackupJobParams struct {
	FunctionName string
	Char         string
}

type BackupJobManager struct {
	controlDB *db.ControlDB
	ctx       context.Context
	S3Client  *s3utils.S3Client

	MaxSiteLastSeen     time.Duration
	MaxSiteLastWritten  time.Duration
	MaxSiteNeverSeen    time.Duration
	MaxSiteNeverWritten time.Duration

	BatchBackupIntervalInSeconds   int
	ControlBackupIntervalInSeconds int
	ArchiveIntervalInSeconds       int
}

// BackupJob
// Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupJob(ctx context.Context, controlDB *db.ControlDB, config config.ServerConfig, s3Client *s3utils.S3Client) *BackupJobManager {

	jobs := &BackupJobManager{
		controlDB:                      controlDB,
		ctx:                            ctx,
		MaxSiteLastSeen:                time.Duration(config.MaxSiteLastSeenInMins) * time.Minute,
		MaxSiteLastWritten:             time.Duration(config.MaxSiteLastWrittenInMins) * time.Minute,
		MaxSiteNeverSeen:               time.Duration(config.MaxSiteNeverSeenInMins) * time.Minute,
		MaxSiteNeverWritten:            time.Duration(config.MaxSiteNeverWrittenInMins) * time.Minute,
		BatchBackupIntervalInSeconds:   config.BatchBackupIntervalInSeconds,
		ControlBackupIntervalInSeconds: config.ControlBackupIntervalInSeconds,
		ArchiveIntervalInSeconds:       config.ArchiveIntervalInSeconds,
		S3Client:                       s3Client,
	}

	return jobs
}

func (b *BackupJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, BACKUP_JOB, 1, b.workerFunc())
}

func (b *BackupJobManager) workerFunc() liteq.WorkerFunc {

	return func(ctx context.Context, j *control.Job) error {
		var params BackupJobParams
		err := json.Unmarshal([]byte(j.Job), &params)
		if err != nil {
			return fmt.Errorf("error unmarshalling job params: %w", err)
		}
		switch params.FunctionName {
		case ControlBackupJob:
			return b.BackupControl()
		case BatchBackupJob:
			return b.BackupSiteBatch(params.Char)
		case ArchiveBackupJob:
			return b.ArchiveInactiveSites()
		default:
			return fmt.Errorf("unknown backup job function: %s", params.FunctionName)
		}
	}
}

func (b *BackupJobManager) SetupCron(c *cron.Cron, queue *liteq.JobQueue) {
	backupBatchCron := fmt.Sprintf("@every %ds", b.BatchBackupIntervalInSeconds)
	backupControlCron := fmt.Sprintf("@every %ds", b.ControlBackupIntervalInSeconds)
	archiveCron := fmt.Sprintf("@every %ds", b.ArchiveIntervalInSeconds)

	c.AddFunc(backupBatchCron, b.buildBackupBatchCronFunc(queue))
	c.AddFunc(backupControlCron, buildBackupControlCronFunc(queue))
	c.AddFunc(archiveCron, buildArchiveCronFunc(queue))
}

func (b *BackupJobManager) buildBackupBatchCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		for index := range utils.SITE_CHAR_SET {
			char := string(utils.SITE_CHAR_SET[index])
			// Optimization to not queue jobs with no sites
			// If it errors just ignore the query all together
			res, err := b.controlDB.HasSitesInBatch(char)
			if err == nil && !res {
				// Skip spaming the jobs queue
				continue
			}

			jsonParams, err := json.Marshal(BackupJobParams{
				FunctionName: string(BatchBackupJob),
				Char:         char,
			})
			if err != nil {
				slog.Error("Error marshalling job params", "error", err)
				continue
			}

			err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
				Queue:       BACKUP_JOB,
				Job:         string(jsonParams),
				DedupingKey: liteq.ReplaceDuplicate(string(jsonParams)),
			})
			if err != nil {
				slog.Error("Error queueing job", "error", err)
			}
		}
	}
}

func buildBackupControlCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		jsonParams, err := json.Marshal(BackupJobParams{
			FunctionName: string(ControlBackupJob),
			Char:         "",
		})
		if err != nil {
			slog.Error("Error marshalling job params", "error", err)
			return
		}

		err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue:       BACKUP_JOB,
			Job:         string(jsonParams),
			DedupingKey: liteq.ReplaceDuplicate(string(jsonParams)),
		})
		if err != nil {
			slog.Error("Error queueing job", "error", err)
		}
	}
}

func buildArchiveCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		jsonParams, err := json.Marshal(BackupJobParams{
			FunctionName: string(ArchiveBackupJob),
		})
		if err != nil {
			slog.Error("Error marshalling job params", "error", err)
			return
		}

		err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue:       BACKUP_JOB,
			Job:         string(jsonParams),
			DedupingKey: liteq.ReplaceDuplicate(string(jsonParams)),
		})
		if err != nil {
			slog.Error("Error queueing job", "error", err)
		}
	}
}
