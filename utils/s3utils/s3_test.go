package s3utils_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestRetentionRules(t *testing.T) {
	c := &s3utils.S3Client{HostName: "host", ServerConfig: &config.ServerConfig{SiteBackupRetentionDays: 30, ControlBackupRetentionDays: 90}}

	// Each rule expires the replaced versions of the files of the sites' replicas, or the control DB's
	days := map[string]int32{}
	for _, rule := range c.RetentionRules() {
		require.Nil(t, rule.Expiration, "the current version must never expire")
		days[aws.ToString(rule.ID)+" "+aws.ToString(rule.Filter.Prefix)] = aws.ToInt32(rule.NoncurrentVersionExpiration.NoncurrentDays)
	}
	require.Equal(t, map[string]int32{"nighttroop-retention-site host/litestream/site": 30, "nighttroop-retention-control host/litestream/control": 90}, days)
}

// An object is put, in one request or in parts when large, got, listed and deleted
func TestObjects(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	c := testutils.FakeS3(t, &cfg)
	small, large := []byte("small"), make([]byte, 6<<20)
	_, err := rand.Read(large)
	require.NoError(t, err)
	for key, content := range map[string][]byte{"objects/small": small, "objects/large": large} {
		sent, err := c.Put(t.Context(), key, bytes.NewReader(content))
		require.NoError(t, err)
		require.Equal(t, int64(len(content)), sent)
		body, err := c.Get(t.Context(), key)
		require.NoError(t, err)
		got, err := io.ReadAll(body)
		require.NoError(t, err)
		require.NoError(t, body.Close())
		require.Equal(t, content, got)
	}

	listed, err := c.List(t.Context(), "objects/")
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.NoError(t, c.Delete(t.Context(), []string{"objects/small", "objects/large"}))
	listed, err = c.List(t.Context(), "objects/")
	require.NoError(t, err)
	require.Empty(t, listed)
	_, err = c.Get(t.Context(), "objects/small")
	require.ErrorIs(t, err, os.ErrNotExist)
}
