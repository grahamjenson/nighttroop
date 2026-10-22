package siteapiutils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
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
