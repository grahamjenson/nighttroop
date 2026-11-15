package nighttroop_test

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// The generated client reads the site's usage, truncates it with its size before and after, and downloads its data
func TestUsageClient(t *testing.T) {
	site, auth := signedInSite(t)
	ctx := context.Background()

	usage, err := site.GetUsageWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, usage.StatusCode(), string(usage.Body))
	// A test's site DB is in memory, so its size is not measured, but its tables are, when it is created
	require.Equal(t, int64(20480), *usage.JSON200.Result.LimitKb)
	require.NotEmpty(t, usage.JSON200.Result.Tables)
	tables := usage.JSON200.Result.Tables
	require.GreaterOrEqual(t, tables[0].Bytes, tables[len(tables)-1].Bytes, "largest first")
	require.Nil(t, usage.JSON200.Result.BeforeSizeKb, "only a truncate has the size before it")

	truncated, err := site.TruncateSiteWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, truncated.StatusCode(), string(truncated.Body))
	require.Equal(t, usage.JSON200.Result.SizeKb, truncated.JSON200.Result.BeforeSizeKb)
	require.Equal(t, int64(20480), *truncated.JSON200.Result.LimitKb)
	require.NotEmpty(t, truncated.JSON200.Result.Tables)

	download, err := site.DownloadSiteDataWithResponse(ctx, auth)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, download.StatusCode())
	require.Equal(t, "application/zip", download.HTTPResponse.Header.Get("Content-Type"))
	archive, err := zip.NewReader(bytes.NewReader(download.Body), int64(len(download.Body)))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
}
