package s3utils_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
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
