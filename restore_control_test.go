package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// logged is what restoreFromS3 logs for cfg's control DB, and whether the control DB is there after
func logged(t *testing.T, cfg *config.ServerConfig, control *replica.Replica) (string, bool) {
	t.Helper()
	var logs bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(before)
	require.NoError(t, restoreFromS3(t.Context(), cfg, control))
	return logs.String(), utils.FileExists(controldb.GetPath(cfg))
}

// A start logs what it does with the control DB: keeps the one here, or restores a missing one from its replica, or
// starts a new one when there is none or S3 is off
func TestRestoreControlDBIsLogged(t *testing.T) {
	t.Run("here", func(t *testing.T) {
		cfg, s3 := testutils.ReplicatedConfig(t)
		control, err := controldb.NewDB(t.Context(), cfg)
		require.NoError(t, err)
		require.NoError(t, control.Close())
		logs, here := logged(t, cfg, replica.New(cfg, s3))
		require.Contains(t, logs, "Control DB is here, so it is not restored")
		require.True(t, here)
	})
	t.Run("S3 off", func(t *testing.T) {
		cfg := *testutils.GetTestServerConfig()
		cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
		logs, here := logged(t, &cfg, nil)
		require.Contains(t, logs, "Control DB is missing and S3 is off, so the server starts with a new control DB")
		require.False(t, here)
	})
	t.Run("no replica", func(t *testing.T) {
		cfg, s3 := testutils.ReplicatedConfig(t)
		logs, here := logged(t, cfg, replica.New(cfg, s3))
		require.Contains(t, logs, "Control DB is missing, so it is restored from its replica in S3")
		require.Contains(t, logs, "Control DB has no replica in S3, so the server starts with a new control DB")
		require.False(t, here)
	})
	t.Run("restored", func(t *testing.T) {
		cfg, s3 := testutils.ReplicatedConfig(t)
		control, err := controldb.NewDB(t.Context(), cfg)
		require.NoError(t, err)
		r := replica.New(cfg, s3)
		require.NoError(t, r.Start(t.Context(), control.Path))
		require.NoError(t, r.Close(context.Background()))
		require.NoError(t, control.Close())
		require.NoError(t, os.Remove(control.Path))
		logs, here := logged(t, cfg, replica.New(cfg, s3))
		require.Contains(t, logs, "Control DB is missing, so it is restored from its replica in S3")
		require.Contains(t, logs, "Control DB restored from its replica in S3")
		require.True(t, here)
	})
}
