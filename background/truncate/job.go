package truncate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/config"
)

const TRUNCATE_JOB = "truncate_job"

type TruncateJobManager struct {
	TruncateIntervalInSeconds int
}

func SetupTruncateJobManager(config config.ServerConfig) *TruncateJobManager {
	jobs := &TruncateJobManager{
		TruncateIntervalInSeconds: config.TruncateIntervalInSeconds,
	}
	return jobs
}

func (t *TruncateJobManager) Truncate() {
	slog.Info("Truncate job unimplemented")
}

func (t *TruncateJobManager) GetConsumer(queue *liteq.JobQueue) *liteq.ConsumerSupervisor {
	return liteq.NewConsumerSupervisor(queue, TRUNCATE_JOB, 1, t.workerFunc())
}

func (t *TruncateJobManager) workerFunc() func(ctx context.Context, j *control.Job) error {
	return func(ctx context.Context, j *control.Job) error {
		return nil
	}
}

func (t *TruncateJobManager) SetupCron(c *cron.Cron, queue *liteq.JobQueue) {
	truncateCron := fmt.Sprintf("@every %ds", t.TruncateIntervalInSeconds)
	c.AddFunc(truncateCron, buildTruncateCronFunc(queue))
}

func buildTruncateCronFunc(queue *liteq.JobQueue) func() {
	return func() {
		err := queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue: TRUNCATE_JOB,
			Job:   "",
		})
		if err != nil {
			slog.Error("Error queueing truncate job", "error", err)
		}
	}
}
