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
	backupjob_manager := backup.SetupBackupJob(ctx, controlDB, config, s3c)
	truncatejob_manager := truncate.SetupTruncateJobManager(config)
	syncjob_manager := ntsync.SetupSyncJobManager(config, controlDB)

	// Backup Jobs
	backupjob_manager.SetupCron(cron, queue)
	// Backup Should be done by exactly one worker at a time, to avoid over working the CPU and IO
	consumers = append(consumers, backupjob_manager.GetConsumer(queue))

	// Truncate Jobs
	truncatejob_manager.SetupCron(cron, queue)
	consumers = append(consumers, truncatejob_manager.GetConsumer(queue))

	// Sync Jobs
	syncjob_manager.SetupCron(cron, queue)
	// Sync jobs can be done in parallel by multiple workers
	consumers = append(consumers, syncjob_manager.GetConsumer(queue))

	return &BackGroundManager{
		db:                 controlDB,
		config:             config,
		cron:               cron,
		queue:              queue,
		consumers:          consumers,
		BackupJobManager:   backupjob_manager,
		TruncateJobManager: truncatejob_manager,
		SyncJobManager:     syncjob_manager,
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
