package config

import (
	"math"
	"strconv"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

// This contains all rate limiters in the system
// This might be more efficient if joined together, but right now this is fine
// TODO: find out if this needs optimisation
type RateLimiters struct {
	serverConfig ServerConfig

	userLoginRateLimiters *lru.Cache[string, *rate.Limiter]
	ipLoginRateLimiters   *lru.Cache[string, *rate.Limiter]

	identityGETRateLimiters *lru.Cache[string, *rate.Limiter]
	siteGETRateLimiters     *lru.Cache[string, *rate.Limiter]

	identityPOSTRateLimiters *lru.Cache[string, *rate.Limiter]
	sitePOSTRateLimiters     *lru.Cache[string, *rate.Limiter]

	siteUpdateSyncRateLimiters *lru.Cache[string, *rate.Limiter]

	SiteCreateLimiter   *rate.Limiter
	siteCreateIpLimiter *lru.Cache[string, *rate.Limiter]

	GlobalRateLimiter    *rate.Limiter // Limiter for all requestes everywhere
	globalIPRateLimiters *lru.Cache[string, *rate.Limiter]
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

		identityGETRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		siteGETRateLimiters:     lruMust(maxSites * wiggle),

		identityPOSTRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		sitePOSTRateLimiters:     lruMust(maxSites * wiggle),

		siteUpdateSyncRateLimiters: lruMust(maxSites * wiggle),

		SiteCreateLimiter:   rate.NewLimiter(rate.Limit(float64(config.SiteCreateGlobalPerMinuteLimit)/60.0), config.SiteCreateGlobalPerMinuteBurst),
		siteCreateIpLimiter: lruMust(maxSites * wiggle),

		GlobalRateLimiter:    rate.NewLimiter(rate.Limit(float64(config.GlobalRateLimit)/60.0), config.GlobalRateBurst),
		globalIPRateLimiters: lruMust(maxSites * maxIdentities * wiggle * 2), // If a DDOS attack wiped out this cache with unique addresses the global rate limiter would still catch it
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
	limiter, exists := limiters.Get(key)
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rps/60.0), burst)
		limiters.Add(key, limiter)
	}
	return limiter
}

func (r *RateLimiters) GetGlobalIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.globalIPRateLimiters, ipAddress, float64(r.serverConfig.GlobalIPPerMinuteLimit), r.serverConfig.GlobalIPPerMinuteBurst)
}

func (r *RateLimiters) GetSiteCreateIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteCreateIpLimiter, ipAddress, r.serverConfig.SiteCreateIPPerMinuteLimit, r.serverConfig.SiteCreateIPPerMinuteBurst)
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

func (r *RateLimiters) SiteLoginAllow(subdomain string, ip string) (bool, int) {
	userLimiter := r.getOrCreateLimiter(r.userLoginRateLimiters, subdomain, float64(r.serverConfig.SiteLoginPerMinuteLimit), r.serverConfig.SiteLoginPerMinuteBurst)
	ipLimiter := r.getOrCreateLimiter(r.ipLoginRateLimiters, ip, float64(r.serverConfig.SiteLoginPerMinuteLimit), r.serverConfig.SiteLoginPerMinuteBurst)

	// Calculate tokens remaining
	userTokens := userLimiter.Tokens()
	ipTokens := ipLimiter.Tokens()

	// The effective remaining attempts is the minimum of user and IP limits
	remaining := int(math.Min(userTokens, ipTokens))

	return userLimiter.Allow() && ipLimiter.Allow(), remaining
}
