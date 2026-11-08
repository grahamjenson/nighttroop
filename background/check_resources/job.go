// Package check_resources watches the server's CPU, memory and disk, warning the control site when one gets high
package check_resources

import (
	"context"
	"errors"
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
	c.AddFunc("SampleCPU", "@every 5s", func(ctx context.Context) error {
		cfg.Usage.SampleCPU(time.Now(), runtime.NumCPU())
		return nil
	})
	c.AddFunc("CheckResources", "@every 5m", func(ctx context.Context) error { return Check(ctx, cfg, controlDB, time.Now()) })
}

// Check reads the CPU sampled since the last check, the memory used and the DBs' disk, keeping them in the live DB and
// recording an event for each that went over its alert level or came back under it, and keeps each site DB's size
func Check(ctx context.Context, cfg *config.ServerConfig, controlDB *db.ControlDB, now time.Time) error {
	cpu, pegged := cfg.Usage.TakeCPU()
	reading := livedb.CreateResourceReadingParams{At: now.UnixMilli(), CPUPercent: cpu, CPUPeggedSeconds: int64(pegged.Seconds())}
	if used, total, ok := usage.Memory(); ok {
		reading.MemoryPercent, reading.MemoryTotalBytes = new(100*used/total), new(int64(total))
	}
	// The DB folder's disk, or the working folder's before the DB folder is made
	used, total, ok := usage.Disk(cfg.DBDir())
	if !ok {
		used, total, ok = usage.Disk(".")
	}
	if ok {
		reading.DiskPercent, reading.DiskTotalBytes = new(100*used/total), new(int64(total))
	}
	// Each site's DB is measured here, as its syncs write to it without a request that would measure it
	return errors.Join(controlDB.RecordResources(ctx, reading), controlDB.RecordSiteSizes(ctx, controlDB.AllSites()))
}
