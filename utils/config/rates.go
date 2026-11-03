package config

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/netip"
	"strconv"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

// This contains all rate limiters in the system
// This might be more efficient if joined together, but right now this is fine
type RateLimiters struct {
	serverConfig ServerConfig

	userLoginRateLimiters *lru.Cache[string, *rate.Limiter]
	ipLoginRateLimiters   *lru.Cache[string, *rate.Limiter]

	siteUnauthRateLimiters *lru.Cache[string, *rate.Limiter]

	identityGETRateLimiters *lru.Cache[string, *rate.Limiter]
	siteGETRateLimiters     *lru.Cache[string, *rate.Limiter]

	identityPOSTRateLimiters *lru.Cache[string, *rate.Limiter]
	sitePOSTRateLimiters     *lru.Cache[string, *rate.Limiter]

	siteUpdateSyncRateLimiters     *lru.Cache[string, *rate.Limiter]
	siteIdentityCreateRateLimiters *lru.Cache[string, *rate.Limiter]
	siteDownloadRateLimiters       *lru.Cache[string, *rate.Limiter]
	sitePushRateLimiters           *lru.Cache[string, *rate.Limiter]

	SiteCreateLimiter   *rate.Limiter
	siteCreateIpLimiter *lru.Cache[string, *rate.Limiter]

	GlobalRateLimiter    *rate.Limiter // Limiter for all requestes everywhere
	globalIPRateLimiters *lru.Cache[string, *rate.Limiter]
	// ControlRateLimiter is the control site's own limit at the global rate, for those not signed in to it, such as
	// its login, so a site using up the global limit cannot stop a control admin signing in to fix it
	ControlRateLimiter *rate.Limiter
}

func newRateLimiters(config ServerConfig) *RateLimiters {
	maxSites := config.MaxSites
	maxIdentities := config.MaxIdentitiesPerSite
	wiggle := 2
	// We are using the hashicorp LRU to make sure that we dont run out of memory if all users with all identitys are hitting us always
	return &RateLimiters{
		serverConfig: config,

		userLoginRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		ipLoginRateLimiters:   lruMust(maxSites * maxIdentities * wiggle * 2),

		siteUnauthRateLimiters: lruMust(maxSites * maxIdentities * wiggle * 2),

		identityGETRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		siteGETRateLimiters:     lruMust(maxSites * wiggle),

		identityPOSTRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		sitePOSTRateLimiters:     lruMust(maxSites * wiggle),

		siteUpdateSyncRateLimiters:     lruMust(maxSites * wiggle),
		siteIdentityCreateRateLimiters: lruMust(maxSites * wiggle),
		siteDownloadRateLimiters:       lruMust(maxSites * wiggle),
		sitePushRateLimiters:           lruMust(maxSites * wiggle),

		SiteCreateLimiter:   rate.NewLimiter(rate.Limit(float64(config.SiteCreateGlobalPerMinuteLimit)/60.0), config.SiteCreateGlobalPerMinuteBurst),
		siteCreateIpLimiter: lruMust(maxSites * wiggle),

		GlobalRateLimiter:    rate.NewLimiter(rate.Limit(float64(config.GlobalRateLimit)/60.0), config.GlobalRateBurst),
		globalIPRateLimiters: lruMust(maxSites * maxIdentities * wiggle * 2), // If a DDOS attack wiped out this cache with unique addresses the global rate limiter would still catch it
		ControlRateLimiter:   rate.NewLimiter(rate.Limit(float64(config.GlobalRateLimit)/60.0), config.GlobalRateBurst),
	}

}

func lruMust(maxEntries int) *lru.Cache[string, *rate.Limiter] {
	cache, err := lru.New[string, *rate.Limiter](maxEntries)
	if err != nil {
		panic(err)
	}

	return cache
}

func (r *RateLimiters) getOrCreateLimiter(limiters *lru.Cache[string, *rate.Limiter], key string, rps float64, burst int) *rate.Limiter {
	// Keys come from user input (e.g. a login name), so hash them to keep every entry a fixed size
	sum := sha256.Sum256([]byte(key))
	key = hex.EncodeToString(sum[:])
	limiter, exists := limiters.Get(key)
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rps/60.0), burst)
		limiters.Add(key, limiter)
	}
	return limiter
}

// ipKey is what an IP's limiters are keyed on. An IPv6 address is keyed on its /64, as a host is usually given a whole
// /64 to pick addresses from, so stepping through it cannot reset its limits.
func ipKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	if addr = addr.Unmap(); addr.Is4() {
		return addr.String()
	}
	prefix, _ := addr.Prefix(64)
	return prefix.String()
}

func (r *RateLimiters) GetGlobalIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.globalIPRateLimiters, ipKey(ipAddress), float64(r.serverConfig.GlobalIPPerMinuteLimit), r.serverConfig.GlobalIPPerMinuteBurst)
}

func (r *RateLimiters) GetSiteCreateIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteCreateIpLimiter, ipKey(ipAddress), r.serverConfig.SiteCreateIPPerMinuteLimit, r.serverConfig.SiteCreateIPPerMinuteBurst)
}

// //
// Site Rate Limiters
// //

func (r *RateLimiters) IdentityGETLimiter(identityID int64) *rate.Limiter {
	return r.getOrCreateLimiter(r.identityGETRateLimiters, strconv.FormatInt(identityID, 10), float64(r.serverConfig.SiteGETPerMinuteLimit), r.serverConfig.SiteGETPerMinuteBurst)
}

func (r *RateLimiters) IdentityPOSTLimiter(identityID int64) *rate.Limiter {

	return r.getOrCreateLimiter(r.identityPOSTRateLimiters, strconv.FormatInt(identityID, 10), float64(r.serverConfig.SitePOSTPerMinuteLimit), r.serverConfig.SitePOSTPerMinuteBurst)
}

func (r *RateLimiters) SiteGETLimiter(subdomain string) *rate.Limiter {

	return r.getOrCreateLimiter(r.siteGETRateLimiters, subdomain, float64(r.serverConfig.SiteGETPerMinuteLimit), r.serverConfig.SiteGETPerMinuteBurst)
}

func (r *RateLimiters) SitePOSTLimiter(subdomain string) *rate.Limiter {

	return r.getOrCreateLimiter(r.sitePOSTRateLimiters, subdomain, float64(r.serverConfig.SitePOSTPerMinuteLimit), r.serverConfig.SitePOSTPerMinuteBurst)
}

func (r *RateLimiters) SiteUpdateSyncLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteUpdateSyncRateLimiters, subdomain, float64(r.serverConfig.SiteUpdateSyncPerMinuteLimit), r.serverConfig.SiteUpdateSyncPerMinuteBurst)
}

func (r *RateLimiters) SiteIdentityCreateLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteIdentityCreateRateLimiters, subdomain, float64(r.serverConfig.SiteIdentityCreatePerMinuteLimit), r.serverConfig.SiteIdentityCreatePerMinuteBurst)
}

func (r *RateLimiters) SiteDownloadLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteDownloadRateLimiters, subdomain, float64(r.serverConfig.SiteDownloadPerMinuteLimit), r.serverConfig.SiteDownloadPerMinuteBurst)
}

// SitePushLimiter is a site's budget of pushes, each a request to Apple or Firebase, so a site cannot use the server to
// send them without end
func (r *RateLimiters) SitePushLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.sitePushRateLimiters, subdomain, float64(r.serverConfig.SitePushPerMinuteLimit), r.serverConfig.SitePushPerMinuteBurst)
}

// SiteUnauthLimiter is one IP's budget on a site before it authenticates: 401s and Socket.IO handshakes spend it, not the site's own limit
func (r *RateLimiters) SiteUnauthLimiter(subdomain string, ip string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteUnauthRateLimiters, subdomain+":"+ipKey(ip), float64(r.serverConfig.SiteUnauthPerMinuteLimit), r.serverConfig.SiteUnauthPerMinuteBurst)
}

func (r *RateLimiters) SiteLoginAllow(subdomain string, ip string) (bool, int) {
	userLimiter := r.getOrCreateLimiter(r.userLoginRateLimiters, subdomain, float64(r.serverConfig.SiteLoginPerMinuteLimit), r.serverConfig.SiteLoginPerMinuteBurst)
	ipLimiter := r.getOrCreateLimiter(r.ipLoginRateLimiters, ipKey(ip), float64(r.serverConfig.SiteLoginPerMinuteLimit), r.serverConfig.SiteLoginPerMinuteBurst)

	// Calculate tokens remaining
	userTokens := userLimiter.Tokens()
	ipTokens := ipLimiter.Tokens()

	// The effective remaining attempts is the minimum of user and IP limits
	remaining := int(math.Min(userTokens, ipTokens))

	return userLimiter.Allow() && ipLimiter.Allow(), remaining
}

// RateLimiterStats is how close one kind of rate limiter is to refusing. One is kept for each Per: the server, an IP, a
// site, an identity, a login name or an IP on a site; each refills at PerMinute up to Burst.
type RateLimiterStats struct {
	Name      string  `json:"name"`
	Per       string  `json:"per"`
	PerMinute float64 `json:"per_minute"`
	Burst     int     `json:"burst"`
	// Tracked is how many are kept, Limited how many would refuse a request now, and FullestPercent how much of its
	// burst the most used one has spent
	Tracked        int     `json:"tracked"`
	Limited        int     `json:"limited"`
	FullestPercent float64 `json:"fullest_percent"`
}

// Stats are how close each kind of rate limiter is to refusing at now, read without spending any
func (r *RateLimiters) Stats(now time.Time) []RateLimiterStats {
	c := r.serverConfig
	login := float64(c.SiteLoginPerMinuteLimit)
	kinds := []struct {
		name, per string
		perMinute float64
		burst     int
		limiters  []*rate.Limiter
	}{
		{"global", "server", float64(c.GlobalRateLimit), c.GlobalRateBurst, []*rate.Limiter{r.GlobalRateLimiter}},
		{"control", "server", float64(c.GlobalRateLimit), c.GlobalRateBurst, []*rate.Limiter{r.ControlRateLimiter}},
		{"global_ip", "ip", float64(c.GlobalIPPerMinuteLimit), c.GlobalIPPerMinuteBurst, r.globalIPRateLimiters.Values()},
		{"site_create", "server", float64(c.SiteCreateGlobalPerMinuteLimit), c.SiteCreateGlobalPerMinuteBurst, []*rate.Limiter{r.SiteCreateLimiter}},
		{"site_create_ip", "ip", c.SiteCreateIPPerMinuteLimit, c.SiteCreateIPPerMinuteBurst, r.siteCreateIpLimiter.Values()},
		{"login_name", "login", login, c.SiteLoginPerMinuteBurst, r.userLoginRateLimiters.Values()},
		{"login_ip", "ip", login, c.SiteLoginPerMinuteBurst, r.ipLoginRateLimiters.Values()},
		{"site_unauth", "site_ip", float64(c.SiteUnauthPerMinuteLimit), c.SiteUnauthPerMinuteBurst, r.siteUnauthRateLimiters.Values()},
		{"site_get", "site", float64(c.SiteGETPerMinuteLimit), c.SiteGETPerMinuteBurst, r.siteGETRateLimiters.Values()},
		{"site_post", "site", float64(c.SitePOSTPerMinuteLimit), c.SitePOSTPerMinuteBurst, r.sitePOSTRateLimiters.Values()},
		{"identity_get", "identity", float64(c.SiteGETPerMinuteLimit), c.SiteGETPerMinuteBurst, r.identityGETRateLimiters.Values()},
		{"identity_post", "identity", float64(c.SitePOSTPerMinuteLimit), c.SitePOSTPerMinuteBurst, r.identityPOSTRateLimiters.Values()},
		{"site_update_sync", "site", float64(c.SiteUpdateSyncPerMinuteLimit), c.SiteUpdateSyncPerMinuteBurst, r.siteUpdateSyncRateLimiters.Values()},
		{"site_identity_create", "site", float64(c.SiteIdentityCreatePerMinuteLimit), c.SiteIdentityCreatePerMinuteBurst, r.siteIdentityCreateRateLimiters.Values()},
		{"site_download", "site", float64(c.SiteDownloadPerMinuteLimit), c.SiteDownloadPerMinuteBurst, r.siteDownloadRateLimiters.Values()},
		{"site_push", "site", float64(c.SitePushPerMinuteLimit), c.SitePushPerMinuteBurst, r.sitePushRateLimiters.Values()},
	}
	stats := make([]RateLimiterStats, 0, len(kinds))
	for _, kind := range kinds {
		s := RateLimiterStats{Name: kind.name, Per: kind.per, PerMinute: kind.perMinute, Burst: kind.burst, Tracked: len(kind.limiters)}
		for _, limiter := range kind.limiters {
			tokens := limiter.TokensAt(now)
			if tokens < 1 {
				s.Limited++
			}
			if kind.burst > 0 {
				s.FullestPercent = max(s.FullestPercent, min(100, 100*(1-tokens/float64(kind.burst))))
			}
		}
		stats = append(stats, s)
	}
	return stats
}
