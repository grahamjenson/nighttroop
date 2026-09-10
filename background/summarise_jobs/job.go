package summarise_jobs

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

type TruncateJobManager struct {
	TruncateIntervalInSeconds int
	JobsDB                    *db.JobsDB
}

func SetupTruncateJobManager(config *config.ServerConfig, controlDB *db.ControlDB) *TruncateJobManager {
	jobs := &TruncateJobManager{
		TruncateIntervalInSeconds: config.TruncateIntervalInSeconds,
		JobsDB:                    controlDB.Jobs,
	}
	return jobs
}

func (t *TruncateJobManager) SetupCron(c cron.CronScheduler) {
	truncateCron := fmt.Sprintf("@every %ds", t.TruncateIntervalInSeconds)
	c.AddFunc("TruncateJobManager", truncateCron, func() error {
		err := t.JobsDB.BuildDailyReports(context.Background(), 60)
		if err != nil {
			slog.Error("Error building daily reports", "error", err)
		}
		return err
	})
}
