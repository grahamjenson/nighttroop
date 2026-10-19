package config

import (
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
