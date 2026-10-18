package s3utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

func TestS3IsSkippedWhenNotWorking(t *testing.T) {
	// No S3 connection at all: any real call would panic
	c := &s3utils.S3Client{ServerConfig: &config.ServerConfig{}}

	require.NoError(t, c.UploadFilesToS3("key", []string{"file"}))

	size, found, err := c.DownloadFilesFromS3("key", t.TempDir())
	require.NoError(t, err)
	require.False(t, found)
	require.Zero(t, size)
}
