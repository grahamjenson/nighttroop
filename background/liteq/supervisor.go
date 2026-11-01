package liteq

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

///////
// Consumer Wrapper
///////

type ConsumerSupervisor struct {
	queue *JobQueue

	queueName string
	poolSize  int

	worker WorkerFunc

	consumerContext context.Context
	stopConsumer    context.CancelFunc

	wg sync.WaitGroup
}

func NewConsumerSupervisor(
	q *JobQueue,
	queueName string,
	poolSize int,
	worker WorkerFunc,
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

}

func (c *ConsumerSupervisor) Wait() {
	c.wg.Wait()
}

func (c *ConsumerSupervisor) loop() {
	retryDelay := 2 * time.Second

	for {
		slog.Info("starting liteq consumer for queue", "queue", c.queueName)

		err := c.queue.Consume(c.consumerContext, ConsumeParams{
			Queue:             c.queueName,
			PoolSize:          c.poolSize,
			Worker:            c.worker,
			VisibilityTimeout: 60, // Max time a job can take before being reset to be failed or retried
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

// RunQueued runs every job waiting in the queue now, one at a time, and returns once it is empty.
// Tests use it in place of Start so background work has finished before they check its results.
func (c *ConsumerSupervisor) RunQueued(ctx context.Context) error {
	for {
		jobs, err := c.queue.GrabJobs(ctx, GrabJobsParams{Queue: c.queueName, Count: 1})
		if err != nil || len(jobs) == 0 {
			return err
		}
		job := jobs[0]
		runJob(ctx, c.queue.queries, c.worker, job)
	}
}
