package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/config"
)

type JobsDB struct {
	*jobsdb.DB
}

// NewJobsDB opens the jobs DB, made again, empty, if it does not open
func NewJobsDB(ctx context.Context, cfg *config.ServerConfig) (*JobsDB, error) {
	rawDB, err := openDisposable(ctx, jobsdb.GetPath(cfg), func() (*jobsdb.DB, error) { return jobsdb.NewDB(ctx, cfg) })
	if err != nil {
		return nil, fmt.Errorf("cannot open jobs db: %w", err)
	}
	return &JobsDB{DB: rawDB}, nil
}

// openDisposable opens a DB that is not backed up, as its rows are not worth stopping a start for: one that does not
// open, as when its file is corrupt or its migration fails, is removed and made again, empty, and logged as an error
func openDisposable[D any](ctx context.Context, path string, open func() (*D, error)) (*D, error) {
	opened, err := open()
	if err == nil {
		return opened, nil
	}
	slog.ErrorContext(ctx, "DB does not open, so it is removed and made again, empty", "path", path, "error", err)
	for _, file := range dbFiles(path) {
		if err := os.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("cannot remove the DB that did not open: %w", err)
		}
	}
	return open()
}

func (j *JobsDB) BuildDailyReports(ctx context.Context, secondsAgo int64) error {
	q := j.Queries

	// Get the Jobs to be summarised
	gotJobs, err := q.GetJobsToBeSummarised(ctx, jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: secondsAgo,
	})
	if err != nil {
		return fmt.Errorf("Error getting jobs to be summarised: %w", err)
	}

	ids := make([]int64, len(gotJobs))
	dayQueueJobMap := make(map[string]jobsdb.JobStatsDaily)
	for i, j := range gotJobs {
		ids[i] = j.ID

		stats, ok := dayQueueJobMap[j.DayQueueJob()]
		if ok {
			stats.TotalJobs += 1
			stats.TotalSuccess += j.TotalSuccess()
			stats.TotalFailed += j.TotalFailed()
			stats.DurationSum += j.Duration()
			stats.DurationSquaredSum += j.DurationSquared()
			stats.Errors = stats.Errors.AddAll(j.Errors)
			stats.FailedRunIdsJson = jobsdb.LatestRunIDs(stats.FailedRunIdsJson, j.FailedRunID()...)
			dayQueueJobMap[j.DayQueueJob()] = stats
		} else {
			dayQueueJobMap[j.DayQueueJob()] = jobsdb.JobStatsDaily{
				DayQueueJob:        j.DayQueueJob(),
				TotalJobs:          1,
				TotalSuccess:       j.TotalSuccess(),
				TotalFailed:        j.TotalFailed(),
				DurationSum:        j.Duration(),
				DurationSquaredSum: j.DurationSquared(),
				Errors:             j.Errors,
				FailedRunIdsJson:   jobsdb.LatestRunIDs(nil, j.FailedRunID()...),
			}
		}
	}

	for k, stats := range dayQueueJobMap {
		existing, err := q.EnsureJobStatsDailyForGroup(ctx, jobsdb.EnsureJobStatsDailyForGroupParams{
			DayQueueJob: k,
		})
		if err != nil {
			return fmt.Errorf("Error ensuring job stats daily for group: %w", err)
		}
		// Merged here, not in SQL, so the stored list stays bounded and the update cannot fail its CHECK
		stats.Errors = existing.Errors.AddAll(stats.Errors)
		stats.FailedRunIdsJson = jobsdb.LatestRunIDs(existing.FailedRunIdsJson, stats.FailedRunIdsJson...)
		dayQueueJobMap[k] = stats
	}

	tx, err := j.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error starting transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := j.Queries.WithTx(tx)

	for _, stats := range dayQueueJobMap {
		err = qtx.UpdateJobStatsDaily(ctx, jobsdb.UpdateJobStatsDailyParams{
			DayQueueJob:        stats.DayQueueJob,
			TotalJobs:          stats.TotalJobs,
			TotalSuccess:       stats.TotalSuccess,
			TotalFailed:        stats.TotalFailed,
			DurationSum:        stats.DurationSum,
			DurationSquaredSum: stats.DurationSquaredSum,
			Errors:             stats.Errors,
			FailedRunIdsJson:   stats.FailedRunIdsJson,
		})
		if err != nil {
			return fmt.Errorf("Error updating job stats daily: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("Error updating job stats daily: %w", err)
	}

	err = qtx.DeleteJobsByIds(ctx, jobsdb.DeleteJobsByIdsParams{
		Ids: ids,
	})
	if err != nil {
		return fmt.Errorf("Error deleting old jobs: %w", err)
	}

	return tx.Commit()
}
