package liteq_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// Many distinct oversized job errors must still summarise into job_stats_daily (CHECK 51200)
// and the summarised jobs must be deleted, or the jobs table grows forever
func TestOversizedJobErrorsStillSummariseAndDeleteJobs(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, jobsDB, queue := t.Context(), controlDB.Jobs, liteq.New(controlDB)
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
		require.NotNil(t, j.RunID, "a failed run keeps its job_id, for its logs")
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
		// Only the newest failed runs are kept, as the oldest runs' logs are trimmed first
		require.Len(t, stats[0].FailedRunIdsJson, errors.MaxErrors)
		require.Equal(t, *failed[count-1].RunID, stats[0].FailedRunIdsJson[errors.MaxErrors-1])
		runRound()
		failed = toSummarise()
	}
	stats, err := jobsDB.Queries.AllJobStatsDaily(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2*count), stats[0].TotalFailed)
}

// A queued job's logs say which queue and job row it is, with an ID for the run
func TestQueuedJobLogsItsIDs(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, queue := t.Context(), liteq.New(controlDB)
	var logs bytes.Buffer
	testutils.UseLogHandler(t, aslog.ContextHandler{Handler: slog.NewTextHandler(&logs, nil)})
	require.NoError(t, queue.QueueJob(ctx, liteq.QueueJobParams{Queue: "logged", Job: "fetch"}))

	var queuedID int64
	worker := func(ctx context.Context, job *jobsdb.Job) error {
		queuedID = job.ID
		slog.InfoContext(ctx, "in the job")
		return nil
	}
	require.NoError(t, liteq.NewConsumerSupervisor(queue, "logged", 1, worker).RunQueued(ctx))
	require.Regexp(t, fmt.Sprintf(`msg="in the job" job_id=[0-9a-f]{16} job=logged queued_job_id=%d`, queuedID), logs.String())
}

// A failed queued job keeps the job_id its run's logs have, so the control site can link to them
func TestFailedQueuedJobKeepsItsRunID(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, queue := t.Context(), liteq.New(controlDB)
	var logs bytes.Buffer
	testutils.UseLogHandler(t, aslog.ContextHandler{Handler: slog.NewTextHandler(&logs, nil)})
	require.NoError(t, queue.QueueJob(ctx, liteq.QueueJobParams{Queue: "failing", Job: "fetch"}))

	var queuedID int64
	worker := func(ctx context.Context, job *jobsdb.Job) error {
		queuedID = job.ID
		slog.ErrorContext(ctx, "upstream is down")
		return fmt.Errorf("upstream is down")
	}
	require.NoError(t, liteq.NewConsumerSupervisor(queue, "failing", 1, worker).RunQueued(ctx))
	job, err := controlDB.Jobs.Queries.FindJob(ctx, jobsdb.FindJobParams{ID: queuedID})
	require.NoError(t, err)
	require.Equal(t, "failed", job.JobStatus)
	require.NotNil(t, job.RunID)
	require.Contains(t, logs.String(), `msg="upstream is down" job_id=`+*job.RunID)
}
