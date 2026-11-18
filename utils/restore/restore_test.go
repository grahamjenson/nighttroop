package restore_test

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"io"
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

// backup is a DB file, too big for one encrypted segment, packed as a backup kept at key, and that file's content.
// It is a tar of the file compressed with LZ4 and named name.lz4, or, when legacy, as backups were packed before, a tar
// of the file as it is, compressed with LZ4 as a whole.
func backup(t *testing.T, c *s3utils.S3Client, key, name string, legacy bool) ([]byte, []byte) {
	content := make([]byte, 2<<20)
	_, err := rand.Read(content)
	require.NoError(t, err)

	// Small blocks, as a real DB is many blocks, so a cut off download has already written some of the file
	compress := func(w io.Writer) *lz4.Writer {
		lz := lz4.NewWriter(w)
		require.NoError(t, lz.Apply(lz4.BlockSizeOption(lz4.Block64Kb)))
		return lz
	}
	entry := content
	if !legacy {
		var compressed bytes.Buffer
		lz := compress(&compressed)
		_, err = lz.Write(content)
		require.NoError(t, err)
		require.NoError(t, lz.Close())
		entry, name = compressed.Bytes(), name+".lz4"
	}

	var archive bytes.Buffer
	enc, err := c.ServerConfig.Security.EncryptingWriterFor(&archive, key)
	require.NoError(t, err)
	var w io.WriteCloser = enc
	if legacy {
		w = compress(enc)
	}
	tw := tar.NewWriter(w)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(entry)), Typeflag: tar.TypeReg}))
	_, err = tw.Write(entry)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	if legacy {
		require.NoError(t, w.Close())
	}
	require.NoError(t, enc.Close())
	return archive.Bytes(), content
}

// legacyKey is where a backup kept at key was kept before backups were compressed file by file
func legacyKey(key string) string {
	return strings.TrimSuffix(key, ".tar.enc") + ".tar.lz4.enc"
}

// A site's backup is downloaded and decompressed
func TestDownloadDecompressesASiteBackup(t *testing.T) {
	c, fake := newClient(t)
	key, path := sitedb.GetS3SitePath("abcdefgh", c.ServerConfig), sitedb.GetPath("abcdefgh", c.ServerConfig)
	archive, content := backup(t, c, key, filepath.Base(path), false)
	fake.set(key, archive)
	out := filepath.Join(t.TempDir(), "site")

	_, found, err := c.DownloadFilesFromS3(t.Context(), key, out)
	require.NoError(t, err)
	require.True(t, found)
	got, err := os.ReadFile(filepath.Join(out, filepath.Base(path)))
	require.NoError(t, err)
	require.Equal(t, content, got)
	require.NoFileExists(t, filepath.Join(out, filepath.Base(path)+".lz4"))
}

// A download that is cut off leaves no DB, so the next try downloads it again, and nothing beside it
func TestCutOffDownloadLeavesNoDB(t *testing.T) {
	c, fake := newClient(t)
	key, path := sitedb.GetS3SitePath("abcdefgh", c.ServerConfig), sitedb.GetPath("abcdefgh", c.ServerConfig)
	archive, content := backup(t, c, key, filepath.Base(path), false)
	dir := t.TempDir()
	out, file := filepath.Join(dir, "site"), filepath.Join(dir, "site", filepath.Base(path))

	fake.set(key, archive[:len(archive)/2])
	_, _, err := c.DownloadFilesFromS3(t.Context(), key, out)
	require.Error(t, err)
	require.NoFileExists(t, file)

	fake.set(key, archive)
	_, _, err = c.DownloadFilesFromS3(t.Context(), key, out)
	require.NoError(t, err)
	got, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, content, got)

	// Nothing is left beside it from either download
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, entry := range entries {
		require.False(t, strings.HasPrefix(entry.Name(), "."), entry.Name())
	}
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
	batch := sitedb.GetSitesFolder("a", cfg)
	backupBatch := filepath.Dir(sitedb.GetBackupPath("abcdef", cfg))

	// Made as the download and the archive make them, so the names are theirs
	staging := []string{}
	for dir, pattern := range map[string]string{
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
