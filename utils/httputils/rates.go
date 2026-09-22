package httputils

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
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

// //////
// Once We know Who the user is then we can apply user-specific rate limiting
// //////
func SiteHTTPActionWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.Subdomain

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

func UserUpdateSyncWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.Subdomain

	writeLimiter := serverConfig.RateLimiters.SiteUpdateSyncLimiter(userSubdomain)

	return WithinRateLimitAll(writeLimiter)
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
