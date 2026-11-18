package replica_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// The control DB's writes reach its replica, also those made just before it closes, before Litestream first read it,
// and restore to a DB with them
func TestControlWritesReachItsReplica(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "replicat", BackupKeyHash: "hash"})
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	require.NotEmpty(t, testutils.ControlReplicaFiles(t, cfg, s3))

	restored := filepath.Join(t.TempDir(), "controldb.sqlite")
	found, err := r.Restore(t.Context(), restored)
	require.NoError(t, err)
	require.True(t, found)
	db, err := controldb.OpenSqliteDB(restored, 0)
	require.NoError(t, err)
	defer db.Close()
	_, err = controldb.New(db).GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: "replicat"})
	require.NoError(t, err)
}

// Replicating logs none of Litestream's routine logs, such as each second's sync, which the server's logs would be full of
func TestReplicatingLogsNoRoutineLogs(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	// After the test config, which sets the default logger as the server does
	var logs bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(before) })
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "quietsit", BackupKeyHash: "hash"})
	require.NoError(t, err)
	// Litestream's own sync, each second, sends it
	require.Eventually(t, func() bool { return len(testutils.ControlReplicaFiles(t, cfg, s3)) > 0 }, 10*time.Second, 100*time.Millisecond)
	require.NoError(t, r.Close(context.Background()))
	require.NotContains(t, logs.String(), "system=store")
}

// A control DB with no replica, as on a new server, is not found
func TestRestoreWithoutAReplica(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	found, err := replica.New(cfg, s3).Restore(t.Context(), filepath.Join(t.TempDir(), "controldb.sqlite"))
	require.NoError(t, err)
	require.False(t, found)
}

// Over HTTPS S3 encrypts the replica with the server's key (SSE-C), and over HTTP, as the development MinIO is, not
// at all, as S3 refuses SSE-C over HTTP
func TestReplicaIsEncryptedOverHTTPS(t *testing.T) {
	t.Setenv("AWS_CA_BUNDLE", "")
	cfg := *testutils.GetTestServerConfig()
	cfg.S3URL = "https://s3.example.com"
	bucket := s3utils.NewS3Client(&cfg)
	client := replica.Client(&cfg, bucket, true)
	require.Equal(t, "AES256", client.SSECustomerAlgorithm)
	key, err := base64.StdEncoding.DecodeString(client.SSECustomerKey)
	require.NoError(t, err)
	require.Len(t, key, 32)
	require.Equal(t, "localhost/litestream/control", client.Path)
	require.NoError(t, client.Init(t.Context()), "the client takes the key")

	require.Empty(t, replica.Client(&cfg, bucket, false).SSECustomerKey)
}

// Without S3 nothing is replicated or restored
func TestNoReplicaWithoutS3(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	require.False(t, cfg.Status.S3Working)
	r := replica.New(&cfg, s3utils.NewS3Client(&cfg))
	require.Nil(t, r)
	path := filepath.Join(t.TempDir(), "controldb.sqlite")
	require.NoError(t, r.Start(t.Context(), path))
	found, err := r.Restore(t.Context(), path)
	require.NoError(t, err)
	require.False(t, found)
	require.NoError(t, r.Close(t.Context()))
	require.NoFileExists(t, path)
}

// A replica that is abandoned, as when another server took this server's S3 lock, sends nothing more, not the writes
// made before it nor those after, and closing it then does nothing
func TestAbandonSendsNothingMore(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "replicat", BackupKeyHash: "hash"})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return len(testutils.ControlReplicaFiles(t, cfg, s3)) > 0 }, 10*time.Second, 100*time.Millisecond)

	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "notsent1", BackupKeyHash: "hash"})
	require.NoError(t, err)
	r.Abandon()
	sent := testutils.ControlReplicaFiles(t, cfg, s3)
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "notsent2", BackupKeyHash: "hash"})
	require.NoError(t, err)
	time.Sleep(2 * time.Second) // two of Litestream's syncs
	require.NoError(t, r.Close(context.Background()))
	require.Equal(t, sent, testutils.ControlReplicaFiles(t, cfg, s3))

	var off *replica.Replica
	off.Abandon()
}
