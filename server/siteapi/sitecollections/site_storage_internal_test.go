package sitecollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
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

// A site's admin can ask for its summaries to be built again only a few times a minute, as each build reads a year of
// its data, sharing the site's downloads; a user cannot ask at all, and spends none of them
func TestRegenerateSummariesIsRateLimited(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	regenerateAs := func(role consts.IdentityRole) int {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		w := httptest.NewRecorder()
		RegenerateSummariesHandler(w, httptest.NewRequest(http.MethodPost, "/reports/summaries/regenerate", nil).WithContext(ctx))
		return w.Code
	}

	require.Equal(t, http.StatusForbidden, regenerateAs(consts.IdentityRoleUser))
	for range cfg.SiteDownloadPerMinuteBurst {
		require.Equal(t, http.StatusAccepted, regenerateAs(consts.IdentityRoleAdmin))
	}
	require.Equal(t, http.StatusTooManyRequests, regenerateAs(consts.IdentityRoleAdmin))
	logged, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	requested := 0
	for _, event := range logged {
		if event.Name == string(events.SummariesRegenerationRequested) {
			requested++
		}
	}
	require.Equal(t, cfg.SiteDownloadPerMinuteBurst, requested)
}
