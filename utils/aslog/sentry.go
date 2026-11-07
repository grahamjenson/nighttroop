package aslog

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
)

// sentryExtra is where slog-sentry puts what is logged with an error, which Sentry shows but cannot search
const sentryExtra = "extra"

// sentryTags are what is logged with an error that Sentry tags its event with, so issues can be searched and counted
// by them, as every error of one site, identity, request, job or route
var sentryTags = []string{"site_id", "identity_id", "role", "request_id", "job", "job_id", "sync_type", "batch", "method", "route"}

// SentryRoute is the method and route of the request ctx is for, as far as it is routed, for its errors to say where
// they were; a request's path is not, as it can hold a token
func SentryRoute(ctx context.Context) []slog.Attr {
	routeContext := chi.RouteContext(ctx)
	if routeContext == nil || routeContext.RoutePattern() == "" {
		return nil
	}
	return []slog.Attr{slog.String("method", routeContext.RouteMethod), slog.String("route", routeContext.RoutePattern())}
}

// SentryEvent moves what is logged with an error to where Sentry can search and count it: its site, identity, request
// and job as tags, its identity as its user, and its request's method and route as its transaction. An error without
// a stack trace, as most are, is given the stack it was logged from, so its issue shows where it happened.
func SentryEvent(event *sentry.Event) *sentry.Event {
	extra := event.Contexts[sentryExtra]
	if event.Tags == nil {
		event.Tags = map[string]string{}
	}
	for _, key := range sentryTags {
		if value, ok := extra[key]; ok {
			event.Tags[key] = fmt.Sprint(value)
		}
	}
	if identityID, ok := extra["identity_id"]; ok {
		// A control identity has no site, and its IDs are its own, so its user is told apart from a site's
		site := "control"
		if siteID, ok := extra["site_id"]; ok {
			site = fmt.Sprint(siteID)
		}
		event.User.ID = fmt.Sprintf("%s/%v", site, identityID)
		if name, ok := extra["identity_name"]; ok {
			event.User.Username = fmt.Sprint(name)
		}
		if event.User.Data == nil {
			event.User.Data = map[string]string{}
		}
		event.User.Data["site_id"] = site
		if role, ok := extra["role"]; ok {
			event.User.Data["role"] = fmt.Sprint(role)
		}
	}
	if method, route := extra["method"], extra["route"]; method != nil && route != nil {
		event.Transaction = fmt.Sprintf("%v %v", method, route)
	}
	// Sentry gives an error without a stack of its own the stack it was converted on, which ends in slog and the
	// handlers that send it on to Sentry; every event would share those, so it ends where slog was called instead
	for _, exception := range event.Exception {
		endAtSlog(exception.Stacktrace)
	}
	if len(event.Exception) == 0 && len(event.Threads) == 0 && (event.Level == sentry.LevelError || event.Level == sentry.LevelFatal) {
		stack := sentry.NewStacktrace()
		endAtSlog(stack)
		event.Threads = []sentry.Thread{{Stacktrace: stack, Current: true}}
	}
	return event
}

// endAtSlog leaves out of a stack the frames from where slog was called on, if it was
func endAtSlog(stack *sentry.Stacktrace) {
	if stack == nil {
		return
	}
	for i, frame := range stack.Frames {
		if frame.Module == "log/slog" {
			stack.Frames = stack.Frames[:i]
			return
		}
	}
}

// SentryContext is what is logged with ctx, as Sentry's extra, for an event captured outside slog, such as a client
// error's report, so SentryEvent gives it the same tags and user as an error logged with ctx
func SentryContext(ctx context.Context) sentry.Context {
	extra := sentry.Context{}
	for _, attr := range append(attrsOf(ctx), SentryRoute(ctx)...) {
		extra[attr.Key] = attr.Value.Any()
	}
	return extra
}
