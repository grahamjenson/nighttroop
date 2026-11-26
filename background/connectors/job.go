package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nighttroop.t1d.nz/data/events"
	"sync"
	"time"

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

const CONNECTOR_JOB = "connector_job"

type ConnectorJobManager struct {
	controlDB *db.ControlDB
	security  *security.SecurityConfig
	// config is the server's, whose events a sync's writes are published to
	config *config.ServerConfig

	ConnectorIntervalInSeconds int
	BackgroundConnectorWorkers int

	// For Nightscout API
	nightscoutStore *nightscoutapi.NightTroopAPIStore

	// Each site's Nightscout client, kept between syncs, the least recently used dropped once there are too many
	nightscoutClientsLock sync.Mutex
	nightscoutClients     *lru.Cache[string, *nightscoutClient]

	// Each site's LibreLinkUp client, kept between syncs like the Nightscout clients
	libreClientsLock sync.Mutex
	libreClients     *lru.Cache[string, *libreClient]
}

// maxNightscoutClients is how many sites' Nightscout clients are kept, about 1KB each, so memory stays bounded however
// many sites sync; a site whose client is dropped logs in again, and reads every collection, on its next sync
const maxNightscoutClients = 10_000

type ConnectorJobParams struct {
	SiteID string
	Type   consts.ConnectorType
	// Backfill is set for a backfill rather than a sync
	Backfill *BackfillParams `json:",omitempty"`
	// Timed is set for a sync as a Nightscout entry is due, on top of the usual schedule
	Timed bool `json:",omitempty"`
}

type ConnectorJobInfo struct {
	SiteID string
	Type   consts.ConnectorType
	// Site is the site the source syncs into
	Site          *ConnectorSite
	urlOrUsername string
	password      string
	// region is as stored: a Dexcom region, or one of consts.LibreRegions
	region string
	// firstSync is true until a sync has worked since the source was saved; a failure then disables it
	firstSync bool
	// timed is a sync as a Nightscout entry is due, on top of the usual schedule
	timed bool
	// ctx logs the job, the site and the sync's type. It is never cancelled, as a first sync stopped part way would be
	// a failure that turns syncing off.
	ctx context.Context
}

// ConnectorSite is the site a sync writes its records into
type ConnectorSite struct {
	SiteModel *controldb.Site
	SiteDB    *db.SiteDB
	ControlDB *db.ControlDB
	// Config is the server's, so a sync's writes are published to the server's events, as the site's own are
	Config *config.ServerConfig
}

// published sends the model change of a row a sync created directly, as a write through the site would
func (s *ConnectorSite) published(row any) {
	if s.Config != nil && s.Config.Events != nil {
		s.Config.Events.Publish(s.SiteModel.ID, events.ModelChange(s.SiteModel.ID, events.Created, row))
	}
}

// siteContext is what a sync writes to the site through the Nightscout collections as, the identity naming the sync
func (s *ConnectorSite) siteContext(identityName string) *httputils.SiteContext {
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

func SetupConnectorJobManager(config *config.ServerConfig, controlDB *db.ControlDB) *ConnectorJobManager {
	// It only fails for a size under 1
	nightscoutClients, _ := lru.New[string, *nightscoutClient](maxNightscoutClients)
	libreClients, _ := lru.New[string, *libreClient](maxNightscoutClients)

	jobs := &ConnectorJobManager{
		controlDB:                  controlDB,
		security:                   config.Security,
		config:                     config,
		ConnectorIntervalInSeconds: config.ConnectorIntervalInSeconds,
		BackgroundConnectorWorkers: config.BackgroundConnectorWorkers,
		nightscoutStore:            nightscoutapi.NewNightTroopAPIStore(controlDB, config),
		nightscoutClients:          nightscoutClients,
		libreClients:               libreClients,
	}

	return jobs
}

func (b *ConnectorJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, CONNECTOR_JOB, b.BackgroundConnectorWorkers, b.workerFunc())
}

// There is no filtering at this level,
// Once a job has been scheduled, it will be run.
func (b *ConnectorJobManager) workerFunc() func(ctx context.Context, j *jobsdb.Job) error {

	return func(ctx context.Context, j *jobsdb.Job) error {
		var connectorJob ConnectorJobParams
		err := json.Unmarshal([]byte(j.Job), &connectorJob)
		if err != nil {
			slog.ErrorContext(ctx, "Error unmarshalling connector job params", "error", err, "params", j.Job)
			return nil
		}
		return b.RunConnector(ctx, connectorJob)
	}
}

// RunConnector runs one sync of a site's connector now, recording the result on the source
func (b *ConnectorJobManager) RunConnector(ctx context.Context, connectorJob ConnectorJobParams) error {
	ctx = aslog.With(aslog.WithSite(ctx, connectorJob.SiteID), slog.String("connector_type", string(connectorJob.Type)))
	connector, err := b.controlDB.Queries.GetConnector(ctx, controldb.GetConnectorParams{
		SiteID: connectorJob.SiteID,
		Type:   connectorJob.Type,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user connectors", "error", err)
		return err
	}
	if connector == nil {
		slog.ErrorContext(ctx, "workerFunc: No user connectors found for site")
		return nil
	}
	if connector.Type == consts.ConnectorTypeVirtual && !b.config.VirtualConnectorAllowed() {
		slog.WarnContext(ctx, "The virtual source is not synced on a production server")
		return nil
	}
	// A timed sync is only on top of a working connector's usual syncs
	if connectorJob.Timed && (!connector.Enabled || connector.ConsecutiveFailedSyncs != nil && *connector.ConsecutiveFailedSyncs > 0) {
		return nil
	}

	site, err := b.connectorSite(ctx, connector.SiteID)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting the site for connector", "error", err)
		return nil
	}
	if site == nil {
		slog.ErrorContext(ctx, "No site found for connector")
		return nil
	}

	// Nothing is synced into a site far over its storage, or while the server's disk is nearly full, as the site's own
	// writes are refused
	if refused := httputils.SiteWriteRefused(&httputils.SiteContext{ControlDB: b.controlDB, SiteModel: site.SiteModel, SiteDB: site.SiteDB}); refused != "" {
		slog.WarnContext(ctx, "A connector is not synced, as its site's writes are refused", "why", refused)
		return nil
	}

	username, password, region, err := b.getUsernamePasswordRegion(connector)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting username, password, and region", "error", err)
		b.controlDB.Logs.ConnectorError(ctx, connectorJob.SiteID, connectorJob.Type, err)
		return nil
	}

	// No sync goes over its limit, so a Nightscout many sites sync from, or a fault in scheduling, is never synced too
	// often. One refused records nothing, so the next run, every few seconds, tries again. A backfill is not limited here:
	// its site's backfill limit was spent when it was asked for (SiteBackfillLimiter).
	if connectorJob.Backfill == nil && !b.config.RateLimiters.ConnectorSyncLimiter(syncLimitKey(connector, username)).Allow() {
		slog.DebugContext(ctx, "A connector sync is over its rate limit, so waits for a later run")
		return nil
	}

	jobInfo := &ConnectorJobInfo{
		SiteID:        connectorJob.SiteID,
		Type:          connectorJob.Type,
		Site:          site,
		urlOrUsername: username,
		password:      password,
		region:        region,
		firstSync:     connector.FirstSuccessfulSyncTime == nil,
		timed:         connectorJob.Timed,
		ctx:           context.WithoutCancel(ctx),
	}
	if connectorJob.Backfill != nil {
		return b.backfill(jobInfo, *connectorJob.Backfill)
	}
	return b.SyncUserData(jobInfo)
}

// syncLimitKey is what a connector's syncs are limited by: its Nightscout's host, as many sites can sync from one, or
// else the connector itself
func syncLimitKey(connector *controldb.SiteConnector, urlOrUsername string) string {
	if connector.Type == consts.ConnectorTypeNightscout {
		return "nightscout " + nightscoutHost(urlOrUsername)
	}
	return string(connector.Type) + " " + connector.SiteID
}

func (b *ConnectorJobManager) getUsernamePasswordRegion(connector *controldb.SiteConnector) (string, string, string, error) {
	if connector == nil {
		return "", "", "", fmt.Errorf("connector is nil")
	}
	if connector.Type == consts.ConnectorTypeVirtual {
		return "virtual_username", "virtual_password", "", nil
	}

	encUrlOrUserName := connector.UrlOrUsernameEnc
	encPassword := connector.PasswordEnc
	region := ""
	if encUrlOrUserName == nil || encPassword == nil {
		return "", "", "", fmt.Errorf("encrypted username or password is nil")
	}
	if connector.Region != nil {
		region = *connector.Region
	}

	unencUsername, err := b.security.DecryptFor(security.PurposeConnector, connector.SiteID, *encUrlOrUserName)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting username: %w", err)
	}
	unencPassword, err := b.security.DecryptFor(security.PurposeConnector, connector.SiteID, *encPassword)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting password: %w", err)
	}
	return unencUsername, unencPassword, region, nil
}

// connectorSite is the site a connector writes into, or nil when it has gone
func (b *ConnectorJobManager) connectorSite(ctx context.Context, siteID string) (*ConnectorSite, error) {
	siteDB, siteModel, err := b.controlDB.GetSite(ctx, siteID)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot find the site for connector", "error", err)
		return nil, err
	}
	if siteModel == nil || siteDB == nil {
		return nil, nil
	}

	return &ConnectorSite{
		SiteDB:    siteDB,
		SiteModel: siteModel,
		ControlDB: b.controlDB,
		Config:    b.config,
	}, nil
}

func (b *ConnectorJobManager) SetupCron(c cron.CronScheduler, queue *liteq.JobQueue) {
	syncCron := fmt.Sprintf("@every %ds", b.ConnectorIntervalInSeconds)

	c.AddFunc("ConnectorJobManager", syncCron, b.buildSyncCronFunc(queue))

}

func (b *ConnectorJobManager) buildSyncCronFunc(queue *liteq.JobQueue) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// This gets reports that are:
		// 1. Have a user with an Enabled connector
		// 2. Are due to run because last attempt is null or was more than the sync interval ago
		// 3. The User has been seen in the last day
		// 4. The User is not deleted
		connectors, err := b.controlDB.Queries.GetConnectorsToRun(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error getting sync reports to run", "error", err)
			return err
		}

		for _, source := range connectors {
			if source.Type == consts.ConnectorTypeVirtual && !b.config.VirtualConnectorAllowed() {
				continue
			}
			queueSync(ctx, queue, ConnectorJobParams{SiteID: source.SiteID, Type: source.Type}, "")
		}
		// Each Nightscout connector whose next entry is due about now, by its own key so it does not replace a usual sync
		for _, siteID := range b.entriesDue(time.Now()) {
			queueSync(ctx, queue, ConnectorJobParams{SiteID: siteID, Type: consts.ConnectorTypeNightscout, Timed: true}, "-timed")
		}

		return nil
	}
}

// queueSync queues a connector's sync, replacing one of it queued with the same key suffix
func queueSync(ctx context.Context, queue *liteq.JobQueue, connectorJob ConnectorJobParams, keySuffix string) {
	jsonRaw, err := json.Marshal(connectorJob)
	if err != nil {
		slog.ErrorContext(ctx, "Error marshalling connector job params", "error", err, "connectorJob", connectorJob)
		return
	}
	err = queue.QueueJob(ctx, liteq.QueueJobParams{
		Queue:       CONNECTOR_JOB,
		Job:         string(jsonRaw),
		DedupingKey: liteq.ReplaceDuplicate(fmt.Sprintf("%s-%s%s", connectorJob.SiteID, connectorJob.Type, keySuffix)),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error queueing connector job", "error", err)
	}
}
