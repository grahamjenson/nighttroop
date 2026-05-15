package sync

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
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
//
// To deal with these issues:
// 1. Cant point URL at Nighttroop (or any subdomain of Nighttroop)
// 2. If a users sync is taking too long or failing decrease the rate at which they sync (until they are disabled)
// 3. Check for users pointing to the same host, find the user that last updated their sync, disable the rest
// 4. Agressive timeouts for jobs as well as logging
// 5. List of bad and blocked hostnames, that are never allowed to be synced from (TODO)
// 6. We only Sync users that have called a read endpoint recently.
// 7. We slow down syncing of users that don't actually get new data (count writes to the DB)

// Dexcom is easier, but if say the dexcom servers are down we might encounter lots of timeouts, which we should detect quickly and slow down all syncing.

type SyncJobs struct {
	controlDB *db.ControlDB
	ctx       context.Context
	cancel    func()
	wg        sync.WaitGroup
	interval  time.Duration

	jobCh chan *db.UserDB
}

// Background Jobs

func SetupSyncJobs(config config.ServerConfig, controlDB *db.ControlDB) *SyncJobs {

	jobs := &SyncJobs{
		controlDB: controlDB,
		interval:  time.Duration(30 * time.Second),
		jobCh:     make(chan *db.UserDB, 100),
	}

	jobs.ctx, jobs.cancel = context.WithCancel(context.Background())

	for i := 0; i < 100; i++ {
		jobs.wg.Add(1)
		go jobs.worker(i)
	}

	return jobs
}

func (b *SyncJobs) Start() {
	syncTicker := time.NewTicker(b.interval)
	defer syncTicker.Stop()

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
			users := b.controlDB.AllUsers()
			// Need to get the user from the DB as well because we need to check last active time (caching in memory would cause a stampede of
			for _, userDB := range users {

				// Non-blocking submit (prevents deadlock if full)
				select {
				case b.jobCh <- userDB:
				case <-b.ctx.Done():
					return
				}
			}
		}
	}
}

func (b *SyncJobs) worker(id int) {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return

		case userDB := <-b.jobCh:
			slog.Info("starting sync", "worker_id", id, "subdomain", userDB.Subdomain)
			SyncUserData(userDB)
		}
	}
}

func SyncUserData(userDB *db.UserDB) {
	// TODO this is where we will sync data from the user DB to the control DB, and also do any processing needed
	// This will be called on a schedule, and also on demand when a user logs in or something like that
}

func (b *SyncJobs) Stop() {
	// This will send a cancel signal to context
	b.cancel()
	// This will wait
	b.wg.Wait()

}
