package config

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

// This contains all rate limiters in the system
// This might be more efficient if joined together, but right now this is fine
// TODO: find out if this needs optimisation
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

	publicRateMutex *sync.Mutex // Having a single mutex will slow things down, but that is kind of the point here
	userRateMutex   *sync.Mutex
}

func newRateLimiters(config ServerConfig) *RateLimiters {
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

		// Keep these seperate to stop public endpoint interfering with users
		publicRateMutex: &sync.Mutex{},
		userRateMutex:   &sync.Mutex{},
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
	r.publicRateMutex.Lock()
	defer r.publicRateMutex.Unlock()

	ipLimiter := r.getOrCreateLimiter(r.userCreateIpLimiter, ipAddress, r.serverConfig.UserCreateIPPerMinuteLimit, r.serverConfig.UserCreateIPPerMinuteBurst)
	ipallow, ipdelay := Allow(ipLimiter)
	userallow, userdelay := Allow(r.userCreateLimiter)

	return ipallow && userallow, maxDuration(ipdelay, userdelay)
}

// This will return false if any rate limiter is blocking, with a min of time left until rate is restored
func (r *RateLimiters) UserReadAllowed(userId string, deviceID string, ipAddress string) (bool, int) {
	r.userRateMutex.Lock()
	defer r.userRateMutex.Unlock()

	ipLimiter := r.getOrCreateLimiter(r.ipReadRateLimiters, ipAddress, r.serverConfig.IPReadPerMinuteLimit, r.serverConfig.IPReadPerMinuteBurst)
	deviceLimiter := r.getOrCreateLimiter(r.deviceReadRateLimiters, deviceID, r.serverConfig.DeviceReadPerMinuteLimit, r.serverConfig.DeviceReadPerMinuteBurst)
	userLimiter := r.getOrCreateLimiter(r.userReadRateLimiters, userId, r.serverConfig.UserReadPerMinuteLimit, r.serverConfig.UserReadPerMinuteBurst)
	ipallow, ipdelay := Allow(ipLimiter)
	deviceallow, devicedelay := Allow(deviceLimiter)
	userallow, userdelay := Allow(userLimiter)

	return ipallow && deviceallow && userallow, maxDuration(ipdelay, devicedelay, userdelay)
}

func (r *RateLimiters) UserWriteAllowed(userId string, deviceID string, ipAddress string) (bool, int) {
	r.userRateMutex.Lock()
	defer r.userRateMutex.Unlock()

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

func maxDuration(durations ...time.Duration) int {
	max := time.Duration(0)
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return int(max.Seconds())
}
