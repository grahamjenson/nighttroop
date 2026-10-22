package httputils

import (
	"log/slog"
	"net/http"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

func AddStandardMiddleware(router chi.Router, config *config.ServerConfig) {
	router.Use(middleware.RequestID)

	if config.SentryDSN != "" {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		})
		router.Use(sentryMiddleware.Handle)
	}

	router.Use(NewLogger(nil))
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestSize(500 * 1024)) // 500KB max body size, should be enough for any request and prevents abuse
	router.Use(middleware.Timeout(30 * time.Second))
}

//////
// Site based middleware
//////

func SiteRateLimitMiddleware(config *config.ServerConfig) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, remainingTime := SiteHTTPActionWithinRateLimit(config, r)
			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				slog.Info("Site rate limit exceeded", "remainingTime", remainingTime)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SiteUnauthRateLimitMiddleware runs before auth, so it limits an IP on a site rather than the site itself
func SiteUnauthRateLimitMiddleware(config *config.ServerConfig) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, remainingTime := SiteUnauthWithinRateLimit(config, r)
			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				slog.Info("Site unauthenticated rate limit exceeded", "remainingTime", remainingTime)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SiteIdentityRateLimitMiddleware(config *config.ServerConfig) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, remainingTime := IdentityHTTPWithinRateLimit(config, r)
			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				slog.Info("Identity rate limit exceeded", "remainingTime", remainingTime)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SiteIdentitySeenFlagsUpdateMiddleware(controlDB *db.ControlDB) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			siteContext := GetSiteContext(r.Context())

			sw := &StatusResponseWriter{
				ResponseWriter: w,
				Status:         http.StatusOK, // default if WriteHeader is never called
			}

			next.ServeHTTP(sw, r)

			if sw.Status == 200 || sw.Status == 201 {
				wasWrite := siteContext.HTTPAction == HTTPActionWrite
				controlDB.UpdateSiteAndIdentityDateFlags(siteContext.SiteModel, siteContext.IdentityModel, true, wasWrite)
			}
		})
	}
}
