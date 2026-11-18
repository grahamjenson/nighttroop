package replica_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"os"
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

// replicated is a replica that a control DB with the site "replicat" was replicated to, and that DB's path
func replicated(t *testing.T) (*replica.Replica, string) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	r := replica.New(cfg, s3)
	require.NoError(t, r.Start(t.Context(), control.Path))
	_, err = control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: "replicat", BackupKeyHash: "hash"})
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	return r, control.Path
}

// another is a DB at a new path that the replica is not of, as one made while S3 was off
func another(t *testing.T) string {
	path := filepath.Join(t.TempDir(), "controldb.sqlite")
	require.NoError(t, os.WriteFile(path, []byte("another DB"), 0o600))
	require.NoError(t, os.WriteFile(path+"-wal", []byte("its log"), 0o600))
	return path
}

func neverSetUp() (bool, error) { return false, nil }
func setUp() (bool, error)      { return true, nil }

// A control DB is its replica's once it has replicated to it or was restored from it, and a DB made apart from it is
// not, nor is one that belongs to a replica somewhere else
func TestReplicaOwnsTheDBItIsOf(t *testing.T) {
	r, path := replicated(t)
	require.True(t, r.Owns(path), "the DB that replicated to it")
	replicate, err := r.Adopt(t.Context(), path, func() (bool, error) { t.Fatal("its own DB is not checked"); return false, nil })
	require.NoError(t, err)
	require.True(t, replicate)

	restored := filepath.Join(t.TempDir(), "controldb.sqlite")
	found, err := r.Restore(t.Context(), restored)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, r.Owns(restored), "the DB restored from it")

	apart := another(t)
	require.False(t, r.Owns(apart), "a DB made apart from it")
	require.NoError(t, os.WriteFile(apart+".replica", []byte("https://s3.example.com/other/host/litestream/control"), 0o600))
	require.False(t, r.Owns(apart), "a DB of a replica somewhere else")
}

// A control DB never set up, made while S3 was off, is not replicated over the replica: it is moved aside, with its
// log, and the replica restored in its place
func TestAdoptRestoresOverADBNeverSetUp(t *testing.T) {
	r, _ := replicated(t)
	path := another(t)
	replicate, err := r.Adopt(t.Context(), path, neverSetUp)
	require.NoError(t, err)
	require.True(t, replicate)
	require.True(t, r.Owns(path))
	require.NoFileExists(t, path+"-wal", "its log is not left beside the restored DB")
	require.NoFileExists(t, path+".restoring")

	db, err := controldb.OpenSqliteDB(path, 0)
	require.NoError(t, err)
	defer db.Close()
	_, err = controldb.New(db).GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: "replicat"})
	require.NoError(t, err, "the replica's DB is in its place")
	aside, err := filepath.Glob(path + ".replaced-*")
	require.NoError(t, err)
	require.Len(t, aside, 2, "the DB and its log are kept")
}

// A control DB that was set up apart from the replica is left as it is and not replicated, so neither is lost
func TestAdoptLeavesADBSetUpApart(t *testing.T) {
	r, _ := replicated(t)
	path := another(t)
	replicate, err := r.Adopt(t.Context(), path, setUp)
	require.NoError(t, err)
	require.False(t, replicate)
	require.False(t, r.Owns(path))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "another DB", string(data))
	require.NoFileExists(t, path+".restoring")
}

// With no replica in S3, as on a server that gets S3 for the first time, the DB here is the replica's first
func TestAdoptWithoutAReplica(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	r := replica.New(cfg, s3)
	path := another(t)
	replicate, err := r.Adopt(t.Context(), path, func() (bool, error) { t.Fatal("there is no replica to choose from"); return false, nil })
	require.NoError(t, err)
	require.True(t, replicate)
	require.True(t, r.Owns(path))

	var off *replica.Replica
	replicate, err = off.Adopt(t.Context(), path, setUp)
	require.NoError(t, err)
	require.True(t, replicate, "without S3 there is nothing to check")
}
