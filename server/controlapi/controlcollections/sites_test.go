package controlcollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestDisablingSiteRecordsEvent(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := context.TODO(), site.SiteDB.Subdomain
	controlContext := &httputils.ControlContext{ControlDB: controlDB, IdentityModel: &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}}

	_, err := (&SitesCollection{}).UpdateRow(ctx, controlContext, &controlmodels.Site{ID: siteID}, &controlmodels.Site{State: "disabled"})
	require.NoError(t, err)

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_disabled", events[0].Name)
	require.Equal(t, `{"by":"admin"}`, *events[0].Attributes)
}

// TestDisablingActiveSiteRefused checks only a site not written to or seen in DisableAfterInactivityDays is disabled
func TestDisablingActiveSiteRefused(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	cfg.DisableAfterInactivityDays = 7
	t.Cleanup(func() { cfg.DisableAfterInactivityDays = 0 })
	ctx, siteID, now := context.TODO(), site.SiteDB.Subdomain, time.Now().UnixMilli()
	controlContext := &httputils.ControlContext{ControlDB: controlDB, IdentityModel: &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}}

	active := &controlmodels.Site{ID: siteID, State: "live", CreatedAt: now, LastSeenAt: &now, LastWriteAt: &now}
	_, err := (&SitesCollection{}).UpdateRow(ctx, controlContext, active, &controlmodels.Site{State: "disabled"})
	require.ErrorContains(t, err, "only a site not written to or seen in 7 days can be disabled")

	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "live", row.State)
}

func TestTruncateHandler(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := context.TODO(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.Post("/sites/{id}/truncate", (&SitesCollection{}).TruncateHandler)
	truncate := func(role consts.ControlRole) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/sites/"+siteID+"/truncate", nil)
		r.Header.Set("Accept", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	require.Equal(t, http.StatusBadRequest, truncate(consts.ControlRoleAdmin).Code, "a site within its limit is not truncated")

	oversize := int64(5000)
	require.NoError(t, controlDB.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: &oversize, ID: siteID}))
	require.Equal(t, http.StatusForbidden, truncate(consts.ControlRoleRead).Code)
	require.Equal(t, http.StatusOK, truncate(consts.ControlRoleAdmin).Code)

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_truncated", events[0].Name)
	require.Equal(t, `{"by":"admin","before_size_kb":5000,"after_size_kb":null}`, *events[0].Attributes)
}
