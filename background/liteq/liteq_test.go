package liteq_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
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

// A job is only fetched when a worker is free to run it. Its visibility timeout counts from when it is fetched, so a
// job fetched to wait behind others would time out, and be queued again, before it ran.
func TestConsumeOnlyFetchesJobsAWorkerIsFreeFor(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, stop := context.WithCancel(t.Context())
	queue := liteq.New(controlDB)
	const queueName, count = "one-at-a-time", 3
	for range count {
		require.NoError(t, queue.QueueJob(ctx, liteq.QueueJobParams{Queue: queueName, Job: "fetch"}))
	}

	started, release := make(chan struct{}), make(chan struct{})
	worker := func(ctx context.Context, job *jobsdb.Job) error {
		select {
		case started <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
		<-release
		return nil
	}
	consumed := make(chan error)
	go func() {
		consumed <- queue.Consume(ctx, liteq.ConsumeParams{Queue: queueName, PoolSize: 1, Worker: worker, OnEmptySleep: time.Millisecond})
	}()

	// Never and Eventually check in a goroutine, and return without waiting for a check still running, so the last
	// one's can run after stop() cancels ctx. A failed read is -1, which fails a check that is waited on, rather than
	// failing the test from that goroutine.
	countStatus := func(status string) int {
		jobs, err := controlDB.Jobs.Queries.SearchJobs(ctx, jobsdb.SearchJobsParams{Limit: count})
		if err != nil {
			return -1
		}
		n := 0
		for _, job := range jobs {
			if job.JobStatus == status {
				n++
			}
		}
		return n
	}
	for ran := range count {
		<-started
		require.Never(t, func() bool { return countStatus("fetched") != 1 || countStatus("queued") != count-ran-1 },
			200*time.Millisecond, 10*time.Millisecond, "while a job runs, the ones behind it stay queued")
		release <- struct{}{}
	}
	require.Eventually(t, func() bool { return countStatus("completed") == count }, 5*time.Second, 10*time.Millisecond)
	stop()
	<-consumed // stopping can cancel a query, so it may return that error, but it does return
}

// Stopping a consumer cancels what it was doing, e.g. resetting jobs, which is not an error to report: servers stop
// whenever they scale to zero
func TestStoppingAConsumerLogsNoError(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	var buf bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(before) })

	// Stopped before it starts, its first query is canceled, unless it sees it is stopped first
	for range 20 {
		consumer := liteq.NewConsumerSupervisor(liteq.New(controlDB), "stopping", 1, func(context.Context, *jobsdb.Job) error { return nil })
		consumer.Stop()
		consumer.Start()
		consumer.Wait()
	}
	require.NotContains(t, buf.String(), "level=ERROR")
}
