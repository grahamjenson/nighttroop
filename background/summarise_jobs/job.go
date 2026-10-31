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
	// SummariseAfterSeconds is how long a finished job is kept before it is summarised, none in tests so they see the stats at once
	SummariseAfterSeconds int64
	JobsDB                *db.JobsDB
}

func SetupTruncateJobManager(config *config.ServerConfig, controlDB *db.ControlDB) *TruncateJobManager {
	jobs := &TruncateJobManager{
		TruncateIntervalInSeconds: config.TruncateIntervalInSeconds,
		SummariseAfterSeconds:     60,
		JobsDB:                    controlDB.Jobs,
	}
	if config.IsTest() {
		jobs.SummariseAfterSeconds = 0
	}
	return jobs
}

func (t *TruncateJobManager) SetupCron(c cron.CronScheduler) {
	truncateCron := fmt.Sprintf("@every %ds", t.TruncateIntervalInSeconds)
	c.AddFunc("TruncateJobManager", truncateCron, func(ctx context.Context) error {
		err := t.JobsDB.BuildDailyReports(ctx, t.SummariseAfterSeconds)
		if err != nil {
			slog.ErrorContext(ctx, "Error building daily reports", "error", err)
		}
		return err
	})
}
