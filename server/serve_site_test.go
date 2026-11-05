package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A request for a site is served with its context, one for a site there is not is 404, and one whose site cannot be
// looked up, here as its request has ended, is 503, so an uploader tries again rather than take its URL as wrong
func TestServeSite(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, httputils.GetSiteContext(r.Context()).SiteID)
	})
	ended, end := context.WithCancel(t.Context())
	end()

	tests := []struct {
		name      string
		ctx       context.Context
		subdomain string
		status    int
		body      string
	}{
		{"a site", t.Context(), siteID, http.StatusOK, siteID},
		{"no site", t.Context(), "zzzzzzzz", http.StatusNotFound, "404 page not found\n"},
		{"cannot look up", ended, siteID, http.StatusServiceUnavailable, "Server busy, try again shortly\n"},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		serveSite(tt.ctx, w, httptest.NewRequest(http.MethodGet, "/", nil), router, config, controlDB, tt.subdomain, "192.0.2.1")
		assert.Equal(t, tt.status, w.Code, tt.name)
		assert.Equal(t, tt.body, w.Body.String(), tt.name)
	}
}
