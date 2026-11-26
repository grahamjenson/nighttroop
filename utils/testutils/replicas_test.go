package testutils_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// The fake S3 has its bucket before any request, so two first requests at once, as a test's and Litestream's, do not
// both make it, and one fail with NoSuchBucket
func TestFakeS3HasItsBucketBeforeAnyRequest(t *testing.T) {
	cfg, _ := testutils.ReplicatedConfig(t)
	// Listing the buckets makes none
	res, err := http.Get(cfg.S3URL + "/")
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "<Name>"+cfg.S3Bucket+"</Name>")
}
