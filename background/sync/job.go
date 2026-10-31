package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nighttroop.t1d.nz/data/events"
	gosync "sync"

	lru "github.com/hashicorp/golang-lru/v2"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/server/nightscoutapi"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

const SYNC_JOB = "sync_job"

type SyncJobManager struct {
	controlDB *db.ControlDB
	security  *security.SecurityConfig
	// config is the server's, whose events a sync's writes are published to
	config *config.ServerConfig

	SyncIntervalInSeconds int
	BackgroundSyncWorkers int

	// For Nightscout API
	nightscoutStore *nightscoutapi.NightTroopAPIStore

	// Each site's Nightscout client, kept between syncs, the least recently used dropped once there are too many
	nightscoutClientsLock gosync.Mutex
	nightscoutClients     *lru.Cache[string, *nightscoutClient]

	// Each site's LibreLinkUp client, kept between syncs like the Nightscout clients
	libreClientsLock gosync.Mutex
	libreClients     *lru.Cache[string, *libreClient]
}

// maxNightscoutClients is how many sites' Nightscout clients are kept, about 1KB each, so memory stays bounded however
// many sites sync; a site whose client is dropped logs in again, and reads every collection, on its next sync
const maxNightscoutClients = 10_000

type SyncJobParams struct {
	SiteID string
	Type   consts.SyncSourceType
	// Backfill is set for a backfill rather than a sync
	Backfill *BackfillParams `json:",omitempty"`
}

type SyncJobInfo struct {
	SiteID string
	Type   consts.SyncSourceType
	// Site is the site the source syncs into
	Site          *SyncSite
	urlOrUsername string
	password      string
	// region is as stored: a Dexcom region, or one of consts.LibreRegions
	region string
	// firstSync is true until a sync has worked since the source was saved; a failure then disables it
	firstSync bool
	// ctx logs the job, the site and the sync's type. It is never cancelled, as a first sync stopped part way would be
	// a failure that turns syncing off.
	ctx context.Context
}

// SyncSite is the site a sync writes its records into
type SyncSite struct {
	SiteModel *controldb.Site
	SiteDB    *db.SiteDB
	ControlDB *db.ControlDB
	// Config is the server's, so a sync's writes are published to the server's events, as the site's own are
	Config *config.ServerConfig
}

// published sends the model change of a row a sync created directly, as a write through the site would
func (s *SyncSite) published(row any) {
	if s.Config != nil && s.Config.Events != nil {
		s.Config.Events.Publish(s.SiteModel.ID, events.ModelChange(s.SiteModel.ID, events.Created, row))
	}
}

// siteContext is what a sync writes to the site through the Nightscout collections as, the identity naming the sync
func (s *SyncSite) siteContext(identityName string) *httputils.SiteContext {
	return &httputils.SiteContext{
		SiteID:        s.SiteModel.ID,
		IdentityModel: &controldb.SiteIdentity{Name: identityName},
		ControlDB:     s.ControlDB,
		SiteDB:        s.SiteDB,
		SiteModel:     s.SiteModel,
		HTTPAction:    httputils.HTTPActionWrite,
		Config:        s.Config,
	}
}

func SetupSyncJobManager(config *config.ServerConfig, controlDB *db.ControlDB) *SyncJobManager {
	// It only fails for a size under 1
	nightscoutClients, _ := lru.New[string, *nightscoutClient](maxNightscoutClients)
	libreClients, _ := lru.New[string, *libreClient](maxNightscoutClients)

	jobs := &SyncJobManager{
		controlDB:             controlDB,
		security:              config.Security,
		config:                config,
		SyncIntervalInSeconds: config.SyncIntervalInSeconds,
		BackgroundSyncWorkers: config.BackgroundSyncWorkers,
		nightscoutStore:       nightscoutapi.NewNightTroopAPIStore(controlDB, config),
		nightscoutClients:     nightscoutClients,
		libreClients:          libreClients,
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
			slog.ErrorContext(ctx, "Error unmarshalling sync job params", "error", err, "params", j.Job)
			return nil
		}
		return b.SyncSource(ctx, syncJob)
	}
}

// SyncSource runs one sync of a site's sync source now, recording the result on the source
func (b *SyncJobManager) SyncSource(ctx context.Context, syncJob SyncJobParams) error {
	ctx = aslog.With(aslog.WithSite(ctx, syncJob.SiteID), slog.String("sync_type", string(syncJob.Type)))
	syncSource, err := b.controlDB.Queries.GetSyncSource(ctx, controldb.GetSyncSourceParams{
		SiteID: syncJob.SiteID,
		Type:   syncJob.Type,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user sync sources", "error", err)
		return err
	}
	if syncSource == nil {
		slog.ErrorContext(ctx, "workerFunc: No user sync sources found for site")
		return nil
	}

	site, err := b.syncSite(ctx, syncSource.SiteID)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting the site for sync source", "error", err)
		return nil
	}
	if site == nil {
		slog.ErrorContext(ctx, "No site found for sync source")
		return nil
	}

	username, password, region, err := b.getUsernamePasswordRegion(syncSource)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting username, password, and region", "error", err)
		b.controlDB.Logs.SyncError(ctx, syncJob.SiteID, syncJob.Type, err)
		return nil
	}

	syncInfo := &SyncJobInfo{
		SiteID:        syncJob.SiteID,
		Type:          syncJob.Type,
		Site:          site,
		urlOrUsername: username,
		password:      password,
		region:        region,
		firstSync:     syncSource.FirstSuccessfulSyncTime == nil,
		ctx:           context.WithoutCancel(ctx),
	}
	if syncJob.Backfill != nil {
		return b.backfill(syncInfo, *syncJob.Backfill)
	}
	return b.SyncUserData(syncInfo)
}

func (b *SyncJobManager) getUsernamePasswordRegion(syncSource *controldb.SiteSyncSource) (string, string, string, error) {
	if syncSource == nil {
		return "", "", "", fmt.Errorf("sync source is nil")
	}
	if syncSource.Type == consts.SyncSourceTypeVirtual {
		return "virtual_username", "virtual_password", "", nil
	}

	encUrlOrUserName := syncSource.UrlOrUsernameEnc
	encPassword := syncSource.PasswordEnc
	region := ""
	if encUrlOrUserName == nil || encPassword == nil {
		return "", "", "", fmt.Errorf("encrypted username or password is nil")
	}
	if syncSource.Region != nil {
		region = *syncSource.Region
	}

	unencUsername, err := b.security.Decrypt(*encUrlOrUserName)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting username: %w", err)
	}
	unencPassword, err := b.security.Decrypt(*encPassword)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting password: %w", err)
	}
	return unencUsername, unencPassword, region, nil
}

// syncSite is the site a sync source writes into, or nil when it has gone
func (b *SyncJobManager) syncSite(ctx context.Context, siteID string) (*SyncSite, error) {
	siteDB, siteModel, err := b.controlDB.GetSite(siteID)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot find the site for sync source", "error", err)
		return nil, err
	}
	if siteModel == nil || siteDB == nil {
		return nil, nil
	}

	return &SyncSite{
		SiteDB:    siteDB,
		SiteModel: siteModel,
		ControlDB: b.controlDB,
		Config:    b.config,
	}, nil
}

func (b *SyncJobManager) SetupCron(c cron.CronScheduler, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.SyncIntervalInSeconds)

	c.AddFunc("SyncJobManager", syncCron, b.buildSyncCronFunc(queue))

}

func (b *SyncJobManager) buildSyncCronFunc(queue *liteq.JobQueue) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// This gets reports that are:
		// 1. Have a user with an Enabled sync_source
		// 2. Are due to run because last attempt is null or was more than the sync interval ago
		// 3. The User has been seen in the last day
		// 4. The User is not deleted
		syncSources, err := b.controlDB.Queries.GetSyncReportsToRun(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error getting sync reports to run", "error", err)
			return err
		}

		for _, source := range syncSources {

			syncJob := &SyncJobParams{
				SiteID: source.SiteID,
				Type:   source.Type,
			}

			jsonRaw, err := json.Marshal(syncJob)
			if err != nil {
				slog.ErrorContext(ctx, "Error marshalling sync job params", "error", err, "syncJob", syncJob)
				continue
			}
			err = queue.QueueJob(ctx, liteq.QueueJobParams{
				Queue:       SYNC_JOB,
				Job:         string(jsonRaw),
				DedupingKey: liteq.ReplaceDuplicate(fmt.Sprintf("%s-%s", syncJob.SiteID, syncJob.Type)),
			})
			if err != nil {
				slog.ErrorContext(ctx, "Error queueing sync job", "error", err)
			}
		}

		return nil
	}
}
