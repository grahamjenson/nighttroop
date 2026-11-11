package httputils

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// ShedLoad works on at most inFlight requests at once, and has at most backlog more wait up to wait for one to finish;
// past those a request is refused at once with a 503 and Retry-After, by chi's throttle. A server that falls behind,
// such as when its CPU is throttled, then answers what it can in time, rather than queueing every request until they
// all time out together. A request exempt is never held, such as a health check or a live update held open for minutes.
func ShedLoad(inFlight int, backlog int, wait time.Duration, exempt func(*http.Request) bool) func(http.Handler) http.Handler {
	throttle := middleware.ThrottleWithOpts(middleware.ThrottleOpts{Limit: inFlight, BacklogLimit: backlog, BacklogTimeout: wait,
		StatusCode: http.StatusServiceUnavailable, RetryAfterFn: func(bool) time.Duration { return retryAfter }})
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exempt(r) {
				next.ServeHTTP(w, r)
				return
			}
			// A refused request never reaches the routers that log each request, so it is logged here, unless its
			// client went away
			served := false
			throttle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				served = true
				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
			if !served && r.Context().Err() == nil {
				slog.InfoContext(r.Context(), "Server busy, request refused", "path", r.URL.Path)
			}
		})
	}
}

// retryAfter is how soon a client refused as the server is busy is asked to try again
const retryAfter = 5 * time.Second

// RefuseBusy answers 503, to try again in a few seconds, as when the server cannot answer the request in time
func RefuseBusy(w http.ResponseWriter, r *http.Request, why string) {
	slog.InfoContext(r.Context(), "Server busy, request refused", "why", why, "path", r.URL.Path)
	w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
	http.Error(w, "Server busy, try again shortly", http.StatusServiceUnavailable)
}

// ControlLogsDBPath is the control site's download of the logs DB, which can take minutes to copy and send
const ControlLogsDBPath = "/logs-db"

// LongLived is a request held open, a live update, a Socket.IO connection or the control site's download of the logs
// DB, or the health check, which must answer while the server is busy so it is not taken for down
func LongLived(r *http.Request) bool {
	return r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/socket.io/") ||
		strings.Contains(r.Header.Get("Accept"), "text/event-stream") ||
		(r.URL.Path == ControlLogsDBPath && strings.HasPrefix(r.Host, "control."))
}

// ShedLoadReserving is ShedLoad keeping a quarter of its requests, in flight and waiting, for those reserved, such as
// uploads from devices, so they are still worked on while the others, such as page loads, fill the rest
func ShedLoadReserving(inFlight int, backlog int, wait time.Duration, exempt func(*http.Request) bool, reserved func(*http.Request) bool) func(http.Handler) http.Handler {
	keptInFlight, keptBacklog := max(inFlight/4, 1), max(backlog/4, 1)
	forReserved := ShedLoad(keptInFlight, keptBacklog, wait, exempt)
	forRest := ShedLoad(max(inFlight-keptInFlight, 1), max(backlog-keptBacklog, 0), wait, exempt)
	return func(next http.Handler) http.Handler {
		reservedNext, restNext := forReserved(next), forRest(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reserved(r) {
				reservedNext.ServeHTTP(w, r)
				return
			}
			restNext.ServeHTTP(w, r)
		})
	}
}
