package background

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/background/backup_control"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/background/check_alarms"
	"nighttroop.t1d.nz/background/check_resources"
	"nighttroop.t1d.nz/background/summary_daily"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/background/summarise_jobs"
	ntsync "nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/background/trim_logs"
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
	backupSiteJobManager.SetupCron(cron)

	// Summarise Jobs
	summariseManager.SetupCron(cron)

	// Trim the logs DB's tables to their newest rows
	trim_logs.SetupCron(cron, config, controlDB.Logs)

	// Check each identity's alarms on new readings, and periodically
	check_alarms.SetupCron(cron, config, controlDB)

	// Watch the server's CPU, memory and disk, warning the control site when one gets high
	check_resources.SetupCron(cron, config, controlDB)

	// Build each site's yesterday soon after its midnight, and its last week again when its timezone changes
	summary_daily.SetupCron(cron, controlDB)
	consumers = append(consumers, onevent.Listen(config.Events, queue, summary_daily.RebuildOnSettingsChange(controlDB))...)

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

func (b *BackGroundManager) Start() {

	for _, consumer := range b.consumers {
		consumer.Start()
	}

	b.cron.Start()
}

// RunAllNow runs every cron job once, then every queued job, and returns when they are done.
// Tests call it after each request so background work, like syncing, is visible straight away.
func (b *BackGroundManager) RunAllNow(ctx context.Context) {
	for _, entry := range b.cron.Entries() {
		entry.Job.Run()
	}
	for _, consumer := range b.consumers {
		if err := consumer.RunQueued(ctx); err != nil {
			slog.Error("RunAllNow: running queued jobs failed", "error", err)
		}
	}
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
