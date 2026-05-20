package sync

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
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
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
	UniqHash         string
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

		b.SyncUserData(&syncJob)
		return nil
	}
}

func (b *SyncJobManager) SetupCron(c *cron.Cron, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.SyncIntervalInSeconds)
	c.AddFunc(syncCron, b.buildSyncCronFunc(queue))
}

func (b *SyncJobManager) buildSyncCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		syncJobs := b.buildSyncJobs()
		for _, syncJob := range syncJobs {
			jsonRaw, err := json.Marshal(syncJob)
			if err != nil {
				slog.Error("Error marshalling sync job params", "error", err, "syncJob", syncJob)
				continue
			}
			err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
				Queue:       SYNC_JOB,
				Job:         string(jsonRaw),
				DedupingKey: liteq.ReplaceDuplicate(syncJob.UniqHash),
			})
			if err != nil {
				slog.Error("Error queueing sync job", "error", err)
			}
		}

	}
}

func (b *SyncJobManager) buildSyncJobs() []*SyncJobParams {
	defer utils.TrackTime(time.Now(), "SyncJobManager.Tick", nil, nil)
	users := b.controlDB.AllUsers()

	nsUniqHashMap := make(map[string]*SyncJobParams)

	for _, userDB := range users {
		ns, dex, err := userDB.GetAllSyncSources(context.TODO())
		if err != nil {
			slog.Error("error getting sync sources for user", "subdomain", userDB.Subdomain, "error", err)
			continue
		}
		err = b.addSyncJobToHashMap(userDB, ns, nsUniqHashMap)
		if err != nil {
			slog.Error("error adding Nightscout sync job to hash map", "subdomain", userDB.Subdomain, "error", err, "ns", ns)
		}

		err = b.addSyncJobToHashMap(userDB, dex, nsUniqHashMap)
		if err != nil {
			slog.Error("error adding Dexcom sync job to hash map", "subdomain", userDB.Subdomain, "error", err, "dexcom", dex)
		}
	}

	syncJobs := make([]*SyncJobParams, 0, len(nsUniqHashMap))
	for _, syncJob := range nsUniqHashMap {
		syncJobs = append(syncJobs, syncJob)
	}

	return syncJobs

}

func (b *SyncJobManager) addSyncJobToHashMap(userDB *db.UserDB, syncSource *user.SyncSource, nsUniqHashMap map[string]*SyncJobParams) error {
	if syncSource == nil || !syncSource.Enabled {
		return nil
	}

	_, _, uniqHash, err := syncSource.GetRawData(b.security)
	if err != nil {
		return fmt.Errorf("error getting raw data for sync source: %w", err)
	}

	jobData, exists := nsUniqHashMap[uniqHash]
	if !exists {
		nsUniqHashMap[uniqHash] = &SyncJobParams{
			UniqHash:         uniqHash,
			UserSubdomains:   []string{userDB.Subdomain},
			Synctype:         string(syncSource.Type),
			EncUrlOrUserName: syncSource.UrlOrUsername,
			EncPassword:      syncSource.Password,
			DexcomRegion:     string(syncSource.DexcomRegion),
		}
	} else {
		jobData.UserSubdomains = append(jobData.UserSubdomains, userDB.Subdomain)
	}

	return nil
}
