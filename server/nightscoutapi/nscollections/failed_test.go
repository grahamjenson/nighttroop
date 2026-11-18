package nscollections

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A reading sent by a client that stopped waiting, as an uploader that timed out, is Client Closed Request, logged as
// info, not the server's error; a database that fails for any other reason is an internal error, logged with why
func TestCreateCancelledIsClientClosedRequest(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	siteDB, model, err := controlDB.GetSite(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	siteContext := &httputils.SiteContext{SiteID: model.ID, ControlDB: controlDB, SiteDB: siteDB, SiteModel: model,
		IdentityModel: &controldb.SiteIdentity{ID: 1, Name: "pwd_phone"}}
	reading := map[string]any{"type": "sgv", "sgv": 120, "date": float64(1762000000123)}
	var logs bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(before) })

	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, httpErr := NewEntriesCollection().Create(cancelled, siteContext, reading)
	require.Equal(t, nsutils.ErrClientClosedRequest, httpErr)
	require.Contains(t, logs.String(), "level=INFO msg=\"Request cancelled, as its client stopped waiting, so it did not begin in database\"")
	require.NotContains(t, logs.String(), "level=ERROR")

	logs.Reset()
	require.NoError(t, siteDB.SQLiteDB.Close())
	_, _, httpErr = NewEntriesCollection().Create(t.Context(), siteContext, reading)
	require.Equal(t, nsutils.ErrInternal, httpErr)
	require.Contains(t, logs.String(), "level=ERROR msg=\"failed to begin in database\" error=\"sql: database is closed\"")
}
