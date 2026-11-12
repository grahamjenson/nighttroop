package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A request for a site is served with its context, one for a site there is not is 404 saying so on its pages and API,
// but 410 on its Nightscout API, so an uploader can tell a site gone from a wrong path, and one whose site cannot be
// looked up, here as its request has ended, is 503, so an uploader tries again rather than take its URL as wrong.
// A site disabled, archived or deleted is answered just as one that never was, so the answer tells nothing about it.
func TestServeSite(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, httputils.GetSiteContext(r.Context()).SiteID)
	})
	ended, end := context.WithCancel(ctx)
	end()

	newSite := func() string {
		created, err := controlDB.CreateSite(ctx)
		require.NoError(t, err)
		return created.SiteDB.Subdomain
	}
	disabled, archived, deleted := newSite(), newSite(), newSite()
	for _, id := range []string{disabled, archived} {
		_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: id})
		require.NoError(t, err)
	}
	require.NoError(t, controlDB.ArchiveSite(ctx, archived, func(context.Context, string, s3utils.Retention, []string) error { return nil }))
	require.NoError(t, controlDB.DeleteSites(ctx, []string{deleted}))

	notFound := "{\"message\":\"site does not exist\",\"status\":404}\n"
	gone := "{\"message\":\"site does not exist\",\"status\":410}\n"
	tests := []struct {
		name          string
		ctx           context.Context
		subdomain     string
		missingStatus int
		status        int
		body          string
	}{
		{"a site", ctx, siteID, http.StatusNotFound, http.StatusOK, siteID},
		{"a site's Nightscout API", ctx, siteID, http.StatusGone, http.StatusOK, siteID},
		{"no site", ctx, "zzzzzzzz", http.StatusNotFound, http.StatusNotFound, notFound},
		{"no site's Nightscout API", ctx, "zzzzzzzz", http.StatusGone, http.StatusGone, gone},
		{"a disabled site's Nightscout API", ctx, disabled, http.StatusGone, http.StatusGone, gone},
		{"an archived site's Nightscout API", ctx, archived, http.StatusGone, http.StatusGone, gone},
		{"a deleted site's Nightscout API", ctx, deleted, http.StatusGone, http.StatusGone, gone},
		{"a deleted site", ctx, deleted, http.StatusNotFound, http.StatusNotFound, notFound},
		{"cannot look up", ended, siteID, http.StatusGone, http.StatusServiceUnavailable, "Server busy, try again shortly\n"},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		serveSite(tt.ctx, w, httptest.NewRequest(http.MethodGet, "/", nil), router, config, controlDB, tt.subdomain, "192.0.2.1", tt.missingStatus)
		assert.Equal(t, tt.status, w.Code, tt.name)
		assert.Equal(t, tt.body, w.Body.String(), tt.name)
	}
}
