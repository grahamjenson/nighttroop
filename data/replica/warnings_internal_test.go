package replica

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Only Litestream's warnings and errors are logged, with the attributes and groups they are given
func TestWarningsOnly(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(warnings{slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})}).With("system", "store").WithGroup("g")
	logger.Debug("synced")
	logger.Info("compacted")
	logger.Warn("slow", "n", 1)
	logger.Error("failed")
	require.NotContains(t, logs.String(), "synced")
	require.NotContains(t, logs.String(), "compacted")
	require.Contains(t, logs.String(), `level=WARN msg=slow system=store g.n=1`)
	require.Contains(t, logs.String(), `level=ERROR msg=failed system=store`)
}

// A replica that S3 takes a request for and never answers fails a restore once restoreTimeout is up, so it does not
// hold the start for ever
func TestRestoreTimesOut(t *testing.T) {
	t.Setenv("AWS_CA_BUNDLE", "")
	defer func(old time.Duration) { restoreTimeout = old }(restoreTimeout)
	restoreTimeout = 300 * time.Millisecond
	stop := make(chan struct{})
	stalled := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-stop }))
	defer stalled.Close()
	defer close(stop)

	cfg := &config.ServerConfig{S3URL: stalled.URL, S3Bucket: "bucket", HostName: "nighttroop.test", S3AccessKeyID: "key", S3AccessSecretKey: "secret"}
	cfg.Status.S3Working = true
	r := New(cfg, s3utils.NewS3Client(cfg))
	started := time.Now()
	found, err := r.Restore(context.Background(), filepath.Join(t.TempDir(), "controldb.sqlite"))
	require.Error(t, err)
	require.False(t, found)
	require.Less(t, time.Since(started), 20*time.Second)
}
