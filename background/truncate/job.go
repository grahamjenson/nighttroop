package truncate

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils/acron"
	"nighttroop.t1d.nz/utils/config"
)

const TRUNCATE_JOB = "truncate_job"

const TRUNCATE_SITE_JOB = "truncate_user_job"
const TRUNCATE_JOBS_JOB = "truncate_jobs"

type TruncateJobManager struct {
	TruncateIntervalInSeconds int
	JobsDB                    *db.JobsDB
}

func SetupTruncateJobManager(config config.ServerConfig, controlDB *db.ControlDB) *TruncateJobManager {
	jobs := &TruncateJobManager{
		TruncateIntervalInSeconds: config.TruncateIntervalInSeconds,
		JobsDB:                    controlDB.Jobs,
	}
	return jobs
}

func (t *TruncateJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, TRUNCATE_JOB, 1, t.workerFunc())
}

func (t *TruncateJobManager) workerFunc() func(ctx context.Context, j *jobs.Job) error {
	return func(ctx context.Context, j *jobs.Job) error {
		switch j.Job {
		case TRUNCATE_SITE_JOB:
			err := t.TruncateUsers()
			if err != nil {
				slog.Error("Error truncating users", "error", err)
				return err
			}
		case TRUNCATE_JOBS_JOB:
			err := t.TruncateJobs()
			if err != nil {
				slog.Error("Error truncating jobs", "error", err)
				return err
			}
		default:
			slog.Error("Unknown truncate job type", "job", j.Job)
		}
		return nil
	}

}

func (t *TruncateJobManager) SetupCron(c acron.CronScheduler, queue *liteq.JobQueue) {
	truncateCron := fmt.Sprintf("@every %ds", t.TruncateIntervalInSeconds)
	c.AddFunc("TruncateJobManager", truncateCron, buildTruncateCronFunc(queue))
}

func buildTruncateCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		err := queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue: TRUNCATE_JOB,
			Job:   TRUNCATE_SITE_JOB,
		})
		err = queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue: TRUNCATE_JOB,
			Job:   TRUNCATE_JOBS_JOB,
		})

		if err != nil {
			slog.Error("Error queueing truncate job", "error", err)
		}
	}
}
