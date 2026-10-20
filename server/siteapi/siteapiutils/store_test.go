package siteapiutils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type noSiteStore struct{}

func (noSiteStore) SearchModels(ctx context.Context, tx *siteapiutils.StoreTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*struct{}, *apiutils.HTTPError) {
	return nil, nil
}
func (noSiteStore) SearchHTMXResponse(models []*struct{}, query *apiutils.Query) templ.Component {
	return nil
}
func (noSiteStore) GetName() string { return "glucose_readings" }

// A request that reaches a store without a site is not found, rather than a panic
func TestHandlerWithoutSiteContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/data/glucose", nil)
	r.Header.Set("Accept", "application/json")
	siteapiutils.SearchHandler(nil, nil, noSiteStore{})(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
