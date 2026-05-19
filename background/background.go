package background

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/config"
)

type BackGroundManager struct {
	db     *db.ControlDB
	config config.ServerConfig
	cron   *cron.Cron
	queue  *liteq.JobQueue

	backgroundConsumer *ConsumerSupervisor
}

func NewBackGroundManager(controlDB *db.ControlDB, config config.ServerConfig) *BackGroundManager {
	queue := liteq.New(controlDB)
	cron := cronSetup(queue)

	backgroundConsumer := newConsumerSupervisor(
		queue,
		"backup",
		2,
		func(ctx context.Context, j *control.Job) error {
			fmt.Printf("job start %s %d\n", j.Job, j.ID)
			time.Sleep(10 * time.Second) // Simulate work
			select {
			case <-time.After(20 * time.Second):
				fmt.Println("Finished job work")
			case <-ctx.Done():
				fmt.Println("DID NOT Finished job work")
				return ctx.Err()
			}

			fmt.Printf("job done %s %d\n", j.Job, j.ID)
			return nil
		},
	)

	return &BackGroundManager{
		db:                 controlDB,
		config:             config,
		cron:               cron,
		queue:              queue,
		backgroundConsumer: backgroundConsumer,
	}
}

func (b *BackGroundManager) Start() {

	b.backgroundConsumer.Start()
	b.cron.Start()
}

func (b *BackGroundManager) Stop() {
	b.cron.Stop()
	b.backgroundConsumer.Stop()
}

// CronJobs will run and queue jobs into the
func cronSetup(queue *liteq.JobQueue) *cron.Cron {
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

	return c
}

///////
// Consumer Wrapper
///////

type ConsumerSupervisor struct {
	queue *liteq.JobQueue

	queueName string
	poolSize  int

	worker liteq.WorkerFunc

	consumerContext context.Context
	stopConsumer    context.CancelFunc

	wg sync.WaitGroup
}

func newConsumerSupervisor(
	q *liteq.JobQueue,
	queueName string,
	poolSize int,
	worker liteq.WorkerFunc,
) *ConsumerSupervisor {

	consumerContext, stopConsumer := context.WithCancel(context.Background())

	return &ConsumerSupervisor{
		queue:           q,
		queueName:       queueName,
		poolSize:        poolSize,
		worker:          worker,
		consumerContext: consumerContext,
		stopConsumer:    stopConsumer,
	}
}

func (c *ConsumerSupervisor) Start() {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		c.loop()
	}()
}

func (c *ConsumerSupervisor) Stop() {
	c.stopConsumer()
	c.wg.Wait()
}

func (c *ConsumerSupervisor) loop() {
	retryDelay := 2 * time.Second

	for {
		slog.Info("starting liteq consumer for queue", "queue", c.queueName)

		err := c.queue.Consume(c.consumerContext, liteq.ConsumeParams{
			Queue:    c.queueName,
			PoolSize: c.poolSize,
			Worker:   c.worker,
		})

		if err != nil {
			slog.Error("QueueConsumer exited with error", "error", err, "queue", c.queueName)
		}

		timer := time.NewTimer(retryDelay)
		select {
		case <-c.consumerContext.Done():
			timer.Stop()
			slog.Info("QueueConsumer loop context cancelled, exiting loop", "queue", c.queueName)
			return
		case <-timer.C:
			slog.Info("QueueConsumer restarting consumer", "queue", c.queueName, "retryDelay", retryDelay.String())
		}
	}
}
