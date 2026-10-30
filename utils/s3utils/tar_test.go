package s3utils

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pierrec/lz4/v4"
	"github.com/stretchr/testify/require"
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

// A backup is the site's files tarred, compressed and encrypted, and unpacks to the same files
func TestBackupRoundTrip(t *testing.T) {
	c, src, dst := testClient(t), t.TempDir(), t.TempDir()
	files := map[string][]byte{"site.sqlite": bytes.Repeat([]byte("glucose "), 10_000), "site.sqlite-wal": {}, "notes.txt": []byte("a note")}
	paths := []string{}
	for name, content := range files {
		path := filepath.Join(src, name)
		require.NoError(t, os.WriteFile(path, content, 0o640))
		paths = append(paths, path)
	}

	var backup bytes.Buffer
	written, err := c.createTarLZ4(paths, &backup)
	require.NoError(t, err)
	require.Equal(t, int64(backup.Len()), written)
	// Encrypted, so none of the files can be read from it
	require.NotContains(t, backup.String(), "a note")
	require.NotContains(t, backup.String(), "glucose glucose")

	require.NoError(t, c.ExtractTarLZ4(&backup, dst))
	for name, content := range files {
		got, err := os.ReadFile(filepath.Join(dst, name))
		require.NoError(t, err)
		require.Equal(t, content, got, name)
		info, err := os.Stat(filepath.Join(dst, name))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o640), info.Mode().Perm(), name)
	}
}

// A backup that has been tampered with, or is not a backup, is refused rather than unpacked
func TestBackupRefusesTampering(t *testing.T) {
	c := testClient(t)
	path := filepath.Join(t.TempDir(), "site.sqlite")
	require.NoError(t, os.WriteFile(path, []byte("readings"), 0o600))
	var backup bytes.Buffer
	_, err := c.createTarLZ4([]string{path}, &backup)
	require.NoError(t, err)

	tampered := bytes.Clone(backup.Bytes())
	tampered[len(tampered)/2] ^= 0xff
	require.Error(t, c.ExtractTarLZ4(bytes.NewReader(tampered), t.TempDir()))
	require.Error(t, c.ExtractTarLZ4(bytes.NewReader([]byte("not a backup")), t.TempDir()))
}

// An archive's entries cannot be written outside the folder it is unpacked into
func TestBackupCannotEscapeItsFolder(t *testing.T) {
	c := testClient(t)
	for _, name := range []string{"../escaped", "/tmp/escaped", "a/../../escaped"} {
		t.Run(name, func(t *testing.T) {
			var archive bytes.Buffer
			enc, err := c.ServerConfig.Security.EncryptingWriter(&archive)
			require.NoError(t, err)
			lz := lz4.NewWriter(enc)
			tw := tar.NewWriter(lz)
			require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: 1, Typeflag: tar.TypeReg}))
			_, err = tw.Write([]byte("x"))
			require.NoError(t, err)
			require.NoError(t, tw.Close())
			require.NoError(t, lz.Close())
			require.NoError(t, enc.Close())

			dir := t.TempDir()
			require.ErrorContains(t, c.ExtractTarLZ4(&archive, filepath.Join(dir, "out")), name)
			_, err = os.Stat(filepath.Join(dir, "escaped"))
			require.True(t, os.IsNotExist(err))
		})
	}
}
