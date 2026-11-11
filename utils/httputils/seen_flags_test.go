package httputils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// An identity is recorded as seen even when its client stops waiting once its response is sent, as htmx cancels a
// request a newer one replaces, rather than failing to record it on the cancelled request
func TestSeenFlagsRecordedWhenClientLeaves(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	siteModel, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	identities, err := controlDB.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	identity := identities[0]
	require.Nil(t, identity.LastSeenAt)

	requestCtx, cancel := context.WithCancel(context.Background())
	requestCtx = httputils.CreateSiteContext(requestCtx, nil, controlDB, site.SiteDB, siteModel, siteID, "", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	requestCtx = httputils.SetIdentityModelInSiteContext(requestCtx, identity)
	handler := httputils.SiteIdentitySeenFlagsUpdateMiddleware(controlDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		cancel() // the client goes once its response is sent
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil).WithContext(requestCtx))

	identities, err = controlDB.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteID})
	require.NoError(t, err)
	for _, seen := range identities {
		if seen.ID == identity.ID {
			require.NotNil(t, seen.LastSeenAt, "the identity is recorded as seen")
		}
	}
}
