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

// TestDeleteSite checks only a site's admin deletes it, which archives it and signs them out, and only with S3 to keep
// its archive, which the browser tests do not have
func TestDeleteSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewReplicatedControlDB(t)
	siteID := site.SiteDB.Subdomain
	deleteAs := func(role consts.IdentityRole, s3Working bool) *httptest.ResponseRecorder {
		cfg.Status.S3Working = s3Working
		ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, siteID, "", httputils.HTTPActionWrite, httputils.AcceptTypeHTMX, httputils.ContentTypeForm)
		ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Name: "someone", Role: role})
		w := httptest.NewRecorder()
		DeleteSiteHandler(w, httptest.NewRequest(http.MethodPost, "/settings/site/delete", nil).WithContext(ctx))
		return w
	}
	state := func() string {
		row, err := controlDB.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: siteID})
		require.NoError(t, err)
		return row.State
	}

	require.Equal(t, http.StatusForbidden, deleteAs(consts.IdentityRoleDevicePWDPhone, true).Code)
	w := deleteAs(consts.IdentityRoleAdmin, false)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "a site can only be deleted when its data can be archived to S3")
	require.Equal(t, "live", state())

	w = deleteAs(consts.IdentityRoleAdmin, true)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, cfg.HostURLWithPort, w.Header().Get("HX-Redirect"))
	require.Contains(t, w.Header().Get("Set-Cookie"), "Max-Age=0", "the admin is signed out")
	require.Equal(t, "archived", state())
	_, _, err := controlDB.GetSite(t.Context(), siteID)
	require.Error(t, err, "it is no longer served")

	events, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_archived", events[0].Name)
	require.Equal(t, `{"by":"site admin"}`, *events[0].Attributes)
}
