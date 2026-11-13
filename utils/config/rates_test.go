package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSiteUpdateConnectorsLimiterIsSeparateFromPOST(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SitePOSTPerMinuteLimit: 60, SitePOSTPerMinuteBurst: 100,
		SiteUpdateConnectorsPerMinuteLimit: 1, SiteUpdateConnectorsPerMinuteBurst: 1,
	})

	// The site POST limiter is made first, as the middleware runs before the sync store
	require.True(t, limiters.SitePOSTLimiter("site").Allow())

	// The sync limiter has its own, stricter, limit
	require.True(t, limiters.SiteUpdateConnectorsLimiter("site").Allow())
	require.False(t, limiters.SiteUpdateConnectorsLimiter("site").Allow())
	require.True(t, limiters.SitePOSTLimiter("site").Allow())
}

// A login name comes straight from the request body, so a huge one must not pin a huge key in the LRU
func TestSiteLoginAllowKeysAreFixedSize(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SiteLoginPerMinuteLimit: 1, SiteLoginPerMinuteBurst: 10, SiteLoginNamePerMinuteLimit: 30, SiteLoginNamePerMinuteBurst: 60,
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

// An IP using up its own login attempts at a name, as one locking its owner out would, does not stop another IP
// logging in to it. A name's own limit, across every IP, only caps guessing spread over many, and an attempt one
// limiter refuses spends nothing from the other.
func TestSiteLoginAllowOneIPCannotLockOthersOut(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SiteLoginPerMinuteLimit: 1, SiteLoginPerMinuteBurst: 3, SiteLoginNamePerMinuteLimit: 1, SiteLoginNamePerMinuteBurst: 6,
	})
	attacker, owner := "203.0.113.1", "203.0.113.2"
	tests := []struct {
		name, ip, login string
		allowed         bool
		remaining       int
	}{
		{"the attacker's first guess", attacker, "site:admin", true, 3},
		{"its second", attacker, "site:admin", true, 2},
		{"its third", attacker, "site:admin", true, 1},
		{"its IP has none left", attacker, "site:admin", false, 0},
		{"nor at another name", attacker, "site:phone", false, 0},
		{"refused guesses spend none of the name's", attacker, "site:admin", false, 0},
		{"so the owner still logs in", owner, "site:admin", true, 3},
		{"as at a site's passkeys", owner, "site:passkey", true, 2},
		{"and its recovery", owner, "site::backup_key", true, 1},
		{"another IP spends one of the name's last two", "203.0.113.3", "site:admin", true, 2},
		{"and another the last", "203.0.113.4", "site:admin", true, 1},
		{"then guessing spread over many IPs is capped", "203.0.113.5", "site:admin", false, 0},
		{"which spent none of that IP's", "203.0.113.5", "site:phone", true, 3},
		{"while other sites' names are not capped", "203.0.113.5", "other:admin", true, 2},
	}
	for _, tt := range tests {
		allowed, remaining := limiters.SiteLoginAllow(tt.login, tt.ip)
		require.Equal(t, tt.allowed, allowed, "%s: %s from %s", tt.name, tt.login, tt.ip)
		require.Equal(t, tt.remaining, remaining, "%s: remaining", tt.name)
	}
}

// A host is usually given a whole IPv6 /64, so an IP's limits hold for every address in it, or stepping through them
// would give it fresh limits on every request
func TestIPLimitersKeyOnIPv6Slash64(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		GlobalIPPerMinuteLimit: 1, GlobalIPPerMinuteBurst: 1,
		SiteLoginPerMinuteLimit: 1, SiteLoginPerMinuteBurst: 1, SiteLoginNamePerMinuteLimit: 30, SiteLoginNamePerMinuteBurst: 60,
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
	// A site's limit is SiteIdentitiesShare identities' worth: 240 a minute and a burst of 16
	for range 16 {
		require.True(t, limiters.SiteGETLimiter("busy").AllowN(now, 1))
	}
	require.True(t, limiters.SiteGETLimiter("quiet").AllowN(now, 1))

	byName := map[string]RateLimiterStats{}
	for _, s := range limiters.Stats(now) {
		byName[s.Name] = s
	}
	require.Len(t, byName, 17)
	require.Equal(t, RateLimiterStats{Name: "site_get", Per: "site", PerMinute: 240, Burst: 16, Tracked: 2, Limited: 1, FullestPercent: 100},
		byName["site_get"], "busy has spent its burst, quiet a sixteenth")
	require.Equal(t, RateLimiterStats{Name: "global", Per: "server", PerMinute: 60, Burst: 10, Tracked: 1}, byName["global"])
	require.Equal(t, RateLimiterStats{Name: "site_post", Per: "site"}, byName["site_post"], "none kept yet")

	// Reading them spent nothing, so a second later busy has 4 of its 16 back and would not refuse
	for _, s := range limiters.Stats(now.Add(time.Second)) {
		if s.Name == "site_get" {
			require.Zero(t, s.Limited)
			require.InDelta(t, 75, s.FullestPercent, 0.01)
		}
	}
}

// A rate a control admin sets is used by the limiters already kept and those made after, until it is set back
func TestSetRatesChangesEveryLimiterOfItsKind(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{
		MaxSites: 10, MaxIdentitiesPerSite: 10,
		SiteDownloadPerMinuteLimit: 1, SiteDownloadPerMinuteBurst: 1, SitePOSTPerMinuteLimit: 60, SitePOSTPerMinuteBurst: 10,
	})
	kept := limiters.SiteDownloadLimiter("kept")
	require.True(t, kept.Allow())
	require.False(t, kept.Allow())

	tests := []struct {
		name  string
		rates map[string]Rate
		set   map[string]Rate
		burst int
	}{
		{"a rate set is used", map[string]Rate{"site_download": {PerMinute: 60, Burst: 3}}, map[string]Rate{"site_download": {60, 3}}, 3},
		{"one that would refuse every request is not", map[string]Rate{"site_download": {PerMinute: 0, Burst: 3}}, map[string]Rate{}, 1},
		{"nor is the env's own, or an unknown one", map[string]Rate{"site_download": {1, 1}, "nope": {5, 5}}, map[string]Rate{}, 1},
		{"one left out is the env's again", map[string]Rate{}, map[string]Rate{}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.set, limiters.SetRates(tt.rates))
			require.Equal(t, tt.burst, kept.Burst())
			require.Equal(t, tt.burst, limiters.SiteDownloadLimiter("new").Burst())
			setting := limiters.Settings()[10]
			require.Equal(t, "site_download", setting.Name)
			require.Equal(t, len(tt.set) > 0, setting.Set)
			require.Equal(t, Rate{1, 1}, setting.Default)
			// Other kinds keep the env's
			require.Equal(t, 10*SiteIdentitiesShare, limiters.SitePOSTLimiter("site").Burst())
		})
	}
	for _, s := range limiters.Stats(time.Now()) {
		if s.Name == "site_download" {
			require.Equal(t, 1, s.Burst)
		}
	}
}

// The limiters are kept for cachedIdentitiesPerSite identities a site, whatever MAX_IDENTITIES_PER_USER is, so a high
// one does not let unique IPs or names fill memory; one pushed out is made again, starting full
func TestRateLimiterCachesKeepAHouseholdsIdentitiesASite(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{MaxSites: 1, MaxIdentitiesPerSite: 100, SitePOSTPerMinuteLimit: 1, SitePOSTPerMinuteBurst: 1})
	now := time.Now()
	require.True(t, limiters.IdentityPOSTLimiter(0).AllowN(now, 1))
	require.False(t, limiters.IdentityPOSTLimiter(0).AllowN(now, 1), "the first identity has spent its burst")

	// The cache keeps 1 site × 20 identities × 2, so the 40 after it push the first out
	for id := int64(1); id <= 2*cachedIdentitiesPerSite; id++ {
		limiters.IdentityPOSTLimiter(id)
	}
	require.True(t, limiters.IdentityPOSTLimiter(0).AllowN(now, 1), "made again, it starts full")
}

// One identity spending all its writes, as a phone uploading a day's backlog does, leaves the site's limit room for its
// other identities, so a carer's carbs or an AAPSClient command still go through
func TestOneIdentityLeavesItsSiteRoomForOthers(t *testing.T) {
	limiters := newRateLimiters(ServerConfig{MaxSites: 10, MaxIdentitiesPerSite: 10, SitePOSTPerMinuteLimit: 60, SitePOSTPerMinuteBurst: 4})
	now := time.Now()
	site, phone, carer := limiters.SitePOSTLimiter("site"), limiters.IdentityPOSTLimiter(1), limiters.IdentityPOSTLimiter(2)
	for range 4 {
		require.True(t, phone.AllowN(now, 1))
		require.True(t, site.AllowN(now, 1))
	}
	require.False(t, phone.AllowN(now, 1), "the phone has spent its own limit")
	require.True(t, carer.AllowN(now, 1), "the carer has its own")
	require.True(t, site.AllowN(now, 1), "and the site has room for the carer's write")
}
