package main

import (
	"bytes"
	"context"
	"database/sql"
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

// A server told to stop while it opens its sites stops there: it does not turn S3 off and page the admin because its
// replication, which a stopped server does not start, could not start
func TestOpenControlDBStopsWhenToldToStop(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	cfg.Status.S3Working = true
	ctx, stop := context.WithCancel(t.Context())
	stop()
	noPage := func(_ context.Context, _ page.Level, message string, _ ...any) { t.Fatalf("paged: %s", message) }

	_, err := openControlDB(ctx, cfg, s3, replica.New(cfg, s3), noPage)
	require.ErrorIs(t, err, context.Canceled)
	require.True(t, cfg.Status.S3Working)
	require.Empty(t, cfg.Status.S3Off)
	pageFailure(noPage, false, err)
}

// A write-ahead log left where the control DB was, as when only the DB's file is moved away after the server was
// killed, is not written over the control DB restored from its replica
func TestRestoreControlDBLeavesOutALeftoverLog(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "replicat", BackupKeyHash: "hash"})
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, control.Close())
	path := controldb.GetPath(cfg)
	require.NoError(t, os.Remove(path))

	// Another control DB, not replicated, whose site is only in its log, as it is still open
	other := *cfg
	other.DataDir = t.TempDir()
	conflict, err := controldb.NewDB(t.Context(), &other)
	require.NoError(t, err)
	defer conflict.Close()
	_, err = conflict.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "conflict", BackupKeyHash: "hash"})
	require.NoError(t, err)
	for _, suffix := range []string{"-wal", "-shm"} {
		left, err := os.ReadFile(conflict.Path + suffix)
		require.NoError(t, err)
		require.NotEmpty(t, left)
		require.NoError(t, os.WriteFile(path+suffix, left, 0o600))
	}

	_, here := logged(t, cfg, replica.New(cfg, s3))
	require.True(t, here)
	restored, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	defer restored.Close()
	_, err = restored.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: "replicat"})
	require.NoError(t, err, "the replica's site is in the restored control DB")
	_, err = restored.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: "conflict"})
	require.ErrorIs(t, err, sql.ErrNoRows, "the leftover log's site is not")
}
