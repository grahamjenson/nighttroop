package config

import (
	"strings"
	"testing"
	"time"

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

// A host is usually given a whole IPv6 /64, so an IP's limits hold for every address in it, or stepping through them
// would give it fresh limits on every request
func TestIPLimitersKeyOnIPv6Slash64(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		GlobalIPPerMinuteLimit: 1, GlobalIPPerMinuteBurst: 1,
		SiteLoginPerMinuteLimit: 1, SiteLoginPerMinuteBurst: 1,
		SiteUnauthPerMinuteLimit: 1, SiteUnauthPerMinuteBurst: 1,
		SiteCreateIPPerMinuteLimit: 1, SiteCreateIPPerMinuteBurst: 1,
	})
	limits := map[string]func(ip string) bool{
		"global":      func(ip string) bool { return limiters.GetGlobalIPRateLimiter(ip).Allow() },
		"site create": func(ip string) bool { return limiters.GetSiteCreateIPRateLimiter(ip).Allow() },
		"site unauth": func(ip string) bool { return limiters.SiteUnauthLimiter("site", ip).Allow() },
		// Each login is for a new name, so only the IP's limiter can refuse it
		"site login": func(ip string) bool { allowed, _ := limiters.SiteLoginAllow("site:"+ip, ip); return allowed },
	}
	tests := []struct {
		ip      string
		allowed bool
	}{
		{"2001:db8:1:2::1", true},
		{"2001:db8:1:2::2", false},
		{"2001:db8:1:2:ffff:ffff:ffff:ffff", false},
		{"2001:db8:1:3::1", true},
		{"192.0.2.1", true},
		{"::ffff:192.0.2.1", false},
		{"192.0.2.2", true},
	}
	for name, allow := range limits {
		for _, test := range tests {
			require.Equal(t, test.allowed, allow(test.ip), "%s limit for %s", name, test.ip)
		}
	}
}

// Stats counts the limiters kept of each kind, how full the most used is and how many would refuse, without spending any
func TestRateLimiterStats(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10, GlobalRateLimit: 60, GlobalRateBurst: 10,
		SiteGETPerMinuteLimit: 60, SiteGETPerMinuteBurst: 4,
	})
	now := time.Now()
	for range 4 {
		require.True(t, limiters.SiteGETLimiter("busy").AllowN(now, 1))
	}
	require.True(t, limiters.SiteGETLimiter("quiet").AllowN(now, 1))

	byName := map[string]RateLimiterStats{}
	for _, s := range limiters.Stats(now) {
		byName[s.Name] = s
	}
	require.Len(t, byName, 15)
	require.Equal(t, RateLimiterStats{Name: "site_get", Per: "site", PerMinute: 60, Burst: 4, Tracked: 2, Limited: 1, FullestPercent: 100},
		byName["site_get"], "busy has spent its burst, quiet a quarter")
	require.Equal(t, RateLimiterStats{Name: "global", Per: "server", PerMinute: 60, Burst: 10, Tracked: 1}, byName["global"])
	require.Equal(t, RateLimiterStats{Name: "site_post", Per: "site"}, byName["site_post"], "none kept yet")

	// Reading them spent nothing, so a second later busy has one of its 4 back and would not refuse
	for _, s := range limiters.Stats(now.Add(time.Second)) {
		if s.Name == "site_get" {
			require.Zero(t, s.Limited)
			require.InDelta(t, 75, s.FullestPercent, 0.01)
		}
	}
}
