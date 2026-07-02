package liteq

import (
	"context"
	"fmt"
	"time"

	"github.com/alitto/pond/v2"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils/errors"
)

func New(controlDB *db.ControlDB) *JobQueue {
	return &JobQueue{queries: controlDB.Queries}
}

type JobQueue struct {
	queries *control.Queries
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

	doParams := control.QueueJobIgnoreDupeParams{
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
		return j.queries.QueueJobReplaceDupe(ctx, control.QueueJobReplaceDupeParams(doParams))
	}

	return j.queries.QueueJobIgnoreDupe(ctx, doParams)
}

type GrabJobsParams struct {
	Queue        string
	ExecuteAfter int64
	Count        int64
}

func (j *JobQueue) GrabJobs(ctx context.Context, params GrabJobsParams) ([]*control.Job, error) {
	executeAfter := time.Now().Unix()
	if params.ExecuteAfter > 0 {
		executeAfter = params.ExecuteAfter
	}
	limit := int64(1)
	if params.Count > 0 {
		limit = params.Count
	}

	return j.queries.MarkJobsForConsumer(ctx, control.MarkJobsForConsumerParams{
		Queue:        params.Queue,
		ExecuteAfter: executeAfter,
		Limit:        limit,
	})
}

type WorkerFunc func(context.Context, *control.Job) error

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
			_, err := j.queries.ResetJobs(ctx, control.ResetJobsParams{
				Queue:             params.Queue,
				ConsumerFetchedAt: time.Now().UnixMilli() - (params.VisibilityTimeout * 1000),
			})

			if err != nil {
				return fmt.Errorf("error resetting jobs: %w", err)
			}
		}

		jobs, err := j.GrabJobs(ctx, GrabJobsParams{
			Queue:        params.Queue,
			ExecuteAfter: time.Now().Unix(),
			Count:        int64(params.PoolSize),
		})

		if err != nil {
			return fmt.Errorf("error grabbing jobs: %w", err)
		}

		if len(jobs) == 0 {
			time.Sleep(sleep)
			continue
		}

		for _, job := range jobs {
			job := job
			workers.Submit(func() {
				err := params.Worker(ctx, job)
				if err != nil {
					j.queries.FailJob(ctx, control.FailJobParams{
						Errors: errors.ErrorList(append(job.Errors, err.Error())),
						ID:     job.ID,
					})
					return
				}

				j.queries.CompleteJob(ctx, control.CompleteJobParams{
					ID: job.ID,
				})
			})
		}
	}
}
