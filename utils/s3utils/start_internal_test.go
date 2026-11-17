package s3utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// Without S3_URL S3 is off and never tested; with it, the server only starts once a test works, tried again after each
// wait, and is told why it never did
func TestStart(t *testing.T) {
	waits := []time.Duration{0, 0}
	tests := []struct {
		name     string
		endpoint string
		failures int
		wantErr  string
		wantRuns int
	}{
		{name: "without S3_URL S3 is off and never tested", endpoint: "", failures: 99, wantRuns: 0},
		{name: "a working S3 is on", endpoint: "http://s3.test", wantRuns: 1},
		{name: "an S3 that works on a later try is on", endpoint: "http://s3.test", failures: 2, wantRuns: 3},
		{name: "an S3 that never works stops the server", endpoint: "http://s3.test", failures: 99, wantRuns: 3, wantErr: "S3_URL is set but S3 is not working: refused"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewS3Client(&config.ServerConfig{S3URL: test.endpoint})
			runs := 0
			err := c.start(context.Background(), func(context.Context) error {
				runs++
				if runs <= test.failures {
					return errors.New("refused")
				}
				// Working S3 has versioning set up next, which fails here against no server, and is only logged
				return nil
			}, waits)
			require.Equal(t, test.wantRuns, runs)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				require.False(t, c.ServerConfig.Status.S3Working)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.endpoint != "", c.ServerConfig.Status.S3Working)
		})
	}
}

// An S3 that takes a request and never answers fails an upload and a download once transferTimeout is up, instead of
// holding the backup job until a restart
func TestTransfersTimeOut(t *testing.T) {
	defer func(old time.Duration) { transferTimeout = old }(transferTimeout)
	transferTimeout = 200 * time.Millisecond
	stop := make(chan struct{})
	stalled := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-stop }))
	defer stalled.Close()
	defer close(stop)

	shared, err := testConfig()
	require.NoError(t, err)
	serverConfig := *shared
	serverConfig.S3URL = stalled.URL
	serverConfig.Status.S3Working = true
	c := NewS3Client(&serverConfig)
	file := filepath.Join(t.TempDir(), "site.db")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0o600))

	start := time.Now()
	require.Error(t, c.UploadFilesToS3(context.Background(), "site.tar.lz4", SiteRetention, []string{file}))
	_, _, err = c.DownloadFilesFromS3(context.Background(), "site.tar.lz4", filepath.Join(t.TempDir(), "out"))
	require.Error(t, err)
	require.Less(t, time.Since(start), 20*time.Second)
}
