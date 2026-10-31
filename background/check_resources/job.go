// Package check_resources watches the server's CPU, memory and disk, warning the control site when one gets high
package check_resources

import (
	"context"
	"runtime"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/usage"
)

// SetupCron samples the server's CPU every 5 seconds, to see how long it stays pegged, and checks its resources every
// 5 minutes
func SetupCron(c cron.CronScheduler, cfg *config.ServerConfig, controlDB *db.ControlDB) {
	c.AddFunc("SampleCPU", "@every 5s", func() error {
		cfg.Usage.SampleCPU(time.Now(), runtime.NumCPU())
		return nil
	})
	c.AddFunc("CheckResources", "@every 5m", func() error { return Check(context.Background(), cfg, controlDB, time.Now()) })
}

// Check reads the CPU sampled since the last check, the memory used and the DBs' disk, keeping them in the live DB and
// recording an event for each that went over its alert level or came back under it
func Check(ctx context.Context, cfg *config.ServerConfig, controlDB *db.ControlDB, now time.Time) error {
	cpu, pegged := cfg.Usage.TakeCPU()
	reading := livedb.CreateResourceReadingParams{At: now.UnixMilli(), CPUPercent: cpu, CPUPeggedSeconds: int64(pegged.Seconds())}
	if memory, ok := usage.MemoryPercent(); ok {
		reading.MemoryPercent = &memory
	}
	// The DB folder's disk, or the working folder's before the DB folder is made
	disk, ok := usage.DiskPercent(cfg.DBDir())
	if !ok {
		disk, ok = usage.DiskPercent(".")
	}
	if ok {
		reading.DiskPercent = &disk
	}
	return controlDB.RecordResources(ctx, reading)
}
