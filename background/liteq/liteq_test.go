package liteq_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// Many distinct oversized job errors must still summarise into job_stats_daily (CHECK 51200)
// and the summarised jobs must be deleted, or the jobs table grows forever
func TestOversizedJobErrorsStillSummariseAndDeleteJobs(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, jobsDB, queue := context.TODO(), controlDB.Jobs, liteq.New(controlDB)
	const queueName, count = "long-errors", 20

	worker := func(ctx context.Context, job *jobsdb.Job) error {
		return fmt.Errorf("upstream for job %d said: %s", job.ID, strings.Repeat("x", 10*1024))
	}
	runRound := func() {
		for range count {
			require.NoError(t, queue.QueueJob(ctx, liteq.QueueJobParams{Queue: queueName, Job: "fetch"}))
		}
		require.NoError(t, liteq.NewConsumerSupervisor(queue, queueName, 1, worker).RunQueued(ctx))
	}
	toSummarise := func() []*jobsdb.Job {
		jobs, err := jobsDB.Queries.GetJobsToBeSummarised(ctx, jobsdb.GetJobsToBeSummarisedParams{SecondsAgo: 0})
		require.NoError(t, err)
		return jobs
	}

	runRound()
	failed := toSummarise()
	require.Len(t, failed, count)
	for _, j := range failed {
		require.Equal(t, "failed", j.JobStatus)
		require.Len(t, j.Errors, 1)
		require.LessOrEqual(t, utf8.RuneCountInString(j.Errors[0]), errors.MaxErrorLength+1)
	}

	// Summarising and a second round into the same day both stay within the bounds
	for range 2 {
		require.NoError(t, jobsDB.BuildDailyReports(ctx, 0))
		require.Empty(t, toSummarise(), "summarised jobs are deleted")
		stats, err := jobsDB.Queries.AllJobStatsDaily(ctx)
		require.NoError(t, err)
		require.Len(t, stats, 1)
		require.Len(t, stats[0].Errors, errors.MaxErrors)
		runRound()
	}
	stats, err := jobsDB.Queries.AllJobStatsDaily(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2*count), stats[0].TotalFailed)
}
