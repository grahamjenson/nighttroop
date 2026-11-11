package sitecollections

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestDownloadData checks a site's admin downloads a zip of a copy of its DB, with its data, a few times a minute, and
// no one else can
func TestDownloadData(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	_, err := site.SiteDB.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	downloadAs := func(role consts.IdentityRole) *httptest.ResponseRecorder {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionRead, httputils.AcceptTypeHTML, httputils.ContentTypeForm)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		w := httptest.NewRecorder()
		DownloadDataHandler(w, httptest.NewRequest(http.MethodGet, "/usage/download", nil).WithContext(ctx))
		return w
	}

	require.Equal(t, http.StatusForbidden, downloadAs(consts.IdentityRoleUser).Code)

	w := downloadAs(consts.IdentityRoleAdmin)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	disposition := w.Header().Get("Content-Disposition")
	require.Contains(t, disposition, `attachment; filename="`+siteID+"-")
	require.True(t, strings.HasSuffix(disposition, `.zip"`), disposition)

	// The zip has the DB, named as the zip is, and nothing else yet
	archive, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	zipName := strings.TrimSuffix(strings.TrimPrefix(disposition, `attachment; filename="`), `.zip"`)
	require.Equal(t, zipName+".sqlite", archive.File[0].Name)
	entry, err := archive.File[0].Open()
	require.NoError(t, err)
	db, err := io.ReadAll(entry)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "download.sqlite")
	require.NoError(t, os.WriteFile(path, db, 0o600))
	downloaded, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer downloaded.Close()
	reading, err := sitedb.New(downloaded).GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, 120.0, reading.Mgdl)

	for range cfg.SiteDownloadPerMinuteBurst - 1 {
		require.Equal(t, http.StatusOK, downloadAs(consts.IdentityRoleAdmin).Code)
	}
	require.Equal(t, http.StatusTooManyRequests, downloadAs(consts.IdentityRoleAdmin).Code)
}
