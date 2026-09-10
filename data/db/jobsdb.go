package db

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils/config"
)

type JobsDB struct {
	*jobs.DB
}

func NewJobsDB(ctx context.Context, cfg *config.ServerConfig) (*JobsDB, error) {

	rawDB, err := jobs.NewDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot open jobs db: %w", err)
	}

	// Create a new jobs db
	jobsDB := &JobsDB{
		DB: rawDB,
	}

	return jobsDB, nil
}

func (j *JobsDB) BuildDailyReports(ctx context.Context, secondsAgo int64) error {
	q := j.Queries

	// Get the Jobs to be summarised
	gotJobs, err := q.GetJobsToBeSummarised(ctx, jobs.GetJobsToBeSummarisedParams{
		SecondsAgo: secondsAgo,
	})
	if err != nil {
		return fmt.Errorf("Error getting jobs to be summarised: %w", err)
	}

	ids := make([]int64, len(gotJobs))
	dayQueueJobMap := make(map[string]jobs.JobStatsDaily)
	for i, j := range gotJobs {
		ids[i] = j.ID

		stats, ok := dayQueueJobMap[j.DayQueueJob()]
		if ok {
			stats.Errors = append(stats.Errors, j.Errors...)
			stats.TotalJobs += 1
			stats.TotalSuccess += j.TotalSuccess()
			stats.TotalFailed += j.TotalFailed()
			stats.DurationSum += j.Duration()
			stats.DurationSquaredSum += j.DurationSquared() // TODO figure this out get better data
			stats.Errors = append(stats.Errors, j.Errors...)
			dayQueueJobMap[j.DayQueueJob()] = stats
		} else {
			dayQueueJobMap[j.DayQueueJob()] = jobs.JobStatsDaily{
				DayQueueJob:        j.DayQueueJob(),
				TotalJobs:          1,
				TotalSuccess:       j.TotalSuccess(),
				TotalFailed:        j.TotalFailed(),
				DurationSum:        j.Duration(),
				DurationSquaredSum: j.DurationSquared(),
				Errors:             j.Errors,
			}
		}
	}

	for k := range dayQueueJobMap {
		_, err := q.EnsureJobStatsDailyForGroup(ctx, jobs.EnsureJobStatsDailyForGroupParams{
			DayQueueJob: k,
		})
		if err != nil {
			return fmt.Errorf("Error ensuring job stats daily for group: %w", err)
		}

	}

	tx, err := j.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error starting transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := j.Queries.WithTx(tx)

	for _, stats := range dayQueueJobMap {
		err = qtx.UpdateJobStatsDaily(ctx, jobs.UpdateJobStatsDailyParams{
			DayQueueJob:        stats.DayQueueJob,
			TotalJobs:          stats.TotalJobs,
			TotalSuccess:       stats.TotalSuccess,
			TotalFailed:        stats.TotalFailed,
			DurationSum:        stats.DurationSum,
			DurationSquaredSum: stats.DurationSquaredSum,
			Errors:             stats.Errors,
		})
		if err != nil {
			return fmt.Errorf("Error updating job stats daily: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("Error updating job stats daily: %w", err)
	}

	err = qtx.DeleteJobsByIds(ctx, jobs.DeleteJobsByIdsParams{
		Ids: ids,
	})
	if err != nil {
		return fmt.Errorf("Error deleting old jobs: %w", err)
	}

	return tx.Commit()
}
