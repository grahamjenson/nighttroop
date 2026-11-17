package replica_test

import (
	"bytes"
	"context"
	"io"
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

// replicated is a site's DB and the control DB, in files, replicated to S3 in memory
func replicated(t *testing.T, siteID string) (*replica.Replicas, *sitedb.DB, *s3utils.S3Client) {
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
	return replicas, site, s3
}

// objects are the keys and contents of the site's replica in S3
func objects(t *testing.T, s3 *s3utils.S3Client, siteID string) map[string][]byte {
	found, err := s3.List(t.Context(), s3.ReplicaPrefix()+replica.SiteName(siteID)+"/")
	require.NoError(t, err)
	contents := map[string][]byte{}
	for _, object := range found {
		body, err := s3.Get(t.Context(), *object.Key)
		require.NoError(t, err)
		contents[*object.Key], err = io.ReadAll(body)
		require.NoError(t, err)
		require.NoError(t, body.Close())
	}
	return contents
}

func addReading(t *testing.T, site *sitedb.DB, id string) {
	_, err := site.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
}

// A site's writes, once synced, are in its replica in S3, encrypted, and restore to a DB with them
func TestSiteWritesReachItsReplicaEncrypted(t *testing.T) {
	replicas, site, s3 := replicated(t, "repsitea")
	addReading(t, site, "reading-in-the-replica")
	require.NoError(t, replicas.Sync(t.Context(), site.Path))

	sent := objects(t, s3, "repsitea")
	require.NotEmpty(t, sent)
	for key, content := range sent {
		require.NotContains(t, string(content), "reading-in-the-replica", key)
		require.NotContains(t, string(content), "CREATE TABLE", key)
	}
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

// A file of one site's replica does not decrypt as another's, so a replica cannot be passed off as another site's
func TestReplicaFilesAreBoundToTheirSite(t *testing.T) {
	replicas, site, s3 := replicated(t, "repsiteb")
	addReading(t, site, "kept")
	require.NoError(t, replicas.Sync(t.Context(), site.Path))
	for key, content := range objects(t, s3, "repsiteb") {
		other := bytes.Replace([]byte(key), []byte("site/repsiteb/"), []byte("site/repsitec/"), 1)
		_, err := s3.Put(t.Context(), string(other), bytes.NewReader(content))
		require.NoError(t, err)
	}
	_, err := replicas.Restore(t.Context(), replica.SiteName("repsitec"), filepath.Join(t.TempDir(), "other.sqlite"))
	require.Error(t, err)
}

// A site archived stops replicating, keeping its replica, and one deleted deletes it, each removing what Litestream kept
// beside its DB
func TestRemoveStopsReplicating(t *testing.T) {
	replicas, site, s3 := replicated(t, "repsited")
	addReading(t, site, "first")
	require.NoError(t, replicas.Sync(t.Context(), site.Path))
	meta := filepath.Join(filepath.Dir(site.Path), ".repsited.sqlite-litestream")
	require.DirExists(t, meta)

	require.NoError(t, replicas.Remove(t.Context(), site.Path, replica.SiteName("repsited"), false))
	require.NoDirExists(t, meta)
	kept := objects(t, s3, "repsited")
	require.NotEmpty(t, kept)
	require.Nil(t, replicas.Replicated(site.Path))
	require.Error(t, replicas.Sync(t.Context(), site.Path), "it is no longer replicated")
	addReading(t, site, "after")
	require.Equal(t, kept, objects(t, s3, "repsited"), "nothing more is sent")

	require.NoError(t, replicas.Add(t.Context(), site.Path, replica.SiteName("repsited")))
	require.NoError(t, replicas.Remove(t.Context(), site.Path, replica.SiteName("repsited"), true))
	require.Empty(t, objects(t, s3, "repsited"))
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
