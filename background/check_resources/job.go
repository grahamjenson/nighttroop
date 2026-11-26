// Package check_resources watches the server's CPU, memory and disk, warning the control site when one gets high
package check_resources

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/metrics"
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/usage"
)

// checkEvery is how often the server's resources are checked, and checkSoon how long after start the first check is,
// once the CPU is sampled
const checkEvery, checkSoon = 30 * time.Second, 10 * time.Second

// SetupCron samples the server's CPU every 5 seconds, to see how long it stays pegged, checks its resources soon after
// start then every 30 seconds, and measures its sites every 5 minutes
func SetupCron(c cron.CronScheduler, cfg *config.ServerConfig, controlDB *db.ControlDB) {
	c.AddFunc("SampleCPU", "@every 5s", func(ctx context.Context) error {
		cfg.Usage.SampleCPU(time.Now(), runtime.NumCPU())
		return nil
	})
	c.AddFuncSoon("CheckResources", checkSoon, checkEvery, func(ctx context.Context) error { return Check(ctx, cfg, controlDB, time.Now()) })
	c.AddFunc("MeasureSites", "@every 5m", func(ctx context.Context) error { return MeasureSites(ctx, controlDB) })
}

// Check reads the CPU sampled since the last check, the memory used, the goroutines running and the DBs' disk, keeping them in the live DB and
// recording an event for each that went over its alert level or came back under it
func Check(ctx context.Context, cfg *config.ServerConfig, controlDB *db.ControlDB, now time.Time) error {
	cpu, pegged := cfg.Usage.TakeCPU()
	reading := livedb.CreateResourceReadingParams{At: now.UnixMilli(), CPUPercent: cpu, CPUPeggedSeconds: int64(pegged.Seconds()),
		Goroutines: int64(runtime.NumGoroutine())}
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
		refuseDiskFull(ctx, cfg, controlDB, total-used)
		metrics.DiskFree(total - used)
	}
	metrics.Sites(len(controlDB.AllSites()), controlDB.UnavailableSites())
	return controlDB.RecordResources(ctx, reading)
}

// refuseDiskFull keeps whether the DBs' disk, with free bytes free, is under DiskRefuseFreeMB, so sites are not created
// and site writes that add data are refused, and pages the admin when it becomes so, as only they can free space
func refuseDiskFull(ctx context.Context, cfg *config.ServerConfig, controlDB *db.ControlDB, free float64) {
	full := free < float64(cfg.DiskRefuseFreeMB)*(1<<20)
	if !controlDB.SetDiskFull(full) {
		return
	}
	if full {
		page.New(cfg.PagerDutyRoutingKey, cfg.PagerDutyEventsURL, cfg.HostName).Admin(ctx, page.SMS,
			"The DBs' disk is nearly full, so sites are not created and their writes are refused until space is freed", "free_mb", int(free)>>20)
		return
	}
	slog.InfoContext(ctx, "The DBs' disk has room again, so sites are created and written to again", "free_mb", int(free)>>20)
}

// MeasureSites keeps each site DB's size and the size of each of its tables, and scores each site for abuse, and
// measures the tables of the server's own DBs too. Each site's DB, and its tables, are measured here, as its syncs write to it
// without a request that would measure it, and measuring tables reads a whole DB, too slow for a page or a check every
// 30 seconds.
func MeasureSites(ctx context.Context, controlDB *db.ControlDB) error {
	sites := controlDB.AllSites()
	var bytes float64
	for _, site := range sites {
		if kb := site.EstimatedSizeKB(); kb != nil {
			bytes += float64(*kb) * 1024
		}
	}
	metrics.SiteDBs(bytes)
	return errors.Join(controlDB.RecordSiteSizes(ctx, sites), controlDB.RecordAbuseScores(ctx, sites),
		controlDB.RecordSiteTableSizes(ctx, sites), controlDB.RecordServerTableSizes(ctx))
}
