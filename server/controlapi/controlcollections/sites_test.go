package controlcollections

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
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

// TestDisablingActiveSite checks an admin can disable a live site however recently it was written to or seen, but not
// one already disabled, and enable a disabled site, but not a live one
func TestDisablingActiveSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	cfg.DisableAfterInactivityDays = 7
	t.Cleanup(func() { cfg.DisableAfterInactivityDays = 0 })
	ctx, siteID, now := context.Background(), site.SiteDB.Subdomain, time.Now().UnixMilli()
	controlContext := &httputils.ControlContext{ControlDB: controlDB, IdentityModel: &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}}

	active := &controlmodels.Site{ID: siteID, State: "live", CreatedAt: now, LastSeenAt: &now, LastWriteAt: &now}
	require.False(t, active.Inactive(7, time.Now()))
	_, err := (&SitesCollection{}).UpdateRow(ctx, controlContext, active, &controlmodels.Site{State: "disabled"})
	require.NoError(t, err)

	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "disabled", row.State)

	_, err = (&SitesCollection{}).UpdateRow(ctx, controlContext, active, &controlmodels.Site{State: "disabled"})
	require.Error(t, err, "a site already disabled is not disabled again")

	_, err = (&SitesCollection{}).UpdateRow(ctx, controlContext, active, &controlmodels.Site{State: "live"})
	require.NoError(t, err)
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "live", row.State)
	require.Nil(t, row.DisabledAt)
	_, err = (&SitesCollection{}).UpdateRow(ctx, controlContext, active, &controlmodels.Site{State: "live"})
	require.ErrorContains(t, err, "only a disabled site can be enabled")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_enabled", events[0].Name)
}

func TestTruncateHandler(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.Post("/sites/{id}/truncate", (&SitesCollection{}).TruncateHandler)
	truncate := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/sites/"+siteID+"/truncate", nil)
		r.Header.Set("Accept", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	// Any site whose DB is here can be truncated by an admin, within its DB size limit or over it
	size := int64(1000)
	require.NoError(t, controlDB.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: &size, ID: siteID}))
	require.Equal(t, http.StatusOK, truncate().Code, "a site within its limit is truncated")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_truncated", events[0].Name)
	require.Equal(t, `{"by":"admin","before_size_kb":1000,"after_size_kb":null}`, *events[0].Attributes)

	// An archived site has no DB here to truncate
	_, err = controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)
	_, err = controlDB.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: `["` + siteID + `"]`})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, truncate().Code, "an archived site is not truncated")
}

// TestSettingsHandler checks an admin can turn public site creation off and on
func TestSettingsHandler(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	settings := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(body))
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
		w := httptest.NewRecorder()
		(&SitesCollection{}).SettingsHandler(w, r.WithContext(rctx))
		return w
	}

	require.True(t, controlDB.PublicSiteCreation(t.Context()), "the public can create sites on a new server")

	off := settings(`{"public_site_creation": false}`)
	require.Equal(t, http.StatusOK, off.Code)
	require.JSONEq(t, `{"status": 200, "result": {"public_site_creation": false}}`, off.Body.String())
	require.False(t, controlDB.PublicSiteCreation(t.Context()))

	require.Equal(t, http.StatusOK, settings(`{"public_site_creation": true}`).Code)
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

// An admin uploads a DB in place of an unavailable site's, which is served again, with an event saying so; a form with
// no DB, or a site that is not unavailable, has none put in place
func TestUploadDBHandler(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.Post("/sites/{id}/db", (&SitesCollection{}).UploadDBHandler)
	upload := func(field string) *httptest.ResponseRecorder {
		body := &bytes.Buffer{}
		form := multipart.NewWriter(body)
		part, err := form.CreateFormFile(field, siteID+".sqlite")
		require.NoError(t, err)
		_, err = part.Write([]byte("SQLite format 3\x00"))
		require.NoError(t, err)
		require.NoError(t, form.Close())
		r := httptest.NewRequest(http.MethodPost, "/sites/"+siteID+"/db", body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("Accept", "application/json")
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionWrite)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	require.Equal(t, http.StatusBadRequest, upload("db").Code, "a site that is not unavailable")
	controlDB.MakeSiteUnavailable(siteID, db.UnavailableCannotOpen)
	require.Equal(t, http.StatusBadRequest, upload("file").Code, "a form with no DB")
	w := upload("db")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Empty(t, controlDB.UnavailableSites(), "it is served again")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_db_uploaded", events[0].Name)
	require.Equal(t, `{"by":"admin","reason":"cannot_open"}`, *events[0].Attributes)
}

// An admin downloads the zip of a site's DB its admins get from Download Data, which its admins see in its events, but
// not of a site with no DB here
func TestDownloadDBHandler(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := chi.NewRouter()
	router.Get("/sites/{id}/db", (&SitesCollection{}).DownloadDBHandler)
	download := func(id string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/sites/"+id+"/db", nil)
		rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionRead)
		rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(rctx))
		return w
	}

	require.Equal(t, http.StatusNotFound, download("nosuchsite").Code)

	w := download(siteID)
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
	copied, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(copied), "SQLite format 3\x00"), "the zip holds a SQLite DB")

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_db_downloaded_by_control", events[0].Name)
	require.Equal(t, `{"by":"admin"}`, *events[0].Attributes)

	// An unavailable site with no DB here has none to download
	controlDB.MakeSiteUnavailable(siteID, db.UnavailableMissing)
	require.Equal(t, http.StatusNotFound, download(siteID).Code)
}

// An admin downloads the DB here of a site unavailable as its migration failed, as it is, to fix it
func TestDownloadDBHandlerOfAnUnavailableSite(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	_, err = controlDB.CreateSiteWithID(t.Context(), "migrfail")
	require.NoError(t, err)
	controlDB.MakeSiteUnavailable("migrfail", db.UnavailableMigrationFailed)

	router := chi.NewRouter()
	router.Get("/sites/{id}/db", (&SitesCollection{}).DownloadDBHandler)
	r := httptest.NewRequest(http.MethodGet, "/sites/migrfail/db", nil)
	rctx := httputils.CreateControlContext(r.Context(), controlDB, "", httputils.HTTPActionRead)
	rctx = httputils.SetAdminControlContext(rctx, &controldb.ControlIdentity{Name: "admin", Role: consts.ControlRoleAdmin}, config.ServerStatus{})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r.WithContext(rctx))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	archive, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	require.Contains(t, controlDB.UnavailableSites(), "migrfail", "it stays unavailable")
}
