package backup_control_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_control"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestBackupControlRecordsEvent(t *testing.T) {
	cfg, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx := context.TODO()
	jobs := backup_control.SetupBackupControlJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})

	require.NoError(t, jobs.BackupToS3())

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "control_backed_up", events[0].Name)
	require.Contains(t, *events[0].Attributes, `"uploaded":false`)
}
