package testutils

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/benbjohnson/litestream"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/superfly/ltx"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
)

// FakeS3 is the client of an S3 in memory, which cfg is changed to use, and working. It takes no SSE-C, as it is
// HTTP, so the control DB's replica is not encrypted.
func FakeS3(t testing.TB, cfg *config.ServerConfig) *s3utils.S3Client {
	t.Helper()
	// Litestream's S3 client cannot start with a CA bundle, as a sandbox's proxy may set
	t.Setenv("AWS_CA_BUNDLE", "")
	// The bucket is made now, not by gofakes3's auto bucket: that makes it in each request that finds it missing, so
	// of two first requests at once, as a test's and Litestream's, one fails with NoSuchBucket
	backend := s3mem.New()
	if err := backend.CreateBucket(cfg.S3Bucket); err != nil {
		t.Fatalf("cannot make the fake S3's bucket: %v", err)
	}
	server := httptest.NewServer(gofakes3.New(backend).Server())
	t.Cleanup(server.Close)
	cfg.S3URL, cfg.Status.S3Working = server.URL, true
	return s3utils.NewS3Client(cfg)
}

// ReplicatedConfig is the test config with its DBs in files in a folder of the test's, and S3 in memory, for tests that
// back up and replicate DBs, as a server with S3 does
func ReplicatedConfig(t testing.TB) (*config.ServerConfig, *s3utils.S3Client) {
	t.Helper()
	cfg := *GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	return &cfg, FakeS3(t, &cfg)
}

// ControlReplicaFiles are the files of the control DB's replica, by level and TXIDs, read as the server reads them
func ControlReplicaFiles(t testing.TB, cfg *config.ServerConfig, s3 *s3utils.S3Client) []string {
	t.Helper()
	client := replica.Client(cfg, s3, false)
	files := []string{}
	for level := range litestream.SnapshotLevel + 1 {
		found, err := litestream.FindLTXFiles(t.Context(), client, level, false, func(*ltx.FileInfo) (bool, error) { return true, nil })
		if err != nil {
			t.Fatalf("cannot list the replica's files: %v", err)
		}
		for _, info := range found {
			files = append(files, fmt.Sprintf("%d/%s", level, ltx.FormatFilename(info.MinTXID, info.MaxTXID)))
		}
	}
	return files
}
