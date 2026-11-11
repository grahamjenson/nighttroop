package controlcollections

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
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
	ctx, siteID, now := context.Background(), site.SiteDB.Subdomain, time.Now().UnixMilli()
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
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.With(AdminOnly).Post("/sites/{id}/truncate", (&SitesCollection{}).TruncateHandler)
	truncate := func(role consts.ControlRole) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/sites/"+siteID+"/truncate", nil)
		r.Header.Set("Accept", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	// Any site whose DB is here can be truncated by an admin, within its DB size limit or over it
	size := int64(1000)
	require.NoError(t, controlDB.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: &size, ID: siteID}))
	require.Equal(t, http.StatusForbidden, truncate(consts.ControlRoleRead).Code)
	require.Equal(t, http.StatusOK, truncate(consts.ControlRoleAdmin).Code, "a site within its limit is truncated")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_truncated", events[0].Name)
	require.Equal(t, `{"by":"admin","before_size_kb":1000,"after_size_kb":null}`, *events[0].Attributes)

	// An archived site has no DB here to truncate
	_, err = controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)
	_, err = controlDB.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: `["` + siteID + `"]`})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, truncate(consts.ControlRoleAdmin).Code, "an archived site is not truncated")
}

// TestSettingsHandler checks only an admin can turn public site creation off and on
func TestSettingsHandler(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	settings := func(role consts.ControlRole, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(body))
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		AdminOnly(http.HandlerFunc((&SitesCollection{}).SettingsHandler)).ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	require.True(t, controlDB.PublicSiteCreation(t.Context()), "the public can create sites on a new server")
	require.Equal(t, http.StatusForbidden, settings(consts.ControlRoleRead, `{"public_site_creation": false}`).Code)
	require.True(t, controlDB.PublicSiteCreation(t.Context()))

	off := settings(consts.ControlRoleAdmin, `{"public_site_creation": false}`)
	require.Equal(t, http.StatusOK, off.Code)
	require.JSONEq(t, `{"status": 200, "result": {"public_site_creation": false}}`, off.Body.String())
	require.False(t, controlDB.PublicSiteCreation(t.Context()))

	require.Equal(t, http.StatusOK, settings(consts.ControlRoleAdmin, `{"public_site_creation": true}`).Code)
	require.True(t, controlDB.PublicSiteCreation(t.Context()))
}

// TestCreateHandlerOverMaxSites checks MAX_SITES and turning off public site creation do not stop an admin
func TestCreateHandlerOverMaxSites(t *testing.T) {
	cfg, controlDB, _ := testutils.NewControlDBWithSite(t)
	maxSites := cfg.MaxSites
	cfg.MaxSites = controlDB.GetNSitesDBs()
	t.Cleanup(func() { cfg.MaxSites = maxSites })
	require.NoError(t, controlDB.Queries.UpdatePublicSiteCreation(t.Context(), controldb.UpdatePublicSiteCreationParams{PublicSiteCreation: false}))

	r := httptest.NewRequest(http.MethodPost, "/sites", strings.NewReader(`{"id": "overmaxs"}`))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
	rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
	w := httptest.NewRecorder()
	(&SitesCollection{}).CreateHandler(w, r.WithContext(rctx))
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.Equal(t, cfg.MaxSites+1, controlDB.GetNSitesDBs())
}

// An admin downloads the zip of a site's DB its admins get from Download Data, which its admins see in its events; a
// control identity who is not an admin cannot, nor can a site with no DB here
func TestDownloadDBHandler(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.With(AdminOnly).Get("/sites/{id}/db", (&SitesCollection{}).DownloadDBHandler)
	download := func(role consts.ControlRole, id string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/sites/"+id+"/db", nil)
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionRead)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: role}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	require.Equal(t, http.StatusForbidden, download(consts.ControlRoleRead, siteID).Code)
	require.Equal(t, http.StatusNotFound, download(consts.ControlRoleAdmin, "nosuchsite").Code)

	w := download(consts.ControlRoleAdmin, siteID)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	name := siteID + "-" + time.Now().UTC().Format(time.DateOnly)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	require.Equal(t, `attachment; filename="`+name+`.zip"`, w.Header().Get("Content-Disposition"))
	archive, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	require.Equal(t, name+".sqlite", archive.File[0].Name)
	entry, err := archive.File[0].Open()
	require.NoError(t, err)
	db, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(db), "SQLite format 3\x00"), "the zip holds a SQLite DB")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_db_downloaded_by_control", events[0].Name)
	require.Equal(t, `{"by":"admin"}`, *events[0].Attributes)
}
