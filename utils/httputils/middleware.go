package httputils

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/usage"
)

// requestTimeout cancels a request that has taken too long
var requestTimeout = 30 * time.Second

// timeoutUnlessLongLived cancels a request after requestTimeout, but not one held open on purpose, as a live update's
// event stream or a Socket.IO connection, which end themselves: the timeout cut each stream off after 30 seconds, so
// its page reconnected and read everything again
func timeoutUnlessLongLived(next http.Handler) http.Handler {
	timed := middleware.Timeout(requestTimeout)(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if LongLived(r) {
			next.ServeHTTP(w, r)
			return
		}
		timed.ServeHTTP(w, r)
	})
}

// AddStandardMiddleware adds what every router has, recording its server errors in the logs DB when there is one
func AddStandardMiddleware(router chi.Router, config *config.ServerConfig, controlDB *db.ControlDB) {
	router.Use(middleware.RequestID)
	router.Use(withRequestLogAttrs)
	// Outside the errors' records, so a request that fails is recorded with what it sent
	router.Use(KeepErrorRequestBodies)
	if controlDB != nil {
		// Before the recoverer, so a panic's 500 is recorded too
		router.Use(RecordServerErrors(controlDB.Logs))
	}

	if config.SentryDSN != "" {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		})
		router.Use(sentryMiddleware.Handle)
		router.Use(reportErrorRequestBodies)
	}

	router.Use(NewLogger(nil))
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestSize(500 * 1024)) // 500KB max body size, should be enough for any request and prevents abuse
	router.Use(timeoutUnlessLongLived)
	// Gzip text responses, like the pages, styles, scripts and JSON, for clients that accept it
	router.Use(middleware.Compress(5))
	if controlDB != nil {
		// Inside the compression, so it reads why a request was refused as it was written
		router.Use(RecordClientErrors(controlDB.Logs))
	}
	router.Use(withSupportEmail(config.SupportEmail))
	router.Use(withFeedbackURL(config.FeedbackURL))
	router.Use(withTestSite(config.TestSite))
	router.Use(withDBIPCredit)
	if controlDB != nil && config.CaptureRequests {
		// Inside the compression, so the response body is captured as the handler wrote it
		router.Use(CaptureRequests(controlDB.Logs))
	}
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

type feedbackURLKey struct{}

// withFeedbackURL puts where users give feedback in each request's context, for its pages' footer
func withFeedbackURL(url string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), feedbackURLKey{}, url)))
		})
	}
}

// FeedbackURL is where users give feedback, or empty when the server has none
func FeedbackURL(ctx context.Context) string {
	url, _ := ctx.Value(feedbackURLKey{}).(string)
	return url
}

type testSiteKey struct{}

// withTestSite puts whether the server is a test site in each request's context, for its pages' banner
func withTestSite(testSite bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), testSiteKey{}, testSite)))
		})
	}
}

// TestSite is whether the server is a test site, whose pages say its data is wiped
func TestSite(ctx context.Context) bool {
	testSite, _ := ctx.Value(testSiteKey{}).(bool)
	return testSite
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

type errorReasonKey struct{}

// withErrorReason is the request with room for SetErrorReason, and that room, which it shares with an outer one's
func withErrorReason(r *http.Request) (*http.Request, *string) {
	if reason, ok := r.Context().Value(errorReasonKey{}).(*string); ok {
		return r, reason
	}
	reason := new(string)
	return r.WithContext(context.WithValue(r.Context(), errorReasonKey{}, reason)), reason
}

// SetErrorReason keeps why the request failed, in the server's own words, for its errors page, when its response
// tells the user something plainer
func SetErrorReason(ctx context.Context, reason string) {
	if kept, ok := ctx.Value(errorReasonKey{}).(*string); ok {
		*kept = reason
	}
}

// RecordServerErrors records each response with a 5xx status in the logs DB's errors, with its site when it has one,
// why, when the handler said with SetErrorReason, and what the request sent, its credentials redacted.
// It records the route's pattern rather than the path, as a Nightscout path can carry a token.
func RecordServerErrors(logs *db.LogsDB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			r, reason := withErrorReason(r)
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
			logs.ServerError(context.WithoutCancel(r.Context()), siteID, db.ServerErrorAttributes{Status: ww.Status(), Method: r.Method, Route: route,
				RequestID: middleware.GetReqID(r.Context()), Reason: clip(*reason), RequestBody: ErrorRequestBody(r.Context())})
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
				writeTooManyRequests(w, remainingTime)
				slog.InfoContext(r.Context(), "Site rate limit exceeded", "remainingTime", remainingTime)
				return
			}
			// A site flagged for abuse spends its flagged write limit too, whatever the identity, API or route
			if siteContext := GetSiteContext(r.Context()); siteContext.HTTPAction == HTTPActionWrite {
				if allowed, remainingTime := FlaggedSiteWritesWithinRateLimit(siteContext, 1); !allowed {
					writeTooManyRequests(w, remainingTime)
					_, _ = w.Write([]byte(SiteOverStorageMessage))
					slog.InfoContext(r.Context(), "Flagged site write rate limit exceeded", "remainingTime", remainingTime)
					return
				}
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
				writeTooManyRequests(w, remainingTime)
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
				writeTooManyRequests(w, remainingTime)
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
				// The response is sent, so a failure to record is only logged. It is recorded even when the client has
				// already gone, as htmx cancels a request a newer one replaces, as the identity was still seen.
				if err := controlDB.UpdateSiteAndIdentityDateFlags(context.WithoutCancel(r.Context()), siteContext.SiteModel, siteContext.SiteDB, siteContext.IdentityModel, true, wasWrite); err != nil {
					slog.ErrorContext(r.Context(), "Cannot update site and identity date flags", "error", err)
				}
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

// writeTooManyRequests refuses a request over a rate limit, saying in Retry-After how many seconds until one is let
// through, at least one, so a client such as AAPS waits rather than retrying at once
func writeTooManyRequests(w http.ResponseWriter, seconds int) {
	SetRetryAfter(w, seconds)
	w.WriteHeader(http.StatusTooManyRequests)
}

// SetRetryAfter says when a refused request can be sent again, at least a second from now
func SetRetryAfter(w http.ResponseWriter, seconds int) {
	w.Header().Set("Retry-After", strconv.Itoa(max(1, seconds)))
}
