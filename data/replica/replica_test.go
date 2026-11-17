package replica_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// replicated is a site's DB and the control DB, in files, replicated to S3 in memory, and the site's replica's files
func replicated(t *testing.T, siteID string) (*replica.Replicas, *sitedb.DB, func() []string) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	control, err := controldb.NewDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	site, err := sitedb.NewDB(t.Context(), siteID, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { site.Close() })
	replicas := replica.New(cfg, s3)
	require.NoError(t, replicas.Start(t.Context(), control.Path, map[string]string{site.Path: replica.SiteName(siteID)}))
	t.Cleanup(func() { replicas.Close(context.Background()) })
	return replicas, site, func() []string { return testutils.ReplicaFiles(t, cfg, s3, replica.SiteName(siteID)) }
}

func addReading(t *testing.T, site *sitedb.DB, id string) {
	_, err := site.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
}

// A site's writes, once synced, are in its replica in S3, and restore to a DB with them
func TestSiteWritesReachItsReplica(t *testing.T) {
	replicas, site, files := replicated(t, "repsitea")
	addReading(t, site, "reading-in-the-replica")
	require.NoError(t, replicas.Sync(t.Context(), site.Path))

	require.NotEmpty(t, files())
	require.NotNil(t, replicas.Replicated(site.Path))

	restored := filepath.Join(t.TempDir(), "restored.sqlite")
	found, err := replicas.Restore(t.Context(), replica.SiteName("repsitea"), restored)
	require.NoError(t, err)
	require.True(t, found)
	db, err := sitedb.OpenSqliteDB(restored, 0)
	require.NoError(t, err)
	defer db.Close()
	_, err = sitedb.New(db).GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: "reading-in-the-replica"})
	require.NoError(t, err)

	// A site with no replica is not found
	found, err = replicas.Restore(t.Context(), replica.SiteName("norepsit"), filepath.Join(t.TempDir(), "none.sqlite"))
	require.NoError(t, err)
	require.False(t, found)
}

// Closing sends a DB's last writes, also those made before Litestream first read it, a sync interval after start, as
// when the server stops soon after it starts
func TestCloseSendsTheLastWrites(t *testing.T) {
	replicas, site, _ := replicated(t, "repsitee")
	addReading(t, site, "last")
	require.NoError(t, replicas.Close(t.Context()))

	restored := filepath.Join(t.TempDir(), "restored.sqlite")
	found, err := replicas.Restore(t.Context(), replica.SiteName("repsitee"), restored)
	require.NoError(t, err)
	require.True(t, found)
	db, err := sitedb.OpenSqliteDB(restored, 0)
	require.NoError(t, err)
	defer db.Close()
	_, err = sitedb.New(db).GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: "last"})
	require.NoError(t, err)
}

// Over HTTPS each DB's replica is encrypted by S3 with a key of its own (SSE-C), and over HTTP, as the development
// MinIO is, not at all, as S3 refuses SSE-C over HTTP
func TestReplicasAreEncryptedOverHTTPS(t *testing.T) {
	t.Setenv("AWS_CA_BUNDLE", "")
	cfg := *testutils.GetTestServerConfig()
	cfg.S3URL = "https://s3.example.com"
	bucket := s3utils.NewS3Client(&cfg)
	keys := map[string]bool{}
	for _, name := range []string{"control", replica.SiteName("repsitef"), replica.SiteName("repsiteg")} {
		client := replica.Client(&cfg, bucket, name, true)
		require.Equal(t, "AES256", client.SSECustomerAlgorithm)
		key, err := base64.StdEncoding.DecodeString(client.SSECustomerKey)
		require.NoError(t, err)
		require.Len(t, key, 32)
		require.Equal(t, "localhost/litestream/"+name, client.Path)
		keys[client.SSECustomerKey] = true
		require.NoError(t, client.Init(t.Context()), "the client takes the key")
	}
	require.Len(t, keys, 3, "each DB has its own key")

	client := replica.Client(&cfg, bucket, "control", false)
	require.Empty(t, client.SSECustomerKey)
}

// A site archived stops replicating, keeping its replica, and one deleted deletes it, each removing what Litestream kept
// beside its DB
func TestRemoveStopsReplicating(t *testing.T) {
	replicas, site, files := replicated(t, "repsited")
	addReading(t, site, "first")
	require.NoError(t, replicas.Sync(t.Context(), site.Path))
	meta := filepath.Join(filepath.Dir(site.Path), ".repsited.sqlite-litestream")
	require.DirExists(t, meta)

	require.NoError(t, replicas.Remove(t.Context(), site.Path, replica.SiteName("repsited"), false))
	require.NoDirExists(t, meta)
	kept := files()
	require.NotEmpty(t, kept)
	require.Nil(t, replicas.Replicated(site.Path))
	require.Error(t, replicas.Sync(t.Context(), site.Path), "it is no longer replicated")
	addReading(t, site, "after")
	require.Equal(t, kept, files(), "nothing more is sent")

	require.NoError(t, replicas.Add(t.Context(), site.Path, replica.SiteName("repsited")))
	require.NoError(t, replicas.Remove(t.Context(), site.Path, replica.SiteName("repsited"), true))
	require.Empty(t, files())
}

// Without S3 nothing is replicated or restored, and nothing is kept beside a DB
func TestNoReplicasWithoutS3(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	require.False(t, cfg.Status.S3Working)
	replicas := replica.New(&cfg, s3utils.NewS3Client(&cfg))
	require.Nil(t, replicas)
	path := filepath.Join(t.TempDir(), "site.sqlite")
	require.NoError(t, replicas.Start(t.Context(), path, map[string]string{path: "site/nosssite"}))
	require.NoError(t, replicas.Add(t.Context(), path, "site/nosssite"))
	require.NoError(t, replicas.Sync(t.Context(), path))
	found, err := replicas.Restore(t.Context(), "site/nosssite", path)
	require.NoError(t, err)
	require.False(t, found)
	require.NoError(t, replicas.Remove(t.Context(), path, "site/nosssite", true))
	require.NoError(t, replicas.Close(t.Context()))
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}
