package s3utils_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

func TestS3IsSkippedWhenNotWorking(t *testing.T) {
	// No S3 connection at all: any real call would panic
	c := &s3utils.S3Client{ServerConfig: &config.ServerConfig{}}

	require.NoError(t, c.UploadFilesToS3("key", s3utils.SiteRetention, []string{"file"}))

	size, found, err := c.DownloadFilesFromS3("key", t.TempDir())
	require.NoError(t, err)
	require.False(t, found)
	require.Zero(t, size)
}

func TestRetentionRules(t *testing.T) {
	c := &s3utils.S3Client{ServerConfig: &config.ServerConfig{SiteBackupRetentionDays: 30, ControlBackupRetentionDays: 90}}

	// Each rule expires the replaced versions of the uploads tagged with its retention
	days := map[string]int32{}
	for _, rule := range c.RetentionRules() {
		require.Equal(t, "nighttroop-retention", aws.ToString(rule.Filter.Tag.Key))
		require.Equal(t, "nighttroop-retention-"+aws.ToString(rule.Filter.Tag.Value), aws.ToString(rule.ID))
		require.Nil(t, rule.Expiration, "the current version must never expire")
		days[aws.ToString(rule.Filter.Tag.Value)] = aws.ToInt32(rule.NoncurrentVersionExpiration.NoncurrentDays)
	}
	require.Equal(t, map[string]int32{string(s3utils.SiteRetention): 30, string(s3utils.ControlRetention): 90}, days)
}
