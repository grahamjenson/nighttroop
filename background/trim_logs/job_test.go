package trim_logs_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/trim_logs"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// scheduler keeps the jobs added to it, to run them straight away
type scheduler map[string]func() error

func (s scheduler) AddFunc(name string, spec string, cmd func() error) { s[name] = cmd }

func TestTrimLogsDBCron(t *testing.T) {
	cfg, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs := context.TODO(), controlDB.Logs
	cfg.LogsDBMaxRowsPerTable = 2
	for _, message := range []string{"first", "second", "third"} {
		logs.Log(ctx, nil, db.LogLevelInfo, message, nil)
	}

	jobs := scheduler{}
	trim_logs.SetupCron(jobs, cfg, logs)
	require.NoError(t, jobs["TrimLogsDB"]())

	rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "third", rows[0].Message)
	require.Equal(t, "second", rows[1].Message)
}
