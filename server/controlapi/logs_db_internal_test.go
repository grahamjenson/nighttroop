package controlapi

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestLogsDBDownload checks the control admin downloads a copy of the logs DB, with its captured requests, one at a
// time, and a read only control identity cannot
func TestLogsDBDownload(t *testing.T) {
	// A config from testutils cannot be made beside TestMain's, as each reads the flags
	controlDB, err := db.NewControlDB(t.Context(), testConfig)
	require.NoError(t, err)
	defer controlDB.Close()
	controlDB.Logs.Capture(t.Context(), nil, "Captured GET /api/v1/entries", map[string]any{"status": 200})
	downloadAs := func(role consts.ControlRole) *httptest.ResponseRecorder {
		ctx := httputils.CreateControlContext(context.Background(), controlDB, "", httputils.HTTPActionRead)
		ctx = httputils.SetAdminControlContext(ctx, &controldb.ControlIdentity{Name: "someone", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		LogsDBHandler(w, httptest.NewRequest(http.MethodGet, "/logs-db", nil).WithContext(ctx))
		return w
	}

	require.Equal(t, http.StatusForbidden, downloadAs(consts.ControlRoleRead).Code)

	// Only one copy is made at a time
	downloadingLogsDB.Lock()
	require.Equal(t, http.StatusTooManyRequests, downloadAs(consts.ControlRoleAdmin).Code)
	downloadingLogsDB.Unlock()

	w := downloadAs(consts.ControlRoleAdmin)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/vnd.sqlite3", w.Header().Get("Content-Type"))
	require.Regexp(t, `^attachment; filename="logs-\d{4}-\d\d-\d\dT\d\d-\d\d-\d\dZ\.sqlite"$`, w.Header().Get("Content-Disposition"))
	path := filepath.Join(t.TempDir(), "logs.sqlite")
	require.NoError(t, os.WriteFile(path, w.Body.Bytes(), 0o600))
	downloaded, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer downloaded.Close()
	search := "Captured GET"
	logs, err := logsdb.New(downloaded).SearchLogs(t.Context(), logsdb.SearchLogsParams{Search: &search, Limit: 10})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Contains(t, *logs[0].Attributes, `"status":200`)

	// The copy made for it is deleted once sent
	copies, err := filepath.Glob(filepath.Join(os.TempDir(), "logs-download-*.sqlite"))
	require.NoError(t, err)
	require.Empty(t, copies)
}
