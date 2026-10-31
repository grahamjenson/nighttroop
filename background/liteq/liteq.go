package liteq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alitto/pond/v2"

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

func (j *JobQueue) Consume(ctx context.Context, params ConsumeParams) error {
	workers := pond.NewPool(params.PoolSize)
	defer workers.StopAndWait()
	sleep := params.OnEmptySleep
	if sleep == 0 {
		sleep = 1 * time.Second
	}
	for {
		// If the context gets canceled for example, stop consuming
		if ctx.Err() != nil {
			return nil
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
			Count:        int64(params.PoolSize),
		})

		if err != nil {
			return fmt.Errorf("error grabbing jobs: %w", err)
		}

		if len(grabbedJobs) == 0 {
			time.Sleep(sleep)
			continue
		}

		for _, job := range grabbedJobs {
			job := job
			workers.Submit(func() {
				err := params.Worker(jobContext(ctx, job), job)
				if err != nil {
					j.queries.FailJob(ctx, jobsdb.FailJobParams{
						Errors: job.Errors.Add(err.Error()),
						ID:     job.ID,
					})
					return
				}

				j.queries.CompleteJob(ctx, jobsdb.CompleteJobParams{
					ID: job.ID,
				})
			})
		}
	}
}
