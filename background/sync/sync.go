package sync

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// This is a background job runner that loops over users and syncs some data (initially just entries) from Nighscout and/or Dexcom.
//
// The big problem is the user defines their own nightscout URL.
// If the user is malicious, they could intend to slow down Nighttroop or DDOS another site.
//
// They could slow down Nighttroop by:
// 1. having it point at Nighttroop (syncing to itself)
// 1. Pointing at a server that takes a long time to respond or never responds
// 1. Pointing to a server that sometimes responds and sometimes doesn't, causing timeouts and retries
//
// They could DDOS another site by:
// 1. pointing many users at the same site.
// 1. pointing at an expencive call on another site causing issues there
// 1. Pointing to different sumdomains of a DNS they control that points to the same server (e.g. a.example.com, b.example.com, c.example.com all point to the same server)
//
// To deal with these issues:
// 1. Cant point URL at Nighttroop (or any subdomain of Nighttroop)
// 2. If a users sync is taking too long or failing decrease the rate at which they sync (until they are disabled)
// 3. Only sync once from a source:password combination (prevents multiple users pointing at the same URL and causing a DDOS)
// 4. Agressive timeouts for jobs as well as logging
// 5. List of bad and blocked hostnames, that are never allowed to be synced from (TODO)
// 6. We only Sync users that have called a read endpoint recently.
// 7. We slow down syncing of users that don't actually get new data (count writes to the DB)

// Dexcom is easier, but if say the dexcom servers are down we might encounter lots of timeouts, which we should detect quickly and slow down all syncing.

type SyncJobs struct {
	controlDB *db.ControlDB
	config    config.ServerConfig
	ctx       context.Context
	cancel    func()
	wg        sync.WaitGroup
	interval  time.Duration

	jobCh chan *SyncJobWithData
}

type UserDBSyncData struct {
	UserDB     *db.UserDB
	SyncSource *user.SyncSource
}

type SyncJobWithData struct {
	uniqHash                 string // hash of the sync source config (type, URL/username, password, dexcom region) used for deduping
	users                    []*UserDBSyncData
	synctype                 consts.SyncSourceType
	unencryptedPassword      string
	unencryptedURLOrUsername string
	dexcomRegion             consts.DexcomRegion      // only used for dexcom, empty for nightscout
	report                   control.SyncSourceReport // the current report for this sync source, used to update after the sync with results like success/failure, time taken, etc.
	// Any other data we want to pass to the job, like the last time we successfully synced, or the number of consecutive failures, etc.
}

func (s *SyncJobWithData) Subdomains() []string {
	subdomains := make([]string, len(s.users))
	for i, user := range s.users {
		subdomains[i] = user.UserDB.Subdomain
	}
	return subdomains
}

// Background Jobs

func SetupSyncJobs(config config.ServerConfig, controlDB *db.ControlDB) *SyncJobs {

	jobs := &SyncJobs{
		controlDB: controlDB,
		config:    config,
		interval:  time.Duration(config.SyncIntervalInSeconds) * time.Second,
		jobCh:     make(chan *SyncJobWithData, config.BackgroundSyncWorkers*2), // Buffer size is 2x the number of workers to allow for some queuing
	}

	jobs.ctx, jobs.cancel = context.WithCancel(context.Background())

	for i := 0; i < config.BackgroundSyncWorkers; i++ {
		jobs.wg.Add(1)
		go jobs.worker(i)
	}

	return jobs
}

func (b *SyncJobs) Start() {
	b.wg.Add(1)
	go b.scheduler()
}

func (b *SyncJobs) scheduler() {
	defer b.wg.Done()

	syncTicker := time.NewTicker(b.interval)
	defer syncTicker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return

		case <-syncTicker.C:
			b.Tick()
		}
	}
}

func (b *SyncJobs) Tick() {
	defer utils.TrackTime(time.Now(), "SyncJobs.Tick", nil, nil)
	users := b.controlDB.AllUsers()

	nsUniqHashMap := map[string]*SyncJobWithData{} // hostname to sync jobs

	for _, userDB := range users {
		ns, dex, err := userDB.GetAllSyncSources(context.TODO())
		if err != nil {
			slog.Error("error getting sync sources for user", "subdomain", userDB.Subdomain, "error", err)
			continue
		}
		err = addSyncJobToHashMap(b.config, userDB, ns, nsUniqHashMap)
		if err != nil {
			slog.Error("error adding Nightscout sync job to hash map", "subdomain", userDB.Subdomain, "error", err, "ns", ns)
		}

		err = addSyncJobToHashMap(b.config, userDB, dex, nsUniqHashMap)
		if err != nil {
			slog.Error("error adding Dexcom sync job to hash map", "subdomain", userDB.Subdomain, "error", err, "dexcom", dex)
		}
	}

	// Need to get the user from the DB as well because we need to check last active time (caching in memory would cause a stampede of
	for _, syncJobs := range nsUniqHashMap {
		// Non-blocking submit (prevents deadlock if full)
		select {
		case b.jobCh <- syncJobs:
		case <-b.ctx.Done():
			return
		}

	}
}

func addSyncJobToHashMap(config config.ServerConfig, userDB *db.UserDB, syncSource *user.SyncSource, nsUniqHashMap map[string]*SyncJobWithData) error {
	if syncSource == nil || !syncSource.Enabled {
		return nil
	}

	unencryptedPassword, err := config.Security.Decrypt(syncSource.Password)
	if err != nil {
		return fmt.Errorf("error decrypting sync source password: %w", err)
	}
	unencryptedURLOrUsername, err := config.Security.Decrypt(syncSource.UrlOrUsername)
	if err != nil {
		return fmt.Errorf("error decrypting sync source URL or username: %w", err)
	}

	// Any two users with this constalations will be making the exact same request to the same URL.
	// We are deduping these requests to prevent multiple calls
	uniqString := fmt.Sprintf("%s:%s:%s", syncSource.Type, unencryptedURLOrUsername, unencryptedPassword)
	if syncSource.Type == consts.SyncSourceTypeDexcom {
		uniqString = fmt.Sprintf("%s:%s:%s:%s", syncSource.Type, unencryptedURLOrUsername, unencryptedPassword, syncSource.DexcomRegion)
	}
	uniqHash, err := config.Security.GenerateHash(uniqString)
	if err != nil {
		return fmt.Errorf("error generating hash for sync job: %w", err)
	}

	jobData, exists := nsUniqHashMap[uniqHash]
	if !exists {
		nsUniqHashMap[uniqHash] = &SyncJobWithData{
			uniqHash:                 uniqHash,
			users:                    []*UserDBSyncData{{UserDB: userDB, SyncSource: syncSource}},
			synctype:                 syncSource.Type,
			unencryptedPassword:      unencryptedPassword,
			unencryptedURLOrUsername: unencryptedURLOrUsername,
			dexcomRegion:             syncSource.DexcomRegion,
		}
	} else {
		jobData.users = append(jobData.users, &UserDBSyncData{UserDB: userDB, SyncSource: syncSource})
	}

	return nil
}

func (b *SyncJobs) worker(id int) {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return

		case syncJob := <-b.jobCh:
			slog.Info("starting sync", "worker_id", id, "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype)
			b.SyncUserData(syncJob)
		}
	}
}

func (b *SyncJobs) Stop() {
	// This will send a cancel signal to context
	b.cancel()
	// This will wait
	b.wg.Wait()

}

/////////////
// Actual SyncJob Work
//////////////

func (b *SyncJobs) SyncUserData(syncJob *SyncJobWithData) {
	var err error
	syncJob.report, err = b.controlDB.Queries.CreateOrUpdateSyncSourceReport(context.TODO(), syncJob.uniqHash)
	if err != nil {
		slog.Error("error creating or updating sync source report", "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype, "error", err)
		return
	}

	startTime := time.Now()
	switch syncJob.synctype {
	case consts.SyncSourceTypeNightscout:
		err = b.syncNightscout(syncJob)
	case consts.SyncSourceTypeDexcom:
		err = b.syncDexcom(syncJob)
	}
	elapsedTimeMS := time.Since(startTime).Milliseconds()

	if err != nil {
		errMsg := err.Error()
		slog.Error("error syncing user data", "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype, "error", err)
		b.controlDB.Queries.UpdateSyncSourceReportsFailure(context.TODO(), &elapsedTimeMS, &errMsg, syncJob.uniqHash)
		return
	}

	slog.Info("finished sync", "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype, "elapsed_time_ms", elapsedTimeMS)
	b.controlDB.Queries.UpdateSyncSourceReportsSuccess(context.TODO(), &elapsedTimeMS, syncJob.uniqHash)

	// TODO this is where we will sync data from the user DB to the control DB, and also do any processing needed
	// This will be called on a schedule, and also on demand when a user logs in or something like that
}

func (b *SyncJobs) syncNightscout(syncJob *SyncJobWithData) error {
	slog.Info("Syncing Nightscout", "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype, "url", syncJob.unencryptedURLOrUsername)
	return nil
}

func (b *SyncJobs) syncDexcom(syncJob *SyncJobWithData) error {
	slog.Info("Syncing Dexcom", "uniq_hash", syncJob.uniqHash, "subdomains", syncJob.Subdomains(), "sync_type", syncJob.synctype, "username", syncJob.unencryptedURLOrUsername, "dexcom_region", syncJob.dexcomRegion)
	return nil
}
