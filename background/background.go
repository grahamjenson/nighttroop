package background

import (
	"context"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/backup"
	"nighttroop.t1d.nz/background/liteq"
	ntsync "nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/background/truncate"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

type BackGroundManager struct {
	db     *db.ControlDB
	config config.ServerConfig
	cron   *cron.Cron
	queue  *liteq.JobQueue

	consumers []*liteq.ConsumerSupervisor

	BackupJobManager   *backup.BackupJobManager
	TruncateJobManager *truncate.TruncateJobManager
	SyncJobManager     *ntsync.SyncJobManager
}

func NewBackGroundManager(ctx context.Context, controlDB *db.ControlDB, config config.ServerConfig, s3c *s3utils.S3Client) *BackGroundManager {
	queue := liteq.New(controlDB)
	cron := cron.New(cron.WithSeconds())
	consumers := []*liteq.ConsumerSupervisor{}
	backupjobManager := backup.SetupBackupJob(ctx, controlDB, config, s3c)
	truncatejobManager := truncate.SetupTruncateJobManager(config, controlDB)
	syncjobManager := ntsync.SetupSyncJobManager(config, controlDB)

	// Backup Jobs
	backupjobManager.SetupCron(cron, queue)
	// Backup Should be done by exactly one worker at a time, to avoid over working the CPU and IO
	consumers = append(consumers, backupjobManager.GetConsumer(queue))

	// Truncate Jobs
	truncatejobManager.SetupCron(cron, queue)
	consumers = append(consumers, truncatejobManager.GetConsumer(queue))

	// Sync Jobs
	syncjobManager.SetupCron(cron, queue)
	// Sync jobs can be done in parallel by multiple workers
	consumers = append(consumers, syncjobManager.GetConsumer(queue))

	return &BackGroundManager{
		db:                 controlDB,
		config:             config,
		cron:               cron,
		queue:              queue,
		consumers:          consumers,
		BackupJobManager:   backupjobManager,
		TruncateJobManager: truncatejobManager,
		SyncJobManager:     syncjobManager,
	}
}

func (b *BackGroundManager) Start() {

	for _, consumer := range b.consumers {
		consumer.Start()
	}

	b.cron.Start()
}

func (b *BackGroundManager) Stop() {
	b.cron.Stop()

	for _, consumer := range b.consumers {
		consumer.Stop()
	}

	for _, consumer := range b.consumers {
		consumer.Wait()
	}

}
