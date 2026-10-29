package sitecollections

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestDownloadData checks a site's admin downloads a copy of its DB, with its data, a few times a minute, and no one
// else can
func TestDownloadData(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	_, err := site.SiteDB.Queries.CreateGlucoseReading(context.TODO(), sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	downloadAs := func(role consts.IdentityRole) *httptest.ResponseRecorder {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionRead, httputils.AcceptTypeHTML, httputils.ContentTypeForm)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		w := httptest.NewRecorder()
		DownloadDataHandler(w, httptest.NewRequest(http.MethodGet, "/settings/site/download", nil).WithContext(ctx))
		return w
	}

	require.Equal(t, http.StatusForbidden, downloadAs(consts.IdentityRoleUser).Code)

	w := downloadAs(consts.IdentityRoleAdmin)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Disposition"), `attachment; filename="`+siteID+"-")
	path := filepath.Join(t.TempDir(), "download.sqlite")
	require.NoError(t, os.WriteFile(path, w.Body.Bytes(), 0o600))
	downloaded, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer downloaded.Close()
	reading, err := sitedb.New(downloaded).GetGlucoseReading(context.TODO(), sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, 120.0, reading.Mgdl)

	for range cfg.SiteDownloadPerMinuteBurst - 1 {
		require.Equal(t, http.StatusOK, downloadAs(consts.IdentityRoleAdmin).Code)
	}
	require.Equal(t, http.StatusTooManyRequests, downloadAs(consts.IdentityRoleAdmin).Code)
}
