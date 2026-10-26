package trim_logs

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

// SetupCron trims the logs, events and errors tables to their newest LogsDBMaxRowsPerTable rows, as often as the
// jobs are summarised
func SetupCron(c cron.CronScheduler, config *config.ServerConfig, logsDB *db.LogsDB) {
	c.AddFunc("TrimLogsDB", fmt.Sprintf("@every %ds", config.TruncateIntervalInSeconds), func() error {
		return logsDB.Trim(context.Background(), int64(config.LogsDBMaxRowsPerTable))
	})
}
