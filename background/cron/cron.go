package cron

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/errors"
)

// CronScheduler runs jobs on a schedule. Each run's context logs the job's name and a job_id of its own.
type CronScheduler interface {
	AddFunc(name string, spec string, cmd func(ctx context.Context) error)
}

type ACron struct {
	*cron.Cron
	jobsDB *db.JobsDB
}

func NewCron(jobsDB *db.JobsDB) *ACron {

	return &ACron{
		Cron: cron.New(
			cron.WithSeconds(),
		),

		jobsDB: jobsDB,
	}
}

func (c *ACron) AddFunc(name string, spec string, cmd func(ctx context.Context) error) {
	_, err := c.RawAddFunc(name, spec, cmd)
	if err != nil {
		panic(fmt.Sprintf("Error adding cron function: %v", err))
	}
}

func (c *ACron) RawAddFunc(name string, spec string, cmd func(ctx context.Context) error) (cron.EntryID, error) {
	// Skip if still running channel
	var ch = make(chan struct{}, 1)
	ch <- struct{}{}

	wrappedCmd := func() {
		startedAt := time.Now().UnixMilli()
		ctx := aslog.NewJob(context.Background(), name)

		defer func() {
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v:%T", r, r)
				}
				slog.ErrorContext(ctx, "Cron: Command panicked", "stack", string(buf), "error", err)
				c.jobsDB.Queries.InsertFailedCronJob(ctx, jobsdb.InsertFailedCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
					Errors:            errors.ErrorList{}.Add(err.Error()),
				})
			}
		}()

		select {
		case v := <-ch:
			// Deferred so a panic in cmd still lets the next run start
			defer func() { ch <- v }()
			if err := cmd(ctx); err != nil {
				slog.ErrorContext(ctx, "Cron: Command returned error", "error", err)
				c.jobsDB.Queries.InsertFailedCronJob(ctx, jobsdb.InsertFailedCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
					Errors:            errors.ErrorList{}.Add(err.Error()),
				})
			} else {
				c.jobsDB.Queries.InsertSuccessfulCronJob(ctx, jobsdb.InsertSuccessfulCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
				})
			}
		default:
			slog.InfoContext(ctx, "Cron: Skipping execution because previous run is still in progress")
		}

	}

	return c.Cron.AddFunc(spec, wrappedCmd)
}
