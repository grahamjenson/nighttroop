package config

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

type RateLimiters struct {
	serverConfig ServerConfig

	ipReadRateLimiters     *lru.Cache[string, *rate.Limiter]
	deviceReadRateLimiters *lru.Cache[string, *rate.Limiter]
	userReadRateLimiters   *lru.Cache[string, *rate.Limiter]

	ipWriteRateLimiters     *lru.Cache[string, *rate.Limiter]
	deviceWriteRateLimiters *lru.Cache[string, *rate.Limiter]
	userWriteRateLimiters   *lru.Cache[string, *rate.Limiter]

	userCreateLimiter   *rate.Limiter
	userCreateIpLimiter *lru.Cache[string, *rate.Limiter]

	rateMutex *sync.Mutex // Having a single mutex will slow things down, but that is kind of the point here
}

func NewRateLimiters(config ServerConfig) *RateLimiters {
	maxUsers := config.MaxUsers
	maxDevices := config.MaxDevicesPerUser
	wiggle := 2
	// We are using the hashicorp LRU to make sure that we dont run out of memory if all users with all devices are hitting us always
	return &RateLimiters{
		serverConfig: config,

		ipReadRateLimiters:     lruMust(maxUsers * maxDevices * wiggle),
		deviceReadRateLimiters: lruMust(maxUsers * maxDevices * wiggle),
		userReadRateLimiters:   lruMust(maxUsers * wiggle),

		ipWriteRateLimiters:     lruMust(maxUsers * maxDevices * wiggle),
		deviceWriteRateLimiters: lruMust(maxUsers * maxDevices * wiggle),
		userWriteRateLimiters:   lruMust(maxUsers * wiggle),

		userCreateLimiter:   rate.NewLimiter(rate.Limit(float64(config.UserCreateGlobalPerMinuteLimit)/60.0), config.UserCreateGlobalPerMinuteBurst),
		userCreateIpLimiter: lruMust(maxUsers * wiggle),
		rateMutex:           &sync.Mutex{},
	}

}

func lruMust(maxEntries int) *lru.Cache[string, *rate.Limiter] {
	cache, err := lru.New[string, *rate.Limiter](maxEntries)
	if err != nil {
		panic(err)
	}
	return cache
}

// Global function to check user create rate limit
func (r *RateLimiters) CreateUserAllowed(ipAddress string) (bool, int) {
	r.rateMutex.Lock()
	defer r.rateMutex.Unlock()

	ipLimiter := r.getOrCreateLimiter(r.userCreateIpLimiter, ipAddress, r.serverConfig.UserCreateIPPerMinuteLimit, r.serverConfig.UserCreateIPPerMinuteBurst)
	ipallow, ipdelay := Allow(ipLimiter)
	userallow, userdelay := Allow(r.userCreateLimiter)

	return ipallow && userallow, maxDuration(ipdelay, userdelay)
}

// This will return false if any rate limiter is blocking, with a min of time left until rate is restored
func (r *RateLimiters) UserReadAllowed(userId string, deviceID string, ipAddress string) (bool, int) {
	r.rateMutex.Lock()
	defer r.rateMutex.Unlock()

	ipLimiter := r.getOrCreateLimiter(r.ipReadRateLimiters, ipAddress, r.serverConfig.IPReadPerMinuteLimit, r.serverConfig.IPReadPerMinuteBurst)
	deviceLimiter := r.getOrCreateLimiter(r.deviceReadRateLimiters, deviceID, r.serverConfig.DeviceReadPerMinuteLimit, r.serverConfig.DeviceReadPerMinuteBurst)
	userLimiter := r.getOrCreateLimiter(r.userReadRateLimiters, userId, r.serverConfig.UserReadPerMinuteLimit, r.serverConfig.UserReadPerMinuteBurst)
	ipallow, ipdelay := Allow(ipLimiter)
	deviceallow, devicedelay := Allow(deviceLimiter)
	userallow, userdelay := Allow(userLimiter)

	return ipallow && deviceallow && userallow, maxDuration(ipdelay, devicedelay, userdelay)
}

func (r *RateLimiters) UserWriteAllowed(userId string, deviceID string, ipAddress string) (bool, int) {
	r.rateMutex.Lock()
	defer r.rateMutex.Unlock()

	ipLimiter := r.getOrCreateLimiter(r.ipWriteRateLimiters, ipAddress, r.serverConfig.IPWritePerMinuteLimit, r.serverConfig.IPWritePerMinuteBurst)
	deviceLimiter := r.getOrCreateLimiter(r.deviceWriteRateLimiters, deviceID, r.serverConfig.DeviceWritePerMinuteLimit, r.serverConfig.DeviceWritePerMinuteBurst)
	userLimiter := r.getOrCreateLimiter(r.userWriteRateLimiters, userId, r.serverConfig.UserWritePerMinuteLimit, r.serverConfig.UserWritePerMinuteBurst)
	ipallow, ipdelay := Allow(ipLimiter)
	deviceallow, devicedelay := Allow(deviceLimiter)
	userallow, userdelay := Allow(userLimiter)

	return ipallow && deviceallow && userallow, maxDuration(ipdelay, devicedelay, userdelay)
}

func (r *RateLimiters) getOrCreateLimiter(limiters *lru.Cache[string, *rate.Limiter], key string, rps int, burst int) *rate.Limiter {
	limiter, exists := limiters.Get(key)
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(float64(rps)/60.0), burst)
		limiters.Add(key, limiter)
	}
	return limiter
}

func Allow(limiter *rate.Limiter) (bool, time.Duration) {
	if limiter.Allow() {
		return true, 0
	}
	return false, limiter.Reserve().Delay()
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
