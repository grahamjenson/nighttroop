package httputils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// flaggedTestRequest is a request to a site whose last recorded DB size is sizeKB, with a 20 MB limit
func flaggedTestRequest(cfg *config.ServerConfig, siteID string, sizeKB int64, action httputils.HTTPAction, siteDB *db.SiteDB) *http.Request {
	limit := int64(20480)
	site := &controldb.Site{ID: siteID, MaxDbSizeKbLimit: &limit, EstimatedDbSizeKb: &sizeKB}
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, siteDB, site, siteID, "10.0.0.3", action, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	return httptest.NewRequest("POST", "/api/v1/entries", nil).WithContext(httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{ID: -9}))
}

func flaggedTestContext(cfg *config.ServerConfig, siteID string, sizeKB int64, siteDB *db.SiteDB) *httputils.SiteContext {
	return httputils.GetSiteContext(flaggedTestRequest(cfg, siteID, sizeKB, httputils.HTTPActionWrite, siteDB).Context())
}

// Sizes, in KB, of sites at each abuse level from their size alone, with a 20 MB limit
const (
	fairKB    = 39000 // 1.9 times its limit
	overUseKB = 45000 // 2.2 times
	abusiveKB = 55300 // 2.7 times
)

// A site at abuse level OverUse spends its flagged write limit too, so it is refused far sooner than one under it
func TestFlaggedSiteWritesAreLimited(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	limited := httputils.SiteRateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serve := func(r *http.Request) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		limited.ServeHTTP(w, r)
		return w
	}
	for _, siteID := range []string{"flaggedzzz", "honestzzz"} {
		limiter := cfg.RateLimiters.SiteFlaggedPOSTLimiter(siteID)
		limiter.SetLimit(rate.Limit(1.0 / 30))
		limiter.SetBurst(3)
	}
	flagged := flaggedTestRequest(cfg, "flaggedzzz", overUseKB, httputils.HTTPActionWrite, nil)

	for range 3 {
		require.Equal(t, http.StatusOK, serve(flagged).Code)
	}
	refused := serve(flagged)
	require.Equal(t, http.StatusTooManyRequests, refused.Code)
	require.Contains(t, []string{"29", "30"}, refused.Header().Get("Retry-After"))
	require.Equal(t, httputils.SiteOverStorageMessage, refused.Body.String())
	// Its reads are not limited by it
	require.Equal(t, http.StatusOK, serve(flaggedTestRequest(cfg, "flaggedzzz", overUseKB, httputils.HTTPActionRead, nil)).Code)

	// A site at Fair spends none of it
	honest := flaggedTestRequest(cfg, "honestzzz", fairKB, httputils.HTTPActionWrite, nil)
	for range 10 {
		require.Equal(t, http.StatusOK, serve(honest).Code)
	}
	require.InDelta(t, 3, cfg.RateLimiters.SiteFlaggedPOSTLimiter("honestzzz").Tokens(), 0.1)
}

// A batch of a site at OverUse spends a flagged write for each record, and none unless it has them all
func TestFlaggedSiteBatchCostsEachRecord(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	limiter := cfg.RateLimiters.SiteFlaggedPOSTLimiter("batchflagzzz")
	limiter.SetLimit(rate.Limit(1.0 / 30))
	limiter.SetBurst(36)
	flagged := flaggedTestContext(cfg, "batchflagzzz", overUseKB, nil)

	allowed, _ := httputils.FlaggedSiteWritesWithinRateLimit(flagged, 30)
	require.True(t, allowed)
	require.InDelta(t, 6, limiter.Tokens(), 0.1)
	allowed, seconds := httputils.FlaggedSiteWritesWithinRateLimit(flagged, 10)
	require.False(t, allowed)
	require.InDelta(t, 120, seconds, 1)
	require.InDelta(t, 6, limiter.Tokens(), 0.1)
	// A batch past its burst is never let through
	allowed, seconds = httputils.FlaggedSiteWritesWithinRateLimit(flagged, 99)
	require.False(t, allowed)
	require.Zero(t, seconds)

	// A batch of 100 from a site at Fair costs none
	allowed, _ = httputils.FlaggedSiteWritesWithinRateLimit(flaggedTestContext(cfg, "batchhonestzzz", fairKB, nil), 99)
	require.True(t, allowed)
}

// Each write of a site at Abusive or above costs 4 flagged writes
func TestAbusiveSiteWritesCostFour(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	for sizeKB, siteID := range map[int64]string{abusiveKB: "abusivezzz", 1 << 20: "blockedzzz"} {
		limiter := cfg.RateLimiters.SiteFlaggedPOSTLimiter(siteID)
		limiter.SetLimit(rate.Limit(1.0 / 30))
		limiter.SetBurst(36)
		abusive := flaggedTestContext(cfg, siteID, sizeKB, nil)

		allowed, _ := httputils.FlaggedSiteWritesWithinRateLimit(abusive, 5)
		require.True(t, allowed)
		require.InDelta(t, 16, limiter.Tokens(), 0.1)
		allowed, _ = httputils.FlaggedSiteWritesWithinRateLimit(abusive, 5)
		require.False(t, allowed)
		require.InDelta(t, 16, limiter.Tokens(), 0.1)
	}
}

// The level is the higher of the level last stored and the level of the DB's size on disk now, as the stored score
// can be minutes old
func TestSiteAbuseLevelUsesLiveSize(t *testing.T) {
	cfg := testutils.GetTestServerConfig()
	path := filepath.Join(t.TempDir(), "site.db")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	siteDB := &db.SiteDB{DB: &sitedb.DB{Path: path}}

	// Recorded as far over, but small now, as it was truncated
	require.Equal(t, controldb.AbuseGood, httputils.SiteAbuseLevel(flaggedTestContext(cfg, "livezzz", 1<<20, siteDB)))
	// Recorded as small, but grown since, and the context's model is not changed
	require.NoError(t, os.Truncate(path, abusiveKB*1024))
	siteContext := flaggedTestContext(cfg, "livezzz", 10, siteDB)
	require.Equal(t, controldb.AbuseAbusive, httputils.SiteAbuseLevel(siteContext))
	require.Equal(t, int64(10), *siteContext.SiteModel.EstimatedDbSizeKb)

	// Small now, but stored as OverUse, from signals other than its size
	siteContext.SiteModel.AbuseScore = 0.45
	require.NoError(t, os.Truncate(path, 0))
	require.Equal(t, controldb.AbuseOverUse, httputils.SiteAbuseLevel(siteContext))
}
