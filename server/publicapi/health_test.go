package publicapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A live site whose DB file is missing at start and not restored makes the health check unavailable, so update.sh
// rolls back the release, without saying which site it is
func TestGETHealthWithASiteMissing(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	ctx := t.Context()
	health := func(controlDB *db.ControlDB) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		GETHealthHandler(controlDB)(w, httptest.NewRequest(http.MethodGet, "/health", nil))
		return w
	}
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(ctx, "kostsite")
	require.NoError(t, err)
	ok := health(controlDB)
	require.Equal(t, http.StatusOK, ok.Code)
	require.JSONEq(t, `{"status":"ok"}`, ok.Body.String())
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.Remove(sitedb.GetPath("kostsite", &cfg)))

	// S3 is not working, so the site cannot be restored
	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	require.NoError(t, controlDB.RestoreMissingSites(ctx, s3utils.NewS3Client(&cfg).DownloadFilesFromS3))
	missing := health(controlDB)
	require.Equal(t, http.StatusServiceUnavailable, missing.Code)
	require.JSONEq(t, `{"status":"unavailable"}`, missing.Body.String())
}
