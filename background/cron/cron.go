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
	"nighttroop.t1d.nz/utils/errors"
)

type CronScheduler interface {
	AddFunc(name string, spec string, cmd func() error)
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

func (c *ACron) AddFunc(name string, spec string, cmd func() error) {
	_, err := c.RawAddFunc(name, spec, cmd)
	if err != nil {
		panic(fmt.Sprintf("Error adding cron function: %v", err))
	}
}

func (c *ACron) RawAddFunc(name string, spec string, cmd func() error) (cron.EntryID, error) {
	// Skip if still running channel
	var ch = make(chan struct{}, 1)
	ch <- struct{}{}

	wrappedCmd := func() {
		startedAt := time.Now().UnixMilli()

		defer func() {
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v:%T", r, r)
				}
				slog.Error("Cron: Command panicked", "name", name, "stack", string(buf), "error", err)
				c.jobsDB.Queries.InsertFailedCronJob(context.TODO(), jobsdb.InsertFailedCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
					Errors:            errors.ErrorList{err.Error()},
				})
			}
		}()

		select {
		case v := <-ch:
			if err := cmd(); err != nil {
				slog.Error("Cron: Command returned error", "name", name, "error", err)
				c.jobsDB.Queries.InsertFailedCronJob(context.TODO(), jobsdb.InsertFailedCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
					Errors:            errors.ErrorList{err.Error()},
				})
			} else {
				c.jobsDB.Queries.InsertSuccessfulCronJob(context.TODO(), jobsdb.InsertSuccessfulCronJobParams{
					Job:               name,
					ConsumerFetchedAt: startedAt,
				})
			}
			ch <- v
		default:
			slog.Info("Cron: Skipping execution because previous run is still in progress", "name", name)
		}

	}

	return c.Cron.AddFunc(spec, wrappedCmd)
}
