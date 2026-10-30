package httputils

import (
	"context"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/usage"
)

// AddStandardMiddleware adds what every router has, recording its server errors in the logs DB when there is one
func AddStandardMiddleware(router chi.Router, config *config.ServerConfig, controlDB *db.ControlDB) {
	router.Use(middleware.RequestID)
	if controlDB != nil {
		// Before the recoverer, so a panic's 500 is recorded too
		router.Use(RecordServerErrors(controlDB.Logs))
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
	// Gzip text responses, like the pages, styles, scripts and JSON, for clients that accept it
	router.Use(middleware.Compress(5))
	router.Use(withSupportEmail(config.SupportEmail))
}

type supportEmailKey struct{}

// withSupportEmail puts the server's support email address in each request's context, for its pages' footer
func withSupportEmail(email string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), supportEmailKey{}, email)))
		})
	}
}

// SupportEmail is the server's support email address, or the default outside a request
func SupportEmail(ctx context.Context) string {
	if email, ok := ctx.Value(supportEmailKey{}).(string); ok && email != "" {
		return email
	}
	return config.DefaultSupportEmail
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
			logs.ServerError(context.WithoutCancel(r.Context()), siteID, ww.Status(), r.Method, route, middleware.GetReqID(r.Context()))
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
				if siteContext.Config != nil && siteContext.IdentityModel != nil {
					siteContext.Config.Usage.Seen(siteContext.SiteID, siteContext.IdentityModel.ID, r.ContentLength, time.Now())
				}
				wasWrite := siteContext.HTTPAction == HTTPActionWrite
				controlDB.UpdateSiteAndIdentityDateFlags(siteContext.SiteModel, siteContext.SiteDB, siteContext.IdentityModel, true, wasWrite)
			}
		})
	}
}

// CountRoutes counts each request to the router named name by its route's pattern, never its path, which can carry a
// token or a site. Static files, whose pattern is a file name, are counted together, and so are requests no route
// matched, whatever their method.
func CountRoutes(tracker *usage.Tracker, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			next.ServeHTTP(w, r)
			route := name + " unmatched"
			if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
				route = name + " " + r.Method + " " + rctx.RoutePattern()
				if strings.Contains(path.Base(rctx.RoutePattern()), ".") {
					route = name + " static files"
				}
			}
			tracker.Route(route, time.Since(started))
		})
	}
}
