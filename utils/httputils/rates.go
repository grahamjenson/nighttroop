package httputils

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/metrics"
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

// SiteAndIdentityWithinRateLimit spends one read or write from the site's and its identity's limits, and none unless
// both have one now, so an identity over its own limit cannot spend the site's that its other apps share
func SiteAndIdentityWithinRateLimit(serverConfig *config.ServerConfig, r *http.Request) (bool, int) {
	siteContext := GetSiteContext(r.Context())
	identity := siteContext.IdentityModel
	if identity == nil {
		return false, 0
	}

	limiters := serverConfig.RateLimiters
	switch siteContext.HTTPAction {
	case HTTPActionRead:
		return WithinRateLimitAll(limiters.SiteGETLimiter(siteContext.SiteID), limiters.IdentityGETLimiter(identity.ID))
	case HTTPActionWrite:
		return WithinRateLimitAll(limiters.SitePOSTLimiter(siteContext.SiteID), limiters.IdentityPOSTLimiter(identity.ID))
	default:
		return false, 0
	}
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

// SiteOverStorageMessage says why a site flagged for abuse, as its DB is far over its size limit, is refused a write
const SiteOverStorageMessage = "Too many writes: this site's data is far over its storage limit"

// SiteStorageFullMessage says why a write to a site whose DB is over WritesRefusedSizeFactor times its size limit is
// refused, and how to write again
const SiteStorageFullMessage = "This site's data is far over its storage limit, so nothing more is stored until some is deleted or the site is truncated"

// ServerDiskFullMessage says why a write is refused while the server's disk is nearly full
const ServerDiskFullMessage = "The server's disk is nearly full, so nothing more is stored until its admin frees space"

// siteStorageFullRetrySeconds is when a client refused as its site is over its storage, or the server's disk is
// nearly full, is told to try again: the hour the server takes to truncate a site over its limit
const siteStorageFullRetrySeconds = 3600

// SiteFlaggedForAbuse is whether the request's site is flagged for abuse. It uses the DB's size now, not the size last
// recorded, which can be minutes old, while a site can write many MB a minute.
func SiteFlaggedForAbuse(siteContext *SiteContext) bool {
	site := siteNow(siteContext)
	return site != nil && site.FlaggedForAbuse()
}

// SiteWriteRefused is why a write that adds to the request's site's DB is refused, or "" when it is not: the server's
// disk is nearly full, or the DB is now over WritesRefusedSizeFactor times its size limit
func SiteWriteRefused(siteContext *SiteContext) string {
	if siteContext != nil && siteContext.ControlDB != nil && siteContext.ControlDB.DiskFull() {
		return ServerDiskFullMessage
	}
	if site := siteNow(siteContext); site != nil && site.WritesRefused() {
		return SiteStorageFullMessage
	}
	return ""
}

// siteNow is the request's site with its DB's size now, or nil
func siteNow(siteContext *SiteContext) *controldb.Site {
	if siteContext == nil || siteContext.SiteModel == nil {
		return nil
	}
	site := *siteContext.SiteModel
	if siteContext.SiteDB != nil {
		if size := siteContext.SiteDB.EstimatedSizeKB(); size != nil {
			site.EstimatedDbSizeKb = size
		}
	}
	return &site
}

// RefuseWritesOverStorage refuses a write that adds to the site's DB, with when to try again, while the DB is over
// WritesRefusedSizeFactor times its size limit or the server's disk is nearly full. A delete, which makes it smaller,
// is let through.
func RefuseWritesOverStorage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete || GetHTTPAction(r) != HTTPActionWrite {
			next.ServeHTTP(w, r)
			return
		}
		if refused := SiteWriteRefused(GetSiteContext(r.Context())); refused != "" {
			writeTooManyRequests(w, siteStorageFullRetrySeconds)
			_, _ = w.Write([]byte(refused))
			slog.InfoContext(r.Context(), "A write that adds to a site's DB is refused", "why", refused)
			metrics.WriteRefused(map[string]string{ServerDiskFullMessage: "disk_full", SiteStorageFullMessage: "site_storage"}[refused])
			return
		}
		next.ServeHTTP(w, r)
	})
}

// FlaggedSiteWritesWithinRateLimit spends n writes from the flagged write limit of a site flagged for abuse, none unless
// it has them all now, and none of a site that is not flagged. It is when to try again if refused, in seconds.
func FlaggedSiteWritesWithinRateLimit(siteContext *SiteContext, n int) (bool, int) {
	if n <= 0 || !SiteFlaggedForAbuse(siteContext) {
		return true, 0
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
	metrics.RateLimited()
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

// WithinRateLimitAll spends one from each limiter, and none unless all have one now, so a request one limiter refuses
// does not spend the others': an IP over its own site create limit cannot empty the server's. It is when to try again
// if refused, in seconds.
func WithinRateLimitAll(limiters ...*rate.Limiter) (bool, int) {
	now := time.Now()
	reservations := make([]*rate.Reservation, len(limiters))
	allowed, delay := true, time.Duration(0)
	for i, limiter := range limiters {
		reservations[i] = limiter.ReserveN(now, 1)
		// One past its burst is never let through, so there is no time to wait for
		if !reservations[i].OK() {
			allowed = false
		} else if wait := reservations[i].DelayFrom(now); wait > 0 {
			allowed, delay = false, max(delay, wait)
		}
	}
	if !allowed {
		metrics.RateLimited()
		// Cancelled at the time reserved, each gives back its token, as it is not yet used
		for _, reservation := range reservations {
			reservation.CancelAt(now)
		}
	}
	return allowed, maxDuration(delay)
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
