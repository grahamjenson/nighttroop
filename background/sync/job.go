package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/security"
)

const SYNC_JOB = "sync_job"

type SyncJobManager struct {
	controlDB             *db.ControlDB
	security              *security.SecurityConfig
	SyncIntervalInSeconds int
	BackgroundSyncWorkers int
}

type SyncJobParams struct {
	UserSubdomains   []string
	UniqueHash       string
	EncUrlOrUserName string
	EncPassword      string
	DexcomRegion     string
	Synctype         string
}

func SetupSyncJobManager(config config.ServerConfig, controlDB *db.ControlDB) *SyncJobManager {

	jobs := &SyncJobManager{
		controlDB:             controlDB,
		security:              config.Security,
		SyncIntervalInSeconds: config.SyncIntervalInSeconds,
		BackgroundSyncWorkers: config.BackgroundSyncWorkers,
	}

	return jobs
}

func (b *SyncJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, SYNC_JOB, b.BackgroundSyncWorkers, b.workerFunc())
}

func (b *SyncJobManager) workerFunc() func(ctx context.Context, j *control.Job) error {

	return func(ctx context.Context, j *control.Job) error {
		var syncJob SyncJobParams
		err := json.Unmarshal([]byte(j.Job), &syncJob)
		if err != nil {
			slog.Error("Error unmarshalling sync job params", "error", err, "job", j.Job)
			return nil
		}

		return b.SyncUserData(&syncJob)
	}
}

func (b *SyncJobManager) SetupCron(c *cron.Cron, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.SyncIntervalInSeconds)
	c.AddFunc(syncCron, b.buildSyncCronFunc(queue))
}

func (b *SyncJobManager) buildSyncCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		// 1. Create All Sync Reports for Sync Sources that need them
		// 2. Get All Sync Reports that belong to at lease 1 enabled sync source, and whose next run is after the last attempt
		// 3. Create a SyncJobParams for each report
		err := b.controlDB.Queries.CreateSyncReportsForAllSyncSources(context.TODO())
		if err != nil {
			slog.Error("error creating sync reports for all sync sources", "error", err)
			return
		}
		syncReports, err := b.controlDB.Queries.GetSyncReportsToRun(context.TODO())
		if err != nil {
			slog.Error("error getting sync reports to run", "error", err)
			return
		}

		syncJobs := make([]*SyncJobParams, 0, len(syncReports))
		for _, report := range syncReports {
			syncJobs = append(syncJobs, &SyncJobParams{
				UniqueHash:       report.UniqueHash,
				UserSubdomains:   strings.Split(report.Subdomains, ","),
				Synctype:         report.Type,
				EncUrlOrUserName: report.UrlOrUsername,
				EncPassword:      report.Password,
				DexcomRegion:     report.DexcomRegion,
			})
		}

		for _, syncJob := range syncJobs {
			jsonRaw, err := json.Marshal(syncJob)
			if err != nil {
				slog.Error("Error marshalling sync job params", "error", err, "syncJob", syncJob)
				continue
			}
			err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
				Queue:       SYNC_JOB,
				Job:         string(jsonRaw),
				DedupingKey: liteq.ReplaceDuplicate(syncJob.UniqueHash),
			})
			if err != nil {
				slog.Error("Error queueing sync job", "error", err)
			}
		}

	}
}
