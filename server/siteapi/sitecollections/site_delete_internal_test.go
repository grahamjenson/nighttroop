package sitecollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestDeleteSite checks only a site's admin deletes it, which archives it and signs them out, and only with an archive
// to upload to, which the browser tests do not have
func TestDeleteSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	uploaded := ""
	upload := func(key string, _ []string) error { uploaded = key; return nil }
	deleteAs := func(role consts.IdentityRole, upload func(string, []string) error) *httptest.ResponseRecorder {
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeHTMX, httputils.ContentTypeForm)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		w := httptest.NewRecorder()
		deleteSite(w, httptest.NewRequest(http.MethodPost, "/settings/site/delete", nil).WithContext(ctx), upload)
		return w
	}
	state := func() string {
		row, err := controlDB.Queries.GetSiteInAnyState(context.TODO(), controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return row.State
	}

	require.Equal(t, http.StatusForbidden, deleteAs(consts.IdentityRoleDevicePWDPhone, upload).Code)
	w := deleteAs(consts.IdentityRoleAdmin, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "a site can only be deleted when its data can be archived to S3")
	require.Equal(t, "live", state())

	w = deleteAs(consts.IdentityRoleAdmin, upload)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, cfg.HostURLWithPort, w.Header().Get("HX-Redirect"))
	require.Contains(t, w.Header().Get("Set-Cookie"), "Max-Age=0", "the admin is signed out")
	require.NotEmpty(t, uploaded)
	require.Equal(t, "archived", state())
	_, _, err := controlDB.GetSite(siteID)
	require.Error(t, err, "it is no longer served")

	events, err := controlDB.Logs.Queries.SearchEvents(context.TODO(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_archived", events[0].Name)
	require.Equal(t, `{"by":"site admin"}`, *events[0].Attributes)
}
