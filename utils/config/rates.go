package config

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

// This contains all rate limiters in the system
// This might be more efficient if joined together, but right now this is fine
// TODO: find out if this needs optimisation
type RateLimiters struct {
	serverConfig ServerConfig

	deviceReadRateLimiters *lru.Cache[string, *rate.Limiter]
	userReadRateLimiters   *lru.Cache[string, *rate.Limiter]

	deviceWriteRateLimiters *lru.Cache[string, *rate.Limiter]
	userWriteRateLimiters   *lru.Cache[string, *rate.Limiter]

	UserCreateLimiter   *rate.Limiter
	userCreateIpLimiter *lru.Cache[string, *rate.Limiter]

	GlobalRateLimiter    *rate.Limiter // Limiter for all requestes everywhere
	globalIPRateLimiters *lru.Cache[string, *rate.Limiter]
}

func newRateLimiters(config ServerConfig) *RateLimiters {
	maxUsers := config.MaxUsers
	maxDevices := config.MaxDevicesPerUser
	wiggle := 2
	// We are using the hashicorp LRU to make sure that we dont run out of memory if all users with all devices are hitting us always
	return &RateLimiters{
		serverConfig: config,

		deviceReadRateLimiters: lruMust(maxUsers * maxDevices * wiggle),
		userReadRateLimiters:   lruMust(maxUsers * wiggle),

		deviceWriteRateLimiters: lruMust(maxUsers * maxDevices * wiggle),
		userWriteRateLimiters:   lruMust(maxUsers * wiggle),

		UserCreateLimiter:   rate.NewLimiter(rate.Limit(float64(config.UserCreateGlobalPerMinuteLimit)/60.0), config.UserCreateGlobalPerMinuteBurst),
		userCreateIpLimiter: lruMust(maxUsers * wiggle),

		GlobalRateLimiter:    rate.NewLimiter(rate.Limit(float64(config.GlobalIPPerMinuteLimit)/60.0), config.GlobalIPPerMinuteBurst),
		globalIPRateLimiters: lruMust(maxUsers * maxDevices * wiggle * 2), // If a DDOS attack wiped out this cache with unique addresses the global rate limiter would still catch it
	}

}

func lruMust(maxEntries int) *lru.Cache[string, *rate.Limiter] {
	cache, err := lru.New[string, *rate.Limiter](maxEntries)
	if err != nil {
		panic(err)
	}

	return cache
}

func (r *RateLimiters) getOrCreateLimiter(limiters *lru.Cache[string, *rate.Limiter], key string, rps int, burst int) *rate.Limiter {
	limiter, exists := limiters.Get(key)
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(float64(rps)/60.0), burst)
		limiters.Add(key, limiter)
	}
	return limiter
}

func (r *RateLimiters) GetGlobalIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.globalIPRateLimiters, ipAddress, r.serverConfig.GlobalIPPerMinuteLimit, r.serverConfig.GlobalIPPerMinuteBurst)
}

func (r *RateLimiters) GetUserCreateIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.userCreateIpLimiter, ipAddress, r.serverConfig.UserCreateIPPerMinuteLimit, r.serverConfig.UserCreateIPPerMinuteBurst)
}

// //
// User Rate Limiters
// //

func (r *RateLimiters) GetDeviceReadLimiter(deviceID string) *rate.Limiter {
	return r.getOrCreateLimiter(r.deviceReadRateLimiters, deviceID, r.serverConfig.UserReadPerMinuteLimit, r.serverConfig.UserReadPerMinuteBurst)
}

func (r *RateLimiters) GetDeviceWriteLimiter(deviceID string) *rate.Limiter {

	return r.getOrCreateLimiter(r.deviceWriteRateLimiters, deviceID, r.serverConfig.UserWritePerMinuteLimit, r.serverConfig.UserWritePerMinuteBurst)
}

func (r *RateLimiters) GetUserReadLimiter(userId string) *rate.Limiter {

	return r.getOrCreateLimiter(r.userReadRateLimiters, userId, r.serverConfig.UserReadPerMinuteLimit, r.serverConfig.UserReadPerMinuteBurst)
}

func (r *RateLimiters) GetUserWriteLimiter(userId string) *rate.Limiter {

	return r.getOrCreateLimiter(r.userWriteRateLimiters, userId, r.serverConfig.UserWritePerMinuteLimit, r.serverConfig.UserWritePerMinuteBurst)
}
