package truncate

import (
	"context"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/control"
)

// TODO this is the background runner that will go over all DBs including control, and truncate to size.
// This should stop abusive users, but also provide a limit.
// A possibility is to limit per user, put the limits in the control/db so that they can be changed if needed
// Truncate shouldnt be based on dates, but on row counts,

// Loop is
// Go over each user DB, check size of DB, if over limit, execute truncation queries on every table to bring back in line
// Then check that the DB has reduced in size, though this might be difficult with the WAL
// TODO write an event for that

// Control DB, check size is not over some limit, execute truncations queries if it is.

func (t *TruncateJobManager) TruncateUsers() error {
	slog.Info("TruncateUsers job unimplemented")
	return fmt.Errorf("Not implemented")
}

// TODO: This whole function could be rewritten in SQL and made cleaner, but it will be harder to maintain
func (t *TruncateJobManager) TruncateJobs() error {
	ctx := context.TODO()
	q := t.ControlDB.Queries

	// Get the Jobs to be summarised
	jobs, err := q.GetJobsToBeSummarised(ctx)
	if err != nil {
		return fmt.Errorf("Error getting jobs to be summarised: %w", err)
	}

	ids := make([]int64, len(jobs))
	dayQueueJobMap := make(map[string]control.JobStatsDaily)
	for i, j := range jobs {
		ids[i] = j.ID

		stats, ok := dayQueueJobMap[j.DayQueueJob()]
		if ok {
			stats.Errors = append(stats.Errors, j.Errors...)
			stats.TotalJobs += 1
			stats.TotalSuccess += j.TotalSuccess()
			stats.TotalFailed += j.TotalFailed()
			stats.DurationSum += j.Duration()
			stats.DurationSquaredSum += j.DurationSquared()
			stats.Errors = append(stats.Errors, j.Errors...)
			dayQueueJobMap[j.DayQueueJob()] = stats
		} else {
			dayQueueJobMap[j.DayQueueJob()] = control.JobStatsDaily{
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

	for k, _ := range dayQueueJobMap {
		_, err := q.EnsureJobStatsDailyForGroup(ctx, k)
		if err != nil {
			return fmt.Errorf("Error ensuring job stats daily for group: %w", err)
		}

	}

	tx, err := t.ControlDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error starting transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := t.ControlDB.Queries.WithTx(tx)

	for _, stats := range dayQueueJobMap {
		err = qtx.UpdateJobStatsDaily(ctx, control.UpdateJobStatsDailyParams{
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

	err = qtx.DeleteJobsByIds(ctx, ids)
	if err != nil {
		return fmt.Errorf("Error deleting old jobs: %w", err)
	}

	return tx.Commit()
}
