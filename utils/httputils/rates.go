package httputils

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

func GlobalIPRateLimitAllowed(serverConfig config.ServerConfig, ipAddress string) bool {
	limiter := serverConfig.RateLimiters.GetGlobalIPRateLimiter(ipAddress)
	return limiter.Allow()
}

func CreateUserAllowed(serverConfig config.ServerConfig, r *http.Request) (bool, int) {
	createLimiter := serverConfig.RateLimiters.UserCreateLimiter
	ipLimiter := serverConfig.RateLimiters.GetUserCreateIPRateLimiter(GetIP(r))

	return AllowAll(createLimiter, ipLimiter)
}

func UserGETAllowed(serverConfig config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.Subdomain
	// If no identity then associate with default to ensure blocking
	identityName := consts.DEFAULT_RATE_LIMITING_IDENTITY_NAME
	if siteContext.IdentityModel != nil {
		identityName = siteContext.IdentityModel.Name
	}

	readLimiter := serverConfig.RateLimiters.UserGETLimiter(userSubdomain)
	identityLimiter := serverConfig.RateLimiters.IdentityGETLimiter(identityName)

	return AllowAll(readLimiter, identityLimiter)
}

func UserPOSTAllowed(serverConfig config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.Subdomain
	// If no identity then associate with default to ensure blocking
	identityName := consts.DEFAULT_RATE_LIMITING_IDENTITY_NAME
	if siteContext.IdentityModel != nil {
		identityName = siteContext.IdentityModel.Name
	}

	writeLimiter := serverConfig.RateLimiters.UserPOSTLimiter(userSubdomain)
	identityLimiter := serverConfig.RateLimiters.IdentityPOSTLimiter(identityName)

	return AllowAll(writeLimiter, identityLimiter)
}

func UserUpdateSyncAllowed(serverConfig config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	userSubdomain := siteContext.Subdomain

	writeLimiter := serverConfig.RateLimiters.UserUpdateSyncLimiter(userSubdomain)

	return AllowAll(writeLimiter)
}

// //
// Utils Methods
// //
func AllowAll(limiters ...*rate.Limiter) (bool, int) {
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
