package httputils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A large batch spends a write from its site and identity for each chunk of records, and none unless both have them
func TestExtraWritesWithinRateLimit(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	identity := &controldb.SiteIdentity{ID: -7}
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, "batchzzz", "10.0.0.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	r := httptest.NewRequest("POST", "/api/v1/entries", nil).WithContext(httputils.SetIdentityModelInSiteContext(ctx, identity))
	// The default write limits, 60 a minute with a burst of 240, as the test config's are far higher
	site, ident := cfg.RateLimiters.SitePOSTLimiter("batchzzz"), cfg.RateLimiters.IdentityPOSTLimiter(identity.ID)
	for _, limiter := range []*rate.Limiter{site, ident} {
		limiter.SetLimit(1)
		limiter.SetBurst(240)
	}

	// A batch of 5000 records costs 49 more writes than the request's own, from both
	require.True(t, httputils.ExtraWritesWithinRateLimit(r, 49))
	require.InDelta(t, 191, site.Tokens(), 1)
	require.InDelta(t, 191, ident.Tokens(), 1)

	// Once the identity has too few, the batch is refused and the site's are not spent
	ident.AllowN(time.Now(), 180)
	require.False(t, httputils.ExtraWritesWithinRateLimit(r, 49))
	require.InDelta(t, 191, site.Tokens(), 1)
	require.InDelta(t, 11, ident.Tokens(), 1)

	// A batch of one chunk costs nothing more
	require.True(t, httputils.ExtraWritesWithinRateLimit(r, 0))
}

// A request over its identity's limit is refused with a Retry-After of when one is let through again, at least a second
func TestIdentityRateLimitSaysWhenToRetry(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	identity := &controldb.SiteIdentity{ID: -8}
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, "retryzzz", "10.0.0.2", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	r := httptest.NewRequest("POST", "/api/v3/treatments", nil).WithContext(httputils.SetIdentityModelInSiteContext(ctx, identity))
	limiter := cfg.RateLimiters.IdentityPOSTLimiter(identity.ID)
	limiter.SetLimit(rate.Limit(1.0 / 30))
	limiter.SetBurst(1)
	limited := httputils.SiteRateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	first := httptest.NewRecorder()
	limited.ServeHTTP(first, r)
	require.Equal(t, http.StatusOK, first.Code)
	require.Empty(t, first.Header().Get("Retry-After"))

	refused := httptest.NewRecorder()
	limited.ServeHTTP(refused, r)
	require.Equal(t, http.StatusTooManyRequests, refused.Code)
	require.Contains(t, []string{"29", "30"}, refused.Header().Get("Retry-After"))
}

// A request its identity's limit refuses spends none of its site's, so a follower's app over its own limit cannot
// use up the limit the site's other identities and apps share
func TestIdentityOverItsLimitSpendsNoneOfItsSites(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	limiters := cfg.RateLimiters
	limited := httputils.SiteRateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	for _, test := range []struct {
		action         httputils.HTTPAction
		siteID         string
		identityID     int64
		site, identity *rate.Limiter
	}{
		{httputils.HTTPActionRead, "sharedreadzzz", -20, limiters.SiteGETLimiter("sharedreadzzz"), limiters.IdentityGETLimiter(-20)},
		{httputils.HTTPActionWrite, "sharedwritezzz", -21, limiters.SitePOSTLimiter("sharedwritezzz"), limiters.IdentityPOSTLimiter(-21)},
	} {
		test.site.SetLimit(rate.Limit(1.0 / 60))
		test.site.SetBurst(5)
		test.identity.SetLimit(rate.Limit(1.0 / 60))
		test.identity.SetBurst(1)
		ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, test.siteID, "10.0.0.4", test.action, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		r := httptest.NewRequest("GET", "/api/v1/entries", nil).WithContext(httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{ID: test.identityID}))

		codes := []int{}
		for range 4 {
			w := httptest.NewRecorder()
			limited.ServeHTTP(w, r)
			codes = append(codes, w.Code)
		}
		require.Equal(t, []int{http.StatusOK, http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusTooManyRequests}, codes)
		// Only the first spent one of the site's, so its other identities have the rest
		require.InDelta(t, 4, test.site.Tokens(), 0.1)
	}
}
