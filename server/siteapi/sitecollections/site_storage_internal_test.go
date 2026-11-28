package sitecollections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's admin sees its DB's size and tables on its Usage page, with no button to truncate it; a user does not
func TestUsage(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	_, err := site.SiteDB.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: time.Now().UnixMilli(), CreatedByIdentityName: "phone"})
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
	require.NotContains(t, w.Body.String(), `id="site--usage--truncate"`)
	require.Contains(t, w.Body.String(), `href="/usage/download"`)
}
