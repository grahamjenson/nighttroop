package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/server/nightscoutapi"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

const SYNC_JOB = "sync_job"

type SyncJobManager struct {
	controlDB *db.ControlDB
	security  *security.SecurityConfig

	SyncIntervalInSeconds int
	BackgroundSyncWorkers int

	// For Nightscout API
	nightscoutStore *nightscoutapi.NightTroopAPIStore
}

type SyncJobParams struct {
	SiteID string
	Type   consts.SyncSourceType
}

type SyncJobInfo struct {
	SiteID          string
	Type            consts.SyncSourceType
	UserSyncSources []*SyncUserInfo
	urlOrUsername   string
	password        string
	region          consts.DexcomRegion
	// firstSync is true until a sync has worked since the source was saved; a failure then disables it
	firstSync bool
}

type SyncUserInfo struct {
	SiteModel *controldb.Site
	SiteDB    *db.SiteDB
	ControlDB *db.ControlDB
}

func SetupSyncJobManager(config *config.ServerConfig, controlDB *db.ControlDB) *SyncJobManager {

	jobs := &SyncJobManager{
		controlDB:             controlDB,
		security:              config.Security,
		SyncIntervalInSeconds: config.SyncIntervalInSeconds,
		BackgroundSyncWorkers: config.BackgroundSyncWorkers,
		nightscoutStore:       nightscoutapi.NewNightTroopAPIStore(controlDB, config),
	}

	return jobs
}

func (b *SyncJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, SYNC_JOB, b.BackgroundSyncWorkers, b.workerFunc())
}

// There is no filtering at this level,
// Once a job has been scheduled, it will be run.
func (b *SyncJobManager) workerFunc() func(ctx context.Context, j *jobsdb.Job) error {

	return func(ctx context.Context, j *jobsdb.Job) error {
		var syncJob SyncJobParams
		err := json.Unmarshal([]byte(j.Job), &syncJob)
		if err != nil {
			slog.Error("Error unmarshalling sync job params", "error", err, "job", j.Job)
			return nil
		}
		return b.SyncSource(ctx, syncJob)
	}
}

// SyncSource runs one sync of a site's sync source now, recording the result on the source
func (b *SyncJobManager) SyncSource(ctx context.Context, syncJob SyncJobParams) error {
	syncSource, err := b.controlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{
		SiteID: syncJob.SiteID,
		Type:   syncJob.Type,
	})
	if err != nil {
		slog.Error("Error getting user sync sources", "error", err, "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
		return err
	}
	if syncSource == nil {
		slog.Error("workerFunc: No user sync sources found for site", "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
		return nil
	}

	userInfoSyncSources := []*SyncUserInfo{}
	for _, syncSource := range []*controldb.SiteSyncSource{syncSource} {
		userInfo, err := b.AttachSyncSourceWithUserInfo(syncSource)
		if err != nil {
			slog.Error("Error attaching sync source with user info", "error", err, "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
			continue
		}

		if userInfo == nil {
			slog.Error("No user info found for sync source", "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
			continue
		}
		userInfoSyncSources = append(userInfoSyncSources, userInfo)
	}

	if len(userInfoSyncSources) == 0 {
		slog.Error("No user info sync sources found for unique hash", "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
		return nil
	}

	username, password, region, err := b.getUsernamePasswordRegion(syncSource)
	if err != nil {
		slog.Error("Error getting username, password, and region", "error", err, "site_id", syncJob.SiteID, "sync_type", syncJob.Type)
		return nil
	}

	syncInfo := &SyncJobInfo{
		SiteID:          syncJob.SiteID,
		Type:            syncJob.Type,
		UserSyncSources: userInfoSyncSources,
		urlOrUsername:   username,
		password:        password,
		region:          region,
		firstSync:       syncSource.FirstSuccessfulSyncTime == nil,
	}
	err = b.SyncUserData(syncInfo)
	if err != nil {
		b.controlDB.Logs.Error(ctx, &syncJob.SiteID, err, map[string]any{"sync_type": syncJob.Type})
	}
	return err
}

func (b *SyncJobManager) getUsernamePasswordRegion(syncSource *controldb.SiteSyncSource) (string, string, consts.DexcomRegion, error) {
	if syncSource == nil {
		return "", "", consts.DexcomRegionOther, fmt.Errorf("sync source is nil")
	}
	if syncSource.Type == consts.SyncSourceTypeVirtual {
		return "virtual_username", "virtual_password", consts.DexcomRegionOther, nil
	}

	encUrlOrUserName := syncSource.UrlOrUsernameEnc
	encPassword := syncSource.PasswordEnc
	dexcomRegion := consts.DexcomRegionOther
	if encUrlOrUserName == nil || encPassword == nil {
		return "", "", consts.DexcomRegionOther, fmt.Errorf("encrypted username or password is nil")
	}
	if syncSource.Region != nil {
		dexcomRegion = consts.NewDexcomRegionFromString(*syncSource.Region)
	}

	unencUsername, err := b.security.Decrypt(*encUrlOrUserName)
	if err != nil {
		return "", "", consts.DexcomRegionOther, fmt.Errorf("error decrypting username: %w", err)
	}
	unencPassword, err := b.security.Decrypt(*encPassword)
	if err != nil {
		return "", "", consts.DexcomRegionOther, fmt.Errorf("error decrypting password: %w", err)
	}
	return unencUsername, unencPassword, dexcomRegion, nil
}

func (b *SyncJobManager) AttachSyncSourceWithUserInfo(syncSource *controldb.SiteSyncSource) (*SyncUserInfo, error) {
	userDb, userModel, err := b.controlDB.GetSite(syncSource.SiteID)
	if err != nil {
		slog.Error("Cannot find the site for sync source", "site_id", syncSource.SiteID, "error", err)
		return nil, err
	}
	if userModel == nil || userDb == nil {
		return nil, nil
	}

	return &SyncUserInfo{
		SiteDB:    userDb,
		SiteModel: userModel,
		ControlDB: b.controlDB,
	}, nil
}

func (b *SyncJobManager) SetupCron(c cron.CronScheduler, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.SyncIntervalInSeconds)

	c.AddFunc("SyncJobManager", syncCron, b.buildSyncCronFunc(queue))

}

func (b *SyncJobManager) buildSyncCronFunc(queue *liteq.JobQueue) func() error {
	return func() error {
		// This gets reports that are:
		// 1. Have a user with an Enabled sync_source
		// 2. Are due to run because last attempt is null or was more than the sync interval ago
		// 3. The User has been seen in the last day
		// 4. The User is not deleted
		syncSources, err := b.controlDB.Queries.GetSyncReportsToRun(context.TODO())
		if err != nil {
			slog.Error("error getting sync reports to run", "error", err)
			return err
		}

		for _, source := range syncSources {

			syncJob := &SyncJobParams{
				SiteID: source.SiteID,
				Type:   source.Type,
			}

			jsonRaw, err := json.Marshal(syncJob)
			if err != nil {
				slog.Error("Error marshalling sync job params", "error", err, "syncJob", syncJob)
				continue
			}
			err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
				Queue:       SYNC_JOB,
				Job:         string(jsonRaw),
				DedupingKey: liteq.ReplaceDuplicate(fmt.Sprintf("%s-%s", syncJob.SiteID, syncJob.Type)),
			})
			if err != nil {
				slog.Error("Error queueing sync job", "error", err)
			}
		}

		return nil
	}
}
