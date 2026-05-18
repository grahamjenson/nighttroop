package background

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
)

// CronJobs will run and queue jobs into the
func CronSetup(queue *liteq.JobQueue) *cron.Cron {
	c := cron.New(cron.WithSeconds())
	// Every 30 seconds
	c.AddFunc("*/10 * * * * *", func() {
		fmt.Println("I am in the cron job " + time.Now().String())
		queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue: "backup",
			Job:   "backup1",

			DedupingKey: liteq.ReplaceDuplicate("backup"),
		})

		queue.QueueJob(context.Background(), liteq.QueueJobParams{
			Queue: "backup",
			Job:   "backup2",

			DedupingKey: liteq.ReplaceDuplicate("backup"),
		})

	})
	c.Start()

	return c
}

func QueueSetup(controlDB *db.ControlDB) *liteq.JobQueue {
	q := liteq.New(controlDB)

	return q
}

func SetupBackGround(ctx context.Context, controlDB *db.ControlDB) (*cron.Cron, *liteq.JobQueue) {
	queue := QueueSetup(controlDB)
	cron := CronSetup(queue)

	go queue.Consume(ctx, liteq.ConsumeParams{
		Queue:    "backup",
		PoolSize: 1,
		Worker: func(ctx context.Context, j *control.Job) error {
			fmt.Printf("I am in the job %s %s\n", j.Job, time.Now().String())
			time.Sleep(20 * time.Second)
			fmt.Printf("finsihed job %s %s\n", j.Job, time.Now().String())
			return nil
		},
	})

	return cron, queue
}
