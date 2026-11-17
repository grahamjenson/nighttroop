package restore_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/restore"
	"nighttroop.t1d.nz/utils/s3utils"
)

// testConfig is the test server config, with its key sets, made once as it parses the command line's flags
var testConfig = sync.OnceValues(config.NewServerConfig)

// fakeS3 serves the backups in objects by key, and NoSuchKey for any other key
type fakeS3 struct {
	mu      sync.Mutex
	objects map[string][]byte
}

// newClient is an S3 client for a fresh data folder, talking to a fake S3
func newClient(t *testing.T) (*s3utils.S3Client, *fakeS3) {
	cfg, err := testConfig()
	require.NoError(t, err)
	fake := &fakeS3{objects: map[string][]byte{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		body, ok := fake.objects[strings.TrimPrefix(r.URL.Path, "/"+cfg.S3Bucket+"/")]
		fake.mu.Unlock()
		if !ok {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`))
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	serverConfig := *cfg
	serverConfig.DataDir = t.TempDir()
	serverConfig.S3URL = server.URL
	serverConfig.Status.S3Working = true
	return s3utils.NewS3Client(&serverConfig), fake
}

// A start removes the staging folders a restore or archive killed part way left beside the DBs, and nothing else
func TestRemoveStagingRemovesOnlyStagingFolders(t *testing.T) {
	c, _ := newClient(t)
	cfg := c.ServerConfig
	write := func(path string) string {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("x"), 0o600))
		return path
	}
	dbDir, batch := cfg.DBDir(), sitedb.GetBatchFolder("a", cfg)
	backupBatch := filepath.Dir(sitedb.GetBackupPath("abcdef", cfg))

	// Made as the download and the archive make them, so the names are theirs
	staging := []string{}
	for dir, pattern := range map[string]string{
		filepath.Dir(dbDir): "." + filepath.Base(dbDir) + "-restore-",
		filepath.Dir(batch): "." + filepath.Base(batch) + "-restore-",
		backupBatch:         ".archive-*",
	} {
		require.NoError(t, os.MkdirAll(dir, 0o755))
		folder, err := os.MkdirTemp(dir, pattern)
		require.NoError(t, err)
		write(filepath.Join(folder, "abcdef.sqlite"))
		staging = append(staging, folder)
	}
	kept := []string{
		write(controldb.GetPath(cfg)),
		write(sitedb.GetPath("abcdef", cfg)),
		write(sitedb.GetBackupPath("abcdef", cfg)),
		write(filepath.Join(cfg.DataDir, ".env")),
		write(filepath.Join(filepath.Dir(batch), ".a-restore")),
		write(filepath.Join(backupBatch, ".archived", "abcdef.sqlite")),
	}

	restore.RemoveStaging(cfg)

	for _, folder := range staging {
		require.NoDirExists(t, folder)
	}
	for _, path := range kept {
		require.FileExists(t, path)
	}
}
