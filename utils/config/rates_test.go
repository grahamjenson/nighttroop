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
