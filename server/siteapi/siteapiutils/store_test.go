package siteapiutils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitedata"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type noSiteStore struct{}

func (noSiteStore) GetName() string                                           { return "glucose_readings" }
func (noSiteStore) ToModel(row *struct{}, _ *httputils.SiteContext) *struct{} { return row }
func (noSiteStore) GetRow(context.Context, *sitedb.Queries, string, *httputils.SiteContext) (*struct{}, error) {
	return nil, nil
}
func (noSiteStore) SearchRows(context.Context, *sitedb.Queries, *apiutils.Query, *httputils.SiteContext) ([]*struct{}, error) {
	return nil, nil
}
func (noSiteStore) SearchHTMXResponse(models []*struct{}, query *apiutils.Query) templ.Component {
	return nil
}

// A request that reaches a store without a site is not found, rather than a panic
func TestHandlerWithoutSiteContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/data/glucose", nil)
	r.Header.Set("Accept", "application/json")
	siteapiutils.SearchHandler(nil, nil, siteapiutils.SiteRows[struct{}, struct{}](noSiteStore{}))(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// A store routed to a handler it cannot do fails when the router is made, not on a request
func TestHandlerForWhatTheStoreCannotDo(t *testing.T) {
	store := siteapiutils.SiteRows[struct{}, struct{}](noSiteStore{})
	require.NotPanics(t, func() { siteapiutils.SearchHandler(nil, nil, store) })
	require.PanicsWithValue(t, "the glucose_readings store cannot create", func() { siteapiutils.CreateHandler(nil, nil, store) })
	require.PanicsWithValue(t, "the glucose_readings store cannot delete", func() { siteapiutils.DeleteHandler(nil, nil, store) })
}

// routes are the method and pattern of each route on the router
func routes(t *testing.T, r chi.Router) []string {
	var found []string
	require.NoError(t, chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		found = append(found, method+" "+route)
		return nil
	}))
	return found
}

// Routes registers a route for each thing the store can do, and none for what it cannot
func TestRoutes(t *testing.T) {
	r := chi.NewRouter()
	siteapiutils.Routes(r, "/data/glucose", nil, nil, siteapiutils.SiteRows[sitedb.GlucoseReading, sitemodels.SiteGlucoseReading](&sitedata.GlucoseReadingsStore{}))
	require.ElementsMatch(t, []string{"GET /data/glucose", "POST /data/glucose", "GET /data/glucose/{id}", "POST /data/glucose/{id}", "DELETE /data/glucose/{id}"}, routes(t, r))

	r = chi.NewRouter()
	siteapiutils.Routes(r, "/data/event-groups", nil, nil, siteapiutils.SiteRows[sitedb.EventGroupsBundle, sitemodels.SiteEventGroup](&sitedata.EventGroupsStore{}))
	require.ElementsMatch(t, []string{"GET /data/event-groups", "GET /data/event-groups/{id}", "DELETE /data/event-groups/{id}"}, routes(t, r))

	r = chi.NewRouter()
	siteapiutils.Routes(r, "/data/glucose", nil, nil, siteapiutils.SiteRows[struct{}, struct{}](noSiteStore{}))
	require.Equal(t, []string{"GET /data/glucose"}, routes(t, r))
}
