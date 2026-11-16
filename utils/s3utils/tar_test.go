package s3utils

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pierrec/lz4/v4"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"nighttroop.t1d.nz/utils/config"
)

// testConfig is the test server config, with its key sets, made once as it parses the command line's flags
var testConfig = sync.OnceValues(config.NewServerConfig)

// testClient is an S3 client with the test key sets, for packing and unpacking backups without S3
func testClient(t *testing.T) *S3Client {
	cfg, err := testConfig()
	require.NoError(t, err)
	return &S3Client{ServerConfig: cfg}
}

// compressLZ4 is data compressed with LZ4, as a DB's backup file is
func compressLZ4(t *testing.T, data []byte) []byte {
	var compressed bytes.Buffer
	lz := lz4.NewWriter(&compressed)
	_, err := lz.Write(data)
	require.NoError(t, err)
	require.NoError(t, lz.Close())
	return compressed.Bytes()
}

// A backup is the site's files, each DB already compressed, tarred and encrypted, and unpacks to the same files, each
// DB decompressed, without logging each, as a restore unpacks every site's
func TestBackupRoundTrip(t *testing.T) {
	c, src, dst := testClient(t), t.TempDir(), t.TempDir()
	db := bytes.Repeat([]byte("glucose "), 10_000)
	files := map[string][]byte{"site.sqlite.lz4": compressLZ4(t, db), "site.sqlite-wal": {}, "notes.txt": []byte("a note")}
	paths := []string{}
	for name, content := range files {
		path := filepath.Join(src, name)
		require.NoError(t, os.WriteFile(path, content, 0o640))
		paths = append(paths, path)
	}
	delete(files, "site.sqlite.lz4")
	files["site.sqlite"] = db

	var backup bytes.Buffer
	written, err := c.createTar(paths, &backup, testKey)
	require.NoError(t, err)
	require.Equal(t, int64(backup.Len()), written)
	// Encrypted, so none of the files can be read from it
	require.NotContains(t, backup.String(), "a note")
	require.NotContains(t, backup.String(), "glucose glucose")

	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logged, nil)))
	err = c.ExtractTar(&backup, dst, testKey)
	slog.SetDefault(previous)
	require.NoError(t, err)
	require.Empty(t, logged.String())
	require.NoFileExists(t, filepath.Join(dst, "site.sqlite.lz4"))
	for name, content := range files {
		got, err := os.ReadFile(filepath.Join(dst, name))
		require.NoError(t, err)
		require.Equal(t, content, got, name)
		info, err := os.Stat(filepath.Join(dst, name))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o640), info.Mode().Perm(), name)
	}
}

// A backup of no files, a batch whose last site went, unpacks to nothing
func TestEmptyBackupRoundTrip(t *testing.T) {
	c, dst := testClient(t), t.TempDir()
	var backup bytes.Buffer
	_, err := c.createTar(nil, &backup, testKey)
	require.NoError(t, err)
	require.NoError(t, c.ExtractTar(&backup, dst, testKey))
	entries, err := os.ReadDir(dst)
	require.NoError(t, err)
	require.Empty(t, entries)
}

// A backup that has been tampered with, or is not a backup, is refused rather than unpacked
func TestBackupRefusesTampering(t *testing.T) {
	c := testClient(t)
	path := filepath.Join(t.TempDir(), "site.sqlite")
	require.NoError(t, os.WriteFile(path, []byte("readings"), 0o600))
	var backup bytes.Buffer
	_, err := c.createTar([]string{path}, &backup, testKey)
	require.NoError(t, err)

	tampered := bytes.Clone(backup.Bytes())
	tampered[len(tampered)/2] ^= 0xff
	require.Error(t, c.ExtractTar(bytes.NewReader(tampered), t.TempDir(), testKey))
	require.Error(t, c.ExtractTar(bytes.NewReader([]byte("not a backup")), t.TempDir(), testKey))
}

// An archive's entries cannot be written outside the folder it is unpacked into
func TestBackupCannotEscapeItsFolder(t *testing.T) {
	c := testClient(t)
	for _, name := range []string{"../escaped", "/tmp/escaped", "a/../../escaped", "../escaped.lz4"} {
		t.Run(name, func(t *testing.T) {
			var archive bytes.Buffer
			enc, err := c.ServerConfig.Security.EncryptingWriterFor(&archive, testKey)
			require.NoError(t, err)
			tw := tar.NewWriter(enc)
			require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: 1, Typeflag: tar.TypeReg}))
			_, err = tw.Write([]byte("x"))
			require.NoError(t, err)
			require.NoError(t, tw.Close())
			require.NoError(t, enc.Close())

			dir := t.TempDir()
			require.ErrorContains(t, c.ExtractTar(&archive, filepath.Join(dir, "out"), testKey), name)
			_, err = os.Stat(filepath.Join(dir, "escaped"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

// failingClose is a file whose Close fails, as a full disk can report only when the file is closed
type failingClose struct{ *os.File }

func (f failingClose) Close() error {
	f.File.Close()
	return errors.New("no space left on device")
}

// A restored file that cannot be closed fails the restore, as it may have been cut short
func TestBackupRestoreFailsWhenAFileCannotBeClosed(t *testing.T) {
	c := testClient(t)
	path := filepath.Join(t.TempDir(), "site.sqlite")
	require.NoError(t, os.WriteFile(path, []byte("readings"), 0o600))
	var backup bytes.Buffer
	_, err := c.createTar([]string{path}, &backup, testKey)
	require.NoError(t, err)

	previous := createFile
	t.Cleanup(func() { createFile = previous })
	createFile = func(name string) (io.WriteCloser, error) {
		f, err := os.Create(name)
		return failingClose{f}, err
	}
	require.ErrorContains(t, c.ExtractTar(&backup, t.TempDir(), testKey), "no space left on device")
}

// testKey is where the tests' backups are kept
const testKey = "example.org/backup/a.tar.enc"

// A backup only unpacks for the object key it was made for, whatever the host, so one batch's archive cannot be put at
// another's key, nor one encrypted with no key
func TestBackupBoundToItsKey(t *testing.T) {
	c := testClient(t)
	path := filepath.Join(t.TempDir(), "site.sqlite")
	content := make([]byte, 3<<20)
	_, err := rand.Read(content)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, content, 0o600))
	var backup bytes.Buffer
	_, err = c.createTar([]string{path}, &backup, testKey)
	require.NoError(t, err)

	require.NoError(t, c.ExtractTar(bytes.NewReader(backup.Bytes()), t.TempDir(), "another.host/backup/a.tar.enc"))
	require.Error(t, c.ExtractTar(bytes.NewReader(backup.Bytes()), t.TempDir(), "example.org/backup/b.tar.enc"))

	unbound, err := streamingaead.New(c.ServerConfig.Security.StreamEncryptionKeySet)
	require.NoError(t, err)
	var archive bytes.Buffer
	enc, err := unbound.NewEncryptingWriter(&archive, nil)
	require.NoError(t, err)
	tw := tar.NewWriter(enc)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "site.sqlite", Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, enc.Close())
	require.Error(t, c.ExtractTar(bytes.NewReader(archive.Bytes()), t.TempDir(), testKey))
}

// A backup kept as backups were before, a tar of the DBs compressed as a whole, unpacks at the legacy key it was kept at
func TestLegacyBackupUnpacks(t *testing.T) {
	c, dst := testClient(t), t.TempDir()
	key := LegacyKey(testKey)
	require.Equal(t, "example.org/backup/a.tar.lz4.enc", key)
	require.Equal(t, key, LegacyKey(key))
	content := bytes.Repeat([]byte("glucose "), 10_000)

	var archive bytes.Buffer
	enc, err := c.ServerConfig.Security.EncryptingWriterFor(&archive, key)
	require.NoError(t, err)
	lz := lz4.NewWriter(enc)
	tw := tar.NewWriter(lz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "site.sqlite", Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, lz.Close())
	require.NoError(t, enc.Close())

	require.Error(t, c.ExtractTar(bytes.NewReader(archive.Bytes()), t.TempDir(), testKey))
	require.NoError(t, c.ExtractTar(&archive, dst, key))
	got, err := os.ReadFile(filepath.Join(dst, "site.sqlite"))
	require.NoError(t, err)
	require.Equal(t, content, got)
}
