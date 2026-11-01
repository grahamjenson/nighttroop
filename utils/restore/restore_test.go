package restore_test

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/pierrec/lz4/v4"
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

func (f *fakeS3) set(key string, body []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = body
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

// backup is a file, too big for one encrypted segment, packed as a backup, and that file's content
func backup(t *testing.T, c *s3utils.S3Client, name string) ([]byte, []byte) {
	content := make([]byte, 2<<20)
	_, err := rand.Read(content)
	require.NoError(t, err)

	var archive bytes.Buffer
	enc, err := c.ServerConfig.Security.EncryptingWriter(&archive)
	require.NoError(t, err)
	// Small blocks, as a real DB is many blocks, so a cut off download has already written some of the file
	lz := lz4.NewWriter(enc)
	require.NoError(t, lz.Apply(lz4.BlockSizeOption(lz4.Block64Kb)))
	tw := tar.NewWriter(lz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, lz.Close())
	require.NoError(t, enc.Close())
	return archive.Bytes(), content
}

// A control DB download that is cut off leaves no control DB, so the next start downloads it again
func TestCutOffControlDownloadLeavesNoControlDB(t *testing.T) {
	c, fake := newClient(t)
	key, path := controldb.GetS3Path(c.ServerConfig), controldb.GetPath(c.ServerConfig)
	archive, content := backup(t, c, filepath.Base(path))

	fake.set(key, archive[:len(archive)/2])
	require.Error(t, restore.TryDownloadExtractControlDbs(c))
	require.NoFileExists(t, path)

	fake.set(key, archive)
	require.NoError(t, restore.TryDownloadExtractControlDbs(c))
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, got)

	// Nothing is left beside the DBs from either download
	entries, err := os.ReadDir(c.ServerConfig.DataDir)
	require.NoError(t, err)
	for _, entry := range entries {
		require.False(t, strings.HasPrefix(entry.Name(), "."), entry.Name())
	}
}

// A site batch whose download fails has no folder, so the next restore downloads it again
func TestFailedBatchDownloadIsRetried(t *testing.T) {
	c, fake := newClient(t)
	key, folder := sitedb.GetS3BatchPath("a", c.ServerConfig), sitedb.GetBatchFolder("a", c.ServerConfig)
	archive, content := backup(t, c, "abcdef.sqlite")

	fake.set(key, archive[:len(archive)/2])
	require.Error(t, restore.TryRestoreAllFiles(c))
	require.NoDirExists(t, folder)

	fake.set(key, archive)
	require.NoError(t, restore.TryRestoreAllFiles(c))
	got, err := os.ReadFile(filepath.Join(folder, "abcdef.sqlite"))
	require.NoError(t, err)
	require.Equal(t, content, got)
	// Batches with no backup get an empty folder
	require.DirExists(t, sitedb.GetBatchFolder("z", c.ServerConfig))
}
