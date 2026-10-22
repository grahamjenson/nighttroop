package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSiteUpdateSyncLimiterIsSeparateFromPOST(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SitePOSTPerMinuteLimit: 60, SitePOSTPerMinuteBurst: 100,
		SiteUpdateSyncPerMinuteLimit: 1, SiteUpdateSyncPerMinuteBurst: 1,
	})

	// The site POST limiter is made first, as the middleware runs before the sync store
	require.True(t, limiters.SitePOSTLimiter("site").Allow())

	// The sync limiter has its own, stricter, limit
	require.True(t, limiters.SiteUpdateSyncLimiter("site").Allow())
	require.False(t, limiters.SiteUpdateSyncLimiter("site").Allow())
	require.True(t, limiters.SitePOSTLimiter("site").Allow())
}

// A login name comes straight from the request body, so a huge one must not pin a huge key in the LRU
func TestSiteLoginAllowKeysAreFixedSize(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SiteLoginPerMinuteLimit: 1, SiteLoginPerMinuteBurst: 10,
	})

	huge := strings.Repeat("a", 500*1024)
	allowed, _ := limiters.SiteLoginAllow("site:"+huge, huge)
	require.True(t, allowed)

	keys := append(limiters.userLoginRateLimiters.Keys(), limiters.ipLoginRateLimiters.Keys()...)
	require.Len(t, keys, 2)
	for _, key := range keys {
		require.Len(t, key, 64, "limiter keys are a sha256 hex digest, not the input")
	}
}
