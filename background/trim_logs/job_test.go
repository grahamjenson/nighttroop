package trim_logs_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/trim_logs"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, to run them straight away
type scheduler map[string]func(context.Context) error

func (s scheduler) AddFunc(name string, spec string, cmd func(context.Context) error) { s[name] = cmd }
func (s scheduler) AddFuncSoon(name string, _, _ time.Duration, cmd func(context.Context) error) {
	s[name] = cmd
}

func TestTrimLogsDBCron(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	ctx, logs := t.Context(), controlDB.Logs
	cfg.LogsDBMaxRowsPerTable = 2
	for _, message := range []string{"first", "second", "third"} {
		logs.Log(ctx, nil, db.LogLevelInfo, message, nil)
	}

	jobs := scheduler{}
	trim_logs.SetupCron(jobs, cfg, logs)
	require.NoError(t, jobs["TrimLogsDB"](ctx))

	rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "third", rows[0].Message)
	require.Equal(t, "second", rows[1].Message)
}

// The requests refused before signing in are summed in the logs each minute, by a job of its own
func TestSumRefusedRequestsCron(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	jobs := scheduler{}
	trim_logs.SetupCron(jobs, cfg, controlDB.Logs)
	require.Contains(t, jobs, "SumRefusedRequests")
	require.NoError(t, jobs["SumRefusedRequests"](t.Context()))
}
