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
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/testutils"
)

// logged is what restoreFromS3 logs for cfg's control DB, and whether the control DB is there after
func logged(t *testing.T, cfg *config.ServerConfig, control *replica.Replica) (string, bool) {
	t.Helper()
	var logs bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(before)
	noPage := func(_ context.Context, _ page.Level, message string, _ ...any) { t.Fatalf("paged: %s", message) }
	require.NoError(t, restoreFromS3(t.Context(), cfg, control, noPage))
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

// A control DB that does not open, as a file a disk damaged, is restored from its replica and the admin is paged; one
// that opens is left, and pages no one
func TestRestoreUnopenedControlDB(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "replicat", BackupKeyHash: "hash"})
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, control.Close())
	pages := []string{}
	admin := func(_ context.Context, level page.Level, message string, _ ...any) {
		pages = append(pages, string(level)+": "+message)
	}

	require.NoError(t, restoreUnopened(t.Context(), cfg, r, admin))
	require.Empty(t, pages, "a control DB that opens is left")

	path := controldb.GetPath(cfg)
	for _, file := range []string{path, path + "-wal", path + "-shm"} {
		_ = os.Remove(file)
	}
	require.NoError(t, os.WriteFile(path, []byte("not a DB, as a disk damaged it"), 0o600))
	require.NoError(t, restoreUnopened(t.Context(), cfg, r, admin))
	require.Len(t, pages, 1)
	require.Contains(t, pages[0], string(page.SMS)+": The control DB did not open, so it was restored from its replica in S3")
	restored, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err, "the restored control DB opens")
	defer restored.Close()
	_, err = restored.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: "replicat"})
	require.NoError(t, err)
}
