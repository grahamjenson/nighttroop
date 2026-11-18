package s3utils_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

func TestS3IsSkippedWhenNotWorking(t *testing.T) {
	// No S3 connection at all: any real call would panic
	c := &s3utils.S3Client{ServerConfig: &config.ServerConfig{}}

	require.NoError(t, c.UploadFilesToS3(context.Background(), "key", s3utils.SiteRetention, []string{"file"}))

	size, found, err := c.DownloadFilesFromS3(context.Background(), "key", t.TempDir())
	require.NoError(t, err)
	require.False(t, found)
	require.Zero(t, size)
}

func TestRetentionRules(t *testing.T) {
	c := &s3utils.S3Client{HostName: "host", ServerConfig: &config.ServerConfig{SiteBackupRetentionDays: 30, ControlBackupRetentionDays: 90}}

	// The site backups, tagged, and the files of the control DB's replica, by prefix, keep replaced versions their days
	rules := c.RetentionRules()
	require.Len(t, rules, 2)
	for _, rule := range rules {
		require.Nil(t, rule.Expiration, "the current version must never expire")
	}
	require.Equal(t, "nighttroop-retention-site", aws.ToString(rules[0].ID))
	require.Equal(t, "nighttroop-retention", aws.ToString(rules[0].Filter.Tag.Key))
	require.Equal(t, string(s3utils.SiteRetention), aws.ToString(rules[0].Filter.Tag.Value))
	require.Equal(t, int32(30), aws.ToInt32(rules[0].NoncurrentVersionExpiration.NoncurrentDays))
	require.Equal(t, "nighttroop-retention-control", aws.ToString(rules[1].ID))
	require.Equal(t, "host/litestream/control", aws.ToString(rules[1].Filter.Prefix))
	require.Equal(t, int32(90), aws.ToInt32(rules[1].NoncurrentVersionExpiration.NoncurrentDays))
}
