package background

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/background/backup_control"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/background/summarise_jobs"
	ntsync "nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

type BackGroundManager struct {
	db     *db.ControlDB
	config *config.ServerConfig
	cron   *cron.ACron
	queue  *liteq.JobQueue

	consumers []*liteq.ConsumerSupervisor

	BackupSiteJobManager    *backup_sites.BackupSiteJobManager
	BackupControlJobManager *backup_control.BackupControlJobManager
	SummariseJobManager     *summarise_jobs.TruncateJobManager
	SyncJobManager          *ntsync.SyncJobManager
}

func NewBackGroundManager(ctx context.Context, controlDB *db.ControlDB, config *config.ServerConfig, s3c *s3utils.S3Client) *BackGroundManager {
	queue := liteq.New(controlDB)

	cron := cron.NewCron(controlDB.Jobs)

	consumers := []*liteq.ConsumerSupervisor{}
	backupControlJobManager := backup_control.SetupBackupControlJob(ctx, controlDB, config, s3c)
	backupSiteJobManager := backup_sites.SetupBackupSiteJob(ctx, controlDB, config, s3c)
	summariseManager := summarise_jobs.SetupTruncateJobManager(config, controlDB)
	syncjobManager := ntsync.SetupSyncJobManager(config, controlDB)

	// Backup Control Jobs
	backupControlJobManager.SetupCron(cron, queue)
	// Backup Site Jobs
	consumers = append(consumers, backupControlJobManager.GetConsumer(queue))

	// Backup Site Jobs
	backupSiteJobManager.SetupCron(cron, queue)
	// Backup Should be done by exactly one worker at a time, to avoid over working the CPU and IO
	consumers = append(consumers, backupSiteJobManager.GetConsumer(queue))

	// Summarise Jobs
	summariseManager.SetupCron(cron)

	// Sync Jobs
	syncjobManager.SetupCron(cron, queue)
	// Sync jobs can be done in parallel by multiple workers
	consumers = append(consumers, syncjobManager.GetConsumer(queue))

	return &BackGroundManager{
		db:                      controlDB,
		config:                  config,
		cron:                    cron,
		queue:                   queue,
		consumers:               consumers,
		BackupSiteJobManager:    backupSiteJobManager,
		BackupControlJobManager: backupControlJobManager,
		SummariseJobManager:     summariseManager,
		SyncJobManager:          syncjobManager,
	}
}

func WrapperFunc(controlDB *db.ControlDB) func(name string, c func()) {
	return func(name string, c func()) {
		// TODO allow and catch errors
		// TODO catch pancis
		fmt.Println("Executing wrapped function for", name)
		c()
		fmt.Println("Finished executing wrapped function for", name)
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
