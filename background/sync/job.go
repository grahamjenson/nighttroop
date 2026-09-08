package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils/acron"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
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
	UniqueHash string
	Type       string
}

type SyncJobInfo struct {
	UniqueHash      string
	Type            string
	userSyncSources []*SyncUserInfo
	urlOrUsername   string
	password        string
	region          consts.DexcomRegion
}

type SyncUserInfo struct {
	siteModel    *control.Site
	siteDB       *db.SiteDB
	controlDB    *db.ControlDB
	syncIdentity *control.Identity
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

// There is no filtering at this level,
// Once a job has been scheduled, it will be run.
func (b *SyncJobManager) workerFunc() func(ctx context.Context, j *jobs.Job) error {

	return func(ctx context.Context, j *jobs.Job) error {
		var syncJob SyncJobParams
		err := json.Unmarshal([]byte(j.Job), &syncJob)
		if err != nil {
			slog.Error("Error unmarshalling sync job params", "error", err, "job", j.Job)
			return nil
		}
		// This will filter out
		syncSources, err := b.controlDB.Queries.GetAllSiteSyncSourcesForUniqueHash(ctx, control.GetAllSiteSyncSourcesForUniqueHashParams{
			UniqueHash: syncJob.UniqueHash,
		})
		if err != nil {
			slog.Error("Error getting user sync sources", "error", err, "unique_hash", syncJob.UniqueHash)
			return err
		}

		if len(syncSources) == 0 {
			slog.Error("workerFunc: No user sync sources found for unique hash", "unique_hash", syncJob.UniqueHash, "sync_type", syncJob.Type)
			return nil
		}

		userInfoSyncSources := []*SyncUserInfo{}
		for _, syncSource := range syncSources {
			userInfo, err := b.AttachSyncSourceWithUserInfo(syncSource)
			if err != nil {
				slog.Error("Error attaching sync source with user info", "error", err, "unique_hash", syncJob.UniqueHash, "sync_source_id", syncSource.ID)
				continue
			}

			if userInfo == nil {
				slog.Error("No user info found for sync source", "unique_hash", syncJob.UniqueHash, "sync_source_id", syncSource.ID)
				continue
			}
			userInfoSyncSources = append(userInfoSyncSources, userInfo)
		}

		if len(userInfoSyncSources) == 0 {
			slog.Error("No user info sync sources found for unique hash", "unique_hash", syncJob.UniqueHash)
			return nil
		}

		username, password, region, err := b.getUsernamePasswordRegion(syncSources[0])
		if err != nil {
			slog.Error("Error getting username, password, and region", "error", err, "unique_hash", syncJob.UniqueHash)
			return nil
		}

		syncInfo := &SyncJobInfo{
			UniqueHash: syncJob.UniqueHash,
			Type:       syncJob.Type,

			userSyncSources: userInfoSyncSources,
			urlOrUsername:   username,
			password:        password,
			region:          region,
		}
		return b.SyncUserData(syncInfo)
	}
}

func (b *SyncJobManager) getUsernamePasswordRegion(syncSource *control.SiteSyncSource) (string, string, consts.DexcomRegion, error) {

	encUrlOrUserName := syncSource.UrlOrUsername
	encPassword := syncSource.Password
	dexcomRegion := syncSource.DexcomRegion

	unencUsername, err := b.security.Decrypt(encUrlOrUserName)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting username: %w", err)
	}
	unencPassword, err := b.security.Decrypt(encPassword)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting password: %w", err)
	}
	return unencUsername, unencPassword, dexcomRegion, nil
}

func (b *SyncJobManager) AttachSyncSourceWithUserInfo(syncSource *control.SiteSyncSource) (*SyncUserInfo, error) {
	userDb, userModel, err := b.controlDB.GetSite(syncSource.Subdomain)
	if err != nil {
		return nil, err
	}
	if userModel == nil || userDb == nil {
		return nil, nil
	}
	identityName := ""
	switch syncSource.Type {
	case consts.SyncSourceTypeNightscout:
		identityName = consts.NIGHTSCOUT_SYNC_IDENTITY_NAME
	case consts.SyncSourceTypeDexcom:
		identityName = consts.DEXCOM_SYNC_IDENTITY_NAME
	}

	if identityName == "" {
		return nil, fmt.Errorf("invalid sync source type: %s", syncSource.Type)
	}
	identity, err := b.controlDB.Queries.GetIdentity(context.TODO(), control.GetIdentityParams{
		Subdomain: syncSource.Subdomain,
		Name:      identityName,
	})

	if err != nil {
		return nil, err
	}

	return &SyncUserInfo{
		siteDB:       userDb,
		siteModel:    userModel,
		syncIdentity: identity,
		controlDB:    b.controlDB,
	}, nil
}

func (b *SyncJobManager) SetupCron(c acron.CronScheduler, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.SyncIntervalInSeconds)

	c.AddFunc("SyncJobManager", syncCron, b.buildSyncCronFunc(queue))

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

		// This gets reports that are:
		// 1. Have a user with an Enabled sync_source
		// 2. Are due to run because last attempt is null or was more than the sync interval ago
		// 3. The User has been seen in the last day
		// 4. The User is not deleted
		syncReports, err := b.controlDB.Queries.GetSyncReportsToRun(context.TODO())
		if err != nil {
			slog.Error("error getting sync reports to run", "error", err)
			return
		}

		for _, report := range syncReports {

			syncJob := &SyncJobParams{
				UniqueHash: report.UniqueHash,
				Type:       report.Type,
			}

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
