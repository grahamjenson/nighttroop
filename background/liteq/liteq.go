package liteq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/db"

	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/aslog"
)

func New(controlDB *db.ControlDB) *JobQueue {
	return &JobQueue{queries: controlDB.Jobs.Queries}
}

type JobQueue struct {
	queries *jobsdb.Queries
}

type QueueJobParams struct {
	Queue             string
	Job               string
	ExecuteAfter      int64
	RemainingAttempts int64
	DedupingKey       DedupingKey
}

type DedupingKey interface {
	String() string
	ReplaceDuplicate() bool
}

type IgnoreDuplicate string

func (i IgnoreDuplicate) String() string {
	return string(i)
}
func (i IgnoreDuplicate) ReplaceDuplicate() bool {
	return false
}

type ReplaceDuplicate string

func (r ReplaceDuplicate) String() string {
	return string(r)
}
func (r ReplaceDuplicate) ReplaceDuplicate() bool {
	return true
}

func (j *JobQueue) QueueJob(ctx context.Context, params QueueJobParams) error {
	if params.RemainingAttempts == 0 {
		params.RemainingAttempts = 1
	}

	if params.DedupingKey == nil {
		params.DedupingKey = IgnoreDuplicate("")
	}

	doParams := jobsdb.QueueJobIgnoreDupeParams{
		Queue:             params.Queue,
		Job:               params.Job,
		ExecuteAfter:      params.ExecuteAfter,
		RemainingAttempts: params.RemainingAttempts,
		DedupingKey:       params.DedupingKey.String(),
	}

	if params.DedupingKey.String() == "" {
		return j.queries.QueueJobIgnoreDupe(ctx, doParams)
	}

	if params.DedupingKey.ReplaceDuplicate() {
		return j.queries.QueueJobReplaceDupe(ctx, jobsdb.QueueJobReplaceDupeParams(doParams))
	}

	return j.queries.QueueJobIgnoreDupe(ctx, doParams)
}

type GrabJobsParams struct {
	Queue        string
	ExecuteAfter int64
	Count        int64
}

func (j *JobQueue) GrabJobs(ctx context.Context, params GrabJobsParams) ([]*jobsdb.Job, error) {
	executeAfter := time.Now().Unix()
	if params.ExecuteAfter > 0 {
		executeAfter = params.ExecuteAfter
	}
	limit := int64(1)
	if params.Count > 0 {
		limit = params.Count
	}

	return j.queries.MarkJobsForConsumer(ctx, jobsdb.MarkJobsForConsumerParams{
		Queue:        params.Queue,
		ExecuteAfter: executeAfter,
		Limit:        limit,
	})
}

type WorkerFunc func(context.Context, *jobsdb.Job) error

// jobContext is ctx for one run of a queued job: what is logged with it says its queue, its row in the jobs DB and a
// job_id of its own
// runJob runs the worker on the job and records how it went. A failure keeps its run's job_id, so the control site can
// link to its logs.
func runJob(ctx context.Context, q *jobsdb.Queries, worker WorkerFunc, job *jobsdb.Job) {
	runCtx := jobContext(ctx, job)
	if err := worker(runCtx, job); err != nil {
		_, runID := aslog.IDs(runCtx)
		q.FailJob(ctx, jobsdb.FailJobParams{Errors: job.Errors.Add(err.Error()), RunID: &runID, ID: job.ID})
		return
	}
	q.CompleteJob(ctx, jobsdb.CompleteJobParams{ID: job.ID})
}

func jobContext(ctx context.Context, job *jobsdb.Job) context.Context {
	return aslog.NewJob(ctx, job.Queue, slog.Int64("queued_job_id", job.ID))
}

type ConsumeParams struct {
	Queue             string
	PoolSize          int
	Worker            WorkerFunc
	VisibilityTimeout int64
	OnEmptySleep      time.Duration
}

// Consume runs the queue's jobs on PoolSize workers until ctx is done. It only grabs as many jobs as there are free
// workers, so a job is fetched just before it runs and its visibility timeout counts from then. Stopping waits for the
// running jobs, not for jobs still queued.
func (j *JobQueue) Consume(ctx context.Context, params ConsumeParams) error {
	// free holds a slot for each worker that is not running a job
	poolSize := max(params.PoolSize, 1)
	free := make(chan struct{}, poolSize)
	for range poolSize {
		free <- struct{}{}
	}
	var running sync.WaitGroup
	defer running.Wait()
	sleep := params.OnEmptySleep
	if sleep == 0 {
		sleep = 1 * time.Second
	}
	for {
		// Wait for a free worker, unless the context gets canceled for example, then stop consuming
		select {
		case <-ctx.Done():
			return nil
		case <-free:
		}
		// Only this loop takes slots, so the other free ones can be taken without waiting
		slots := 1
		for len(free) > 0 {
			<-free
			slots++
		}

		if params.VisibilityTimeout > 0 {
			_, err := j.queries.ResetJobs(ctx, jobsdb.ResetJobsParams{
				Queue:             params.Queue,
				ConsumerFetchedAt: time.Now().UnixMilli() - (params.VisibilityTimeout * 1000),
			})

			if err != nil {
				return fmt.Errorf("error resetting jobs: %w", err)
			}
		}

		grabbedJobs, err := j.GrabJobs(ctx, GrabJobsParams{
			Queue:        params.Queue,
			ExecuteAfter: time.Now().Unix(),
			Count:        int64(slots),
		})

		if err != nil {
			return fmt.Errorf("error grabbing jobs: %w", err)
		}

		// Give back the slots no job was grabbed for
		for range slots - len(grabbedJobs) {
			free <- struct{}{}
		}

		if len(grabbedJobs) == 0 {
			time.Sleep(sleep)
			continue
		}

		for _, job := range grabbedJobs {
			running.Add(1)
			go func() {
				defer running.Done()
				defer func() { free <- struct{}{} }()
				runJob(ctx, j.queries, params.Worker, job)
			}()
		}
	}
}
