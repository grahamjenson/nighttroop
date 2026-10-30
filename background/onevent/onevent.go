// Package onevent runs background jobs that events start: each job listens to the server's events, and one that
// matches its filter queues the job, which a consumer runs later, away from whatever published the event
package onevent

import (
	"context"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/jobsdb"
)

// Job is a background job an event starts. It runs once for each site, however many of its matching events were
// published before it ran, as jobs are kept apart by their queue and site.
type Job struct {
	// Queue names the job's queue, and its runs in the jobs DB
	Queue string
	// Filter is whether an event starts the job
	Filter func(events.Event) bool
	Run    func(context.Context, Fired) error
}

// Fired is what a job knows of the event that started it
type Fired struct {
	Name   events.Name `json:"name"`
	SiteID string      `json:"site_id"`
	Batch  string      `json:"batch"`
}

// Listen queues each job when an event published on bus matches it, and is the consumers that run them, to start
// with the server's others
func Listen(bus *events.SiteBus, queue *liteq.JobQueue, jobs ...Job) []*liteq.ConsumerSupervisor {
	consumers := []*liteq.ConsumerSupervisor{}
	for _, job := range jobs {
		// Publishing waits for a handler, so it only queues the job, as one quick write
		bus.SubscribeAll(func(event events.Event) {
			if !job.Filter(event) {
				return
			}
			fired, _ := json.Marshal(Fired{Name: event.Name, SiteID: event.SiteID, Batch: event.Batch})
			if err := queue.QueueJob(context.Background(), liteq.QueueJobParams{Queue: job.Queue, Job: string(fired),
				DedupingKey: liteq.ReplaceDuplicate(job.Queue + "-" + event.SiteID)}); err != nil {
				slog.Error("Cannot queue an event's job", "queue", job.Queue, "event", event.Name, "error", err)
			}
		})
		consumers = append(consumers, liteq.NewConsumerSupervisor(queue, job.Queue, 1, func(ctx context.Context, queued *jobsdb.Job) error {
			var fired Fired
			if err := json.Unmarshal([]byte(queued.Job), &fired); err != nil {
				return err
			}
			return job.Run(ctx, fired)
		}))
	}
	return consumers
}
