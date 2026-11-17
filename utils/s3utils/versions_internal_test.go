package s3utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// versionsS3 is a stand in for S3 that lists two versions of sites/aaaaaaaa.tar.enc, and one of a longer key with the
// same prefix, and records the version a download asks for, which it does not have
func versionsS3(t *testing.T, asked *string) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["versions"]; ok {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListVersionsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>bucket</Name><IsTruncated>false</IsTruncated>
<Version><Key>sites/aaaaaaaa.tar.enc</Key><VersionId>older</VersionId><IsLatest>false</IsLatest><LastModified>2026-11-17T09:00:00.000Z</LastModified><Size>100</Size></Version>
<Version><Key>sites/aaaaaaaab.tar.enc</Key><VersionId>other</VersionId><IsLatest>true</IsLatest><LastModified>2026-11-17T11:00:00.000Z</LastModified><Size>300</Size></Version>
<Version><Key>sites/aaaaaaaa.tar.enc</Key><VersionId>newest</VersionId><IsLatest>true</IsLatest><LastModified>2026-11-17T10:00:00.000Z</LastModified><Size>200</Size></Version>
</ListVersionsResult>`))
			return
		}
		*asked = r.URL.Query().Get("versionId")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchKey</Code><Message>no</Message></Error>`))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

// The versions S3 keeps of a backup are listed newest first, leaving out another key with the same prefix, and a
// version is downloaded by its ID
func TestBackupVersions(t *testing.T) {
	shared, err := testConfig()
	require.NoError(t, err)
	serverConfig := *shared
	asked := ""
	serverConfig.S3URL, serverConfig.Status.S3Working = versionsS3(t, &asked), true
	c := NewS3Client(&serverConfig)

	versions, err := c.ListVersions(context.Background(), "sites/aaaaaaaa.tar.enc")
	require.NoError(t, err)
	at := func(s string) int64 { parsed, _ := time.Parse(time.RFC3339, s); return parsed.UnixMilli() }
	require.Equal(t, []BackupVersion{
		{ID: "newest", At: at("2026-11-17T10:00:00Z"), SizeBytes: 200, Latest: true},
		{ID: "older", At: at("2026-11-17T09:00:00Z"), SizeBytes: 100},
	}, versions)

	_, found, err := c.VersionDownload("older")(context.Background(), "sites/aaaaaaaa.tar.enc", filepath.Join(t.TempDir(), "out"))
	require.NoError(t, err)
	require.False(t, found, "the stand in does not have it")
	require.Equal(t, "older", asked, "the download asks for the version")

	serverConfig.Status.S3Working = false
	versions, err = c.ListVersions(context.Background(), "sites/aaaaaaaa.tar.enc")
	require.NoError(t, err)
	require.Empty(t, versions, "without S3 there are none")
}
