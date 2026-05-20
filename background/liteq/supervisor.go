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
