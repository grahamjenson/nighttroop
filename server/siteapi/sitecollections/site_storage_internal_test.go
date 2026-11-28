package sitecollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's admin sees its DB's size and tables on its Usage page and truncates it there, recorded as the control
// site's Truncate is; a user sees neither and cannot
func TestUsageAndTruncate(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	_, err := site.SiteDB.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	request := func(role consts.IdentityRole, method, path string) *http.Request {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeHTMX, httputils.ContentTypeForm)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		return httptest.NewRequest(method, path, nil).WithContext(ctx)
	}

	w := httptest.NewRecorder()
	UsageHandler(w, request(consts.IdentityRoleUser, http.MethodGet, "/usage"))
	require.Equal(t, http.StatusForbidden, w.Code)
	w = httptest.NewRecorder()
	UsageHandler(w, request(consts.IdentityRoleAdmin, http.MethodGet, "/usage"))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `id="site--usage--db-size"`)
	require.Contains(t, w.Body.String(), `id="site--usage--db-table--glucose_readings"`)
	require.Contains(t, w.Body.String(), `id="site--usage--truncate"`)
	require.Contains(t, w.Body.String(), `href="/usage/download"`)

	w = httptest.NewRecorder()
	TruncateHandler(w, request(consts.IdentityRoleUser, http.MethodPost, "/usage/truncate"))
	require.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	TruncateHandler(w, request(consts.IdentityRoleAdmin, http.MethodPost, "/usage/truncate"))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `id="site--usage--storage"`)
	require.Contains(t, w.Body.String(), "Site Truncated")
	events, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_truncated", events[0].Name)
	require.Contains(t, *events[0].Attributes, `"by":"someone"`)
	// Its data is kept: truncating only clears and thins device reports
	_, err = site.SiteDB.Queries().GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
}
