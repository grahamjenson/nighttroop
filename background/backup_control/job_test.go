package backup_control_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_control"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestBackupControlRecordsEvent(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	ctx := context.TODO()
	jobs := backup_control.SetupBackupControlJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupToS3(ctx))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "control_backed_up", events[0].Name)
	require.Contains(t, *events[0].Attributes, `"uploaded":false`)
	// Recorded in a defer, it started when the backup did and finished when the backup returned
	require.GreaterOrEqual(t, events[0].StartedAt, started)
	require.NotNil(t, events[0].FinishedAt)
	require.GreaterOrEqual(t, *events[0].FinishedAt, events[0].StartedAt)
}
