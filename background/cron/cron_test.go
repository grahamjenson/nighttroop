package cron_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestAddFunc(t *testing.T) {

	config := testutils.GetTestServerConfig()
	jobsDB, err := db.NewJobsDB(context.TODO(), config)
	require.NoError(t, err)

	testutils.DumpJobsOnFailure(t, jobsDB, "TestAddFunc")

	cron := cron.NewCron(jobsDB)

	entryid, err := cron.RawAddFunc("testing", "@every 1s", func() error { return nil })
	require.NoError(t, err)
	require.NotZero(t, entryid)

	require.Len(t, cron.Entries(), 1)
	jobsTBS, err := jobsDB.Queries.GetJobsToBeSummarised(context.TODO(), jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: 0,
	})
	require.NoError(t, err)
	require.Len(t, jobsTBS, 0)

	// Run the entry
	entry := cron.Entry(entryid)
	entry.Job.Run()

	// SecondsAgo is seconds: a job from a few milliseconds ago is not a second old yet
	time.Sleep(10 * time.Millisecond)
	jobsTBS, err = jobsDB.Queries.GetJobsToBeSummarised(context.TODO(), jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: 1,
	})
	require.NoError(t, err)
	require.Len(t, jobsTBS, 0)

	jobsTBS, err = jobsDB.Queries.GetJobsToBeSummarised(context.TODO(), jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: 0,
	})
	require.NoError(t, err)
	require.Len(t, jobsTBS, 1)
	require.Equal(t, "completed", jobsTBS[0].JobStatus)

	err = jobsDB.BuildDailyReports(context.TODO(), 0)
	require.NoError(t, err)

	jobStats, err := jobsDB.Queries.AllJobStatsDaily(context.TODO())
	require.NoError(t, err)
	require.Len(t, jobStats, 1)
	require.Equal(t, int64(1), jobStats[0].TotalSuccess)

}

func TestAddFuncPanic(t *testing.T) {
	config := testutils.GetTestServerConfig()
	jobsDB, _ := db.NewJobsDB(context.TODO(), config)

	cron := cron.NewCron(jobsDB)
	entryid, _ := cron.RawAddFunc("testing", "@every 1s", func() error { panic("test panic"); return nil })
	entry := cron.Entry(entryid)
	// Should not panic and be caught by the jobs DB
	entry.Job.Run()

	jobsTBS, err := jobsDB.Queries.GetJobsToBeSummarised(context.TODO(), jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: 0,
	})
	require.NoError(t, err)
	require.Len(t, jobsTBS, 1)
	require.Equal(t, "failed", jobsTBS[0].JobStatus)

	err = jobsDB.BuildDailyReports(context.TODO(), 0)
	require.NoError(t, err)

	jobStats, err := jobsDB.Queries.AllJobStatsDaily(context.TODO())
	require.NoError(t, err)
	require.Len(t, jobStats, 1)
	require.Equal(t, int64(1), jobStats[0].TotalFailed)

	require.Len(t, jobStats[0].Errors, 1)
	require.Contains(t, jobStats[0].Errors[0], "test panic")
}

func TestAddFuncError(t *testing.T) {
	config := testutils.GetTestServerConfig()
	jobsDB, _ := db.NewJobsDB(context.TODO(), config)

	cron := cron.NewCron(jobsDB)
	entryid, _ := cron.RawAddFunc("testing", "@every 1s", func() error { return fmt.Errorf("test error") })
	entry := cron.Entry(entryid)
	// Simulate an error in the job
	entry.Job.Run()

	jobsTBS, err := jobsDB.Queries.GetJobsToBeSummarised(context.TODO(), jobsdb.GetJobsToBeSummarisedParams{
		SecondsAgo: 0,
	})
	require.NoError(t, err)
	require.Len(t, jobsTBS, 1)
	require.Equal(t, "failed", jobsTBS[0].JobStatus)

	err = jobsDB.BuildDailyReports(context.TODO(), 0)
	require.NoError(t, err)

	jobStats, err := jobsDB.Queries.AllJobStatsDaily(context.TODO())
	require.NoError(t, err)
	require.Len(t, jobStats, 1)
	require.Equal(t, int64(1), jobStats[0].TotalFailed)
	require.Len(t, jobStats[0].Errors, 1)
	require.Contains(t, jobStats[0].Errors[0], "test error")
}

func TestAddFuncRunsAgainAfterPanic(t *testing.T) {
	config := testutils.GetTestServerConfig()
	jobsDB, _ := db.NewJobsDB(context.TODO(), config)

	cron := cron.NewCron(jobsDB)
	runs := 0
	entryid, _ := cron.RawAddFunc("testing", "@every 1s", func() error { runs++; panic("test panic") })
	entry := cron.Entry(entryid)
	entry.Job.Run()
	// A panic must not leave the job marked as still running
	entry.Job.Run()

	require.Equal(t, 2, runs)
}
