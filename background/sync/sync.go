// Make the backgorund jobs orthogonal to the DB
package sync

import (
	"context"
	"sync"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

type SyncJobs struct {
	controlDB *db.ControlDB
	ctx       context.Context
	cancel    func()
	wg        sync.WaitGroup
}

// Background Jobs

func SetupSyncJobs(config config.ServerConfig, controlDB *db.ControlDB) *SyncJobs {

	jobs := &SyncJobs{
		controlDB: controlDB,
	}

	jobs.ctx, jobs.cancel = context.WithCancel(context.Background())

	return jobs
}

func (b *SyncJobs) Start() {
	// Start backup Jobs
	// TODO sync jobs and all that stuff
}

func (b *SyncJobs) Stop() {
	// This will send a cancel signal to context
	b.cancel()
	// This will wait
	b.wg.Wait()

}
