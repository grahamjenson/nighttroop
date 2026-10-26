package httputils

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

// AddStandardMiddleware adds what every router has, recording its server errors in logs when it is set
func AddStandardMiddleware(router chi.Router, config *config.ServerConfig, logs *db.LogsDB) {
	router.Use(middleware.RequestID)
	if logs != nil {
		// Before the recoverer, so a panic's 500 is recorded too
		router.Use(RecordServerErrors(logs))
	}

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

// RecordServerErrors records each response with a 5xx status in the logs DB's errors, with its site when it has one.
// It records the route's pattern rather than the path, as a Nightscout path can carry a token.
func RecordServerErrors(logs *db.LogsDB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			if ww.Status() < http.StatusInternalServerError {
				return
			}
			var siteID *string
			if siteContext := GetSiteContext(r.Context()); siteContext != nil && siteContext.SiteModel != nil {
				siteID = &siteContext.SiteModel.ID
			}
			route := ""
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
				route = routeContext.RoutePattern()
			}
			logs.Error(context.WithoutCancel(r.Context()), siteID, fmt.Errorf("%d %s", ww.Status(), http.StatusText(ww.Status())), map[string]any{
				"method": r.Method, "route": route, "request_id": middleware.GetReqID(r.Context()),
			})
		})
	}
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
				controlDB.UpdateSiteAndIdentityDateFlags(siteContext.SiteModel, siteContext.SiteDB, siteContext.IdentityModel, true, wasWrite)
			}
		})
	}
}
