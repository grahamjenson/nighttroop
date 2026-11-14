package httputils

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// /////
// Public Routes So All We have is IP Address
// ////

func GlobalIPWithinRateLimit(serverConfig *config.ServerConfig, ipAddress string) bool {
	limiter := serverConfig.RateLimiters.GetGlobalIPRateLimiter(ipAddress)
	return limiter.Allow()
}

func CreateSiteWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	createLimiter := serverConfig.RateLimiters.SiteCreateLimiter
	ipLimiter := serverConfig.RateLimiters.GetSiteCreateIPRateLimiter(GetIPFromHTTPContext(r.Context()))

	return WithinRateLimitAll(createLimiter, ipLimiter)
}

func ControlLoginAttemptsWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request, username string) (bool, int) {
	return serverConfig.RateLimiters.SiteLoginAllow(fmt.Sprintf("%s:%s", consts.CONTROL_IDENTITY_SUBDOMAIN, username), GetIPFromHTTPContext(r.Context()))
}

func SiteLoginAttemptsWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request, subdomain string, username string) (bool, int) {
	return serverConfig.RateLimiters.SiteLoginAllow(fmt.Sprintf("%s:%s", subdomain, username), GetIPFromHTTPContext(r.Context()))
}

func SiteUnauthWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	return WithinRateLimitAll(serverConfig.RateLimiters.SiteUnauthLimiter(siteContext.SiteID, siteContext.IPAddress))
}

// SiteUnauthSpent is whether an IP has no unauthenticated budget left on the site, without spending any of it
func SiteUnauthSpent(serverConfig *config.ServerConfig, r *http.Request) bool {
	siteContext := GetSiteContext(r.Context())
	return serverConfig.RateLimiters.SiteUnauthLimiter(siteContext.SiteID, siteContext.IPAddress).Tokens() < 1
}

// //////
// Once We know Who the user is then we can apply user-specific rate limiting
// //////
func SiteHTTPActionWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.SiteID

	var limiter *rate.Limiter
	switch siteContext.HTTPAction {
	case HTTPActionRead:
		limiter = serverConfig.RateLimiters.SiteGETLimiter(userSubdomain)
	case HTTPActionWrite:
		limiter = serverConfig.RateLimiters.SitePOSTLimiter(userSubdomain)
	default:
		return false, 0
	}

	return WithinRateLimitAll(limiter)
}

func IdentityHTTPWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	identity := siteContext.IdentityModel

	if identity == nil {
		return false, 0
	}

	var limiter *rate.Limiter
	switch siteContext.HTTPAction {
	case HTTPActionRead:
		limiter = serverConfig.RateLimiters.IdentityGETLimiter(identity.ID)
	case HTTPActionWrite:
		limiter = serverConfig.RateLimiters.IdentityPOSTLimiter(identity.ID)
	default:
		return false, 0
	}

	return WithinRateLimitAll(limiter)
}

// ExtraWritesWithinRateLimit spends n more writes from the site's and its identity's limits, as for a batch of many
// records, spending none unless both have them all now
func ExtraWritesWithinRateLimit(r *http.Request, n int) bool {
	siteContext := GetSiteContext(r.Context())
	if n <= 0 {
		return true
	}
	if siteContext.IdentityModel == nil {
		return false
	}
	now := time.Now()
	site := siteContext.Config.RateLimiters.SitePOSTLimiter(siteContext.SiteID).ReserveN(now, n)
	identity := siteContext.Config.RateLimiters.IdentityPOSTLimiter(siteContext.IdentityModel.ID).ReserveN(now, n)
	if site.OK() && identity.OK() && site.DelayFrom(now) == 0 && identity.DelayFrom(now) == 0 {
		return true
	}
	site.CancelAt(now)
	identity.CancelAt(now)
	return false
}

// SiteOverStorageMessage says why a site at abuse level OverUse or above, as its DB is far over its size limit, is
// refused a write
const SiteOverStorageMessage = "Too many writes: this site's data is far over its storage limit"

// abusiveWriteCost is how many flagged writes each write of a site at abuse level Abusive or above costs
const abusiveWriteCost = 4

// SiteAbuseLevel is the request's site's abuse level: the higher of the level last stored and the level of its DB's
// size now, as the score is stored only every 5 minutes, while a site can write many MB a minute.
func SiteAbuseLevel(siteContext *SiteContext) controldb.AbuseLevel {
	if siteContext == nil || siteContext.SiteModel == nil {
		return controldb.AbuseGood
	}
	site := siteContext.SiteModel
	live := controldb.AbuseSignals{SizeKB: site.EstimatedDbSizeKb, LimitKB: site.MaxDbSizeKbLimit}
	if siteContext.SiteDB != nil {
		if size := siteContext.SiteDB.EstimatedSizeKB(); size != nil {
			live.SizeKB = size
		}
	}
	return max(site.AbuseLevel(), live.Level())
}

// FlaggedSiteWritesWithinRateLimit spends n writes from the flagged write limit of a site at abuse level OverUse, and
// 4 times n at Abusive or above, none unless it has them all now, and none of a site under OverUse. It is when to try
// again if refused, in seconds.
func FlaggedSiteWritesWithinRateLimit(siteContext *SiteContext, n int) (bool, int) {
	level := SiteAbuseLevel(siteContext)
	if n <= 0 || level < controldb.AbuseOverUse {
		return true, 0
	}
	if level >= controldb.AbuseAbusive {
		n *= abusiveWriteCost
	}
	now := time.Now()
	writes := siteContext.Config.RateLimiters.SiteFlaggedPOSTLimiter(siteContext.SiteID).ReserveN(now, n)
	if writes.OK() && writes.DelayFrom(now) == 0 {
		return true, 0
	}
	// More than its burst is never let through, so there is no time to wait for
	delay := time.Duration(0)
	if writes.OK() {
		delay = writes.DelayFrom(now)
	}
	writes.CancelAt(now)
	return false, maxDuration(delay)
}

func UserUpdateConnectorsWithinRateLimit(siteContext *SiteContext) (bool, int) {

	userSubdomain := siteContext.SiteID

	writeLimiter := siteContext.Config.RateLimiters.SiteUpdateConnectorsLimiter(userSubdomain)

	return WithinRateLimitAll(writeLimiter)
}

func IdentityCreateWithinRateLimit(siteContext *SiteContext) (bool, int) {
	return WithinRateLimitAll(siteContext.Config.RateLimiters.SiteIdentityCreateLimiter(siteContext.SiteID))
}

// //
// Utils Methods
// //
func WithinRateLimitAll(limiters ...*rate.Limiter) (bool, int) {
	alloweds := make([]bool, len(limiters))
	remainingTimes := make([]time.Duration, len(limiters))

	for i, limiter := range limiters {
		alloweds[i], remainingTimes[i] = allowWithTimeRemaining(limiter)
	}

	return allTrue(alloweds), maxDuration(remainingTimes...)
}

func allowWithTimeRemaining(limiter *rate.Limiter) (bool, time.Duration) {
	// Use a reservation to return the amount of time until the next event is allowed.
	res := limiter.Reserve()
	defer res.Cancel() // Cancel the reservation to avoid consuming tokens

	// If the reservation is not possible (e.g., burst size exceeded)
	if !res.OK() {
		return false, 0
	}

	delay := res.Delay()
	if delay == 0 {
		return true, 0
	}

	return false, delay
}

func allTrue(bools []bool) bool {
	for _, b := range bools {
		if !b {
			return false
		}
	}
	return true
}

func maxDuration(durations ...time.Duration) int {
	max := time.Duration(0)
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return int(max.Seconds())
}
