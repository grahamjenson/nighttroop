package trim_logs

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
)

// SetupCron trims the logs, events and errors tables to their newest LogsDBMaxRowsPerTable rows, as often as the
// jobs are summarised
func SetupCron(c cron.CronScheduler, config *config.ServerConfig, logsDB *db.LogsDB) {
	// Each minute, a summary of the requests refused before they signed in, which are each only a debug log
	c.AddFunc("SumRefusedRequests", "@every 1m", func(ctx context.Context) error { aslog.RefusedRequests.Flush(ctx); return nil })
	c.AddFunc("TrimLogsDB", fmt.Sprintf("@every %ds", config.TruncateIntervalInSeconds), func(ctx context.Context) error {
		return logsDB.Trim(ctx, int64(config.LogsDBMaxRowsPerTable))
	})
}
