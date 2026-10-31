package cron_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestAddFunc checks a job that returns, errors or panics is run and kept as completed or failed, then summarised
// into the day's stats with its error
func TestAddFunc(t *testing.T) {
	tests := []struct {
		name    string
		job     func(context.Context) error
		status  string
		success int64
		failed  int64
		err     string
	}{
		{"returns", func(context.Context) error { return nil }, "completed", 1, 0, ""},
		{"errors", func(context.Context) error { return fmt.Errorf("test error") }, "failed", 0, 1, "test error"},
		{"panics", func(context.Context) error { panic("test panic") }, "failed", 0, 1, "test panic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			jobsDB, err := db.NewJobsDB(ctx, testutils.GetTestServerConfig())
			require.NoError(t, err)
			testutils.DumpJobsOnFailure(t, jobsDB, "TestAddFunc")
			toBeSummarised := func(secondsAgo int64) []*jobsdb.Job {
				jobs, err := jobsDB.Queries.GetJobsToBeSummarised(ctx, jobsdb.GetJobsToBeSummarisedParams{SecondsAgo: secondsAgo})
				require.NoError(t, err)
				return jobs
			}

			cron := cron.NewCron(jobsDB)
			entryid, err := cron.RawAddFunc("testing", "@every 1s", tt.job)
			require.NoError(t, err)
			require.NotZero(t, entryid)
			require.Len(t, cron.Entries(), 1)
			require.Empty(t, toBeSummarised(0))

			// Running it is caught by the jobs DB, even when it panics
			cron.Entry(entryid).Job.Run()

			// SecondsAgo is seconds: a job from a few milliseconds ago is not a second old yet
			time.Sleep(10 * time.Millisecond)
			require.Empty(t, toBeSummarised(1))
			jobs := toBeSummarised(0)
			require.Len(t, jobs, 1)
			require.Equal(t, tt.status, jobs[0].JobStatus)

			require.NoError(t, jobsDB.BuildDailyReports(ctx, 0))
			jobStats, err := jobsDB.Queries.AllJobStatsDaily(ctx)
			require.NoError(t, err)
			require.Len(t, jobStats, 1)
			require.Equal(t, tt.success, jobStats[0].TotalSuccess)
			require.Equal(t, tt.failed, jobStats[0].TotalFailed)
			if tt.err == "" {
				require.Empty(t, jobStats[0].Errors)
				return
			}
			require.Len(t, jobStats[0].Errors, 1)
			require.Contains(t, jobStats[0].Errors[0], tt.err)
		})
	}
}

func TestAddFuncRunsAgainAfterPanic(t *testing.T) {
	config := testutils.GetTestServerConfig()
	jobsDB, _ := db.NewJobsDB(context.TODO(), config)

	cron := cron.NewCron(jobsDB)
	runs := 0
	entryid, _ := cron.RawAddFunc("testing", "@every 1s", func(context.Context) error { runs++; panic("test panic") })
	entry := cron.Entry(entryid)
	entry.Job.Run()
	// A panic must not leave the job marked as still running
	entry.Job.Run()

	require.Equal(t, 2, runs)
}

// Each run of a job logs the job and an ID of its own
func TestAddFuncLogsEachRunsID(t *testing.T) {
	jobsDB, err := db.NewJobsDB(context.TODO(), testutils.GetTestServerConfig())
	require.NoError(t, err)
	var logs bytes.Buffer
	testutils.UseLogHandler(t, aslog.ContextHandler{Handler: slog.NewTextHandler(&logs, nil)})

	cron := cron.NewCron(jobsDB)
	entryid, err := cron.RawAddFunc("testing", "@every 1s", func(ctx context.Context) error {
		slog.InfoContext(ctx, "in the job")
		return nil
	})
	require.NoError(t, err)
	cron.Entry(entryid).Job.Run()
	cron.Entry(entryid).Job.Run()

	ids := regexp.MustCompile(`msg="in the job" job_id=([0-9a-f]{16}) job=testing`).FindAllStringSubmatch(logs.String(), -1)
	require.Len(t, ids, 2)
	require.NotEqual(t, ids[0][1], ids[1][1])
}
