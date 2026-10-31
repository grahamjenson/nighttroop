package httputils

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/usage"
)

// AddStandardMiddleware adds what every router has, recording its server errors in the logs DB when there is one
func AddStandardMiddleware(router chi.Router, config *config.ServerConfig, controlDB *db.ControlDB) {
	router.Use(middleware.RequestID)
	router.Use(withRequestLogAttrs)
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
	router.Use(withDBIPCredit)
}

// withRequestLogAttrs logs the request's ID, and its site's, with everything logged with its context; the identity is
// added once it is signed in
func withRequestLogAttrs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attrs := []slog.Attr{slog.String("request_id", middleware.GetReqID(r.Context()))}
		if siteContext := GetSiteContext(r.Context()); siteContext != nil {
			attrs = append(attrs, slog.String("site_id", siteContext.SiteID))
		}
		next.ServeHTTP(w, r.WithContext(aslog.NewContext(r.Context(), attrs...)))
	})
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

type dbipCreditKey struct{}

// dbipCredit is whether a page shows locations from DB-IP's database, whose licence asks for a link to DB-IP on it
type dbipCredit struct {
	shown bool
	// partial is a part of a page loaded into one, whose footer is already drawn
	partial bool
}

// withDBIPCredit lets a request's page say it shows DB-IP's locations, for its footer to credit DB-IP
func withDBIPCredit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		credit := &dbipCredit{partial: GetAcceptType(r) == AcceptTypeHTMX}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), dbipCreditKey{}, credit)))
	})
}

// CreditDBIP marks the page as showing DB-IP's locations, and is whether it is a part of a page, whose drawn footer
// is sent the credit out of band
func CreditDBIP(ctx context.Context) bool {
	credit, ok := ctx.Value(dbipCreditKey{}).(*dbipCredit)
	if !ok {
		return false
	}
	credit.shown = true
	return credit.partial
}

// DBIPCredited is whether the page shows DB-IP's locations, so its footer credits DB-IP
func DBIPCredited(ctx context.Context) bool {
	credit, ok := ctx.Value(dbipCreditKey{}).(*dbipCredit)
	return ok && credit.shown
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
				slog.InfoContext(r.Context(), "Site rate limit exceeded", "remainingTime", remainingTime)
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
				slog.InfoContext(r.Context(), "Site unauthenticated rate limit exceeded", "remainingTime", remainingTime)
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
				slog.InfoContext(r.Context(), "Identity rate limit exceeded", "remainingTime", remainingTime)
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

// CountRoutes counts each request to a route of the router named name by the route's pattern, never its path, which
// can carry a token or a site. It goes in a router's signed in group, after its sign in, so only signed in requests
// to the group's own routes are counted, and no one can add routes to count.
func CountRoutes(tracker *usage.Tracker, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			next.ServeHTTP(w, r)
			if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
				tracker.Route(name+" "+r.Method+" "+rctx.RoutePattern(), time.Since(started))
			}
		})
	}
}
