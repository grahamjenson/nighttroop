package config

import (
	"context"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/consts"
)

// setupLogging makes slog.Default() the single logger for the server. Everything logs through it,
// including request logs and the std log package. Use slog.With(...) for sub loggers with context, and
// slog.ErrorContext(ctx, ...) in a request, so it is logged with the request's ID, site and identity. Every log is
// also copied into aslog.Logs, which the logs DB writes to its logs table in batches.
func setupLogging(mode consts.ServerMode) {
	buffer := aslog.BufferHandler{Buffer: aslog.Logs}
	if mode == consts.Test {
		slog.SetDefault(slog.New(aslog.ContextHandler{Handler: slogmulti.Fanout(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Set minimum log level
		}), buffer)}))
		return
	}

	// 1. Create a JSON handler writing to standard output
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum log level
	})

	// The request's route is read as each error is logged, as it is only known once the request is routed
	sentryHandler := slogsentry.Option{
		Level:           slog.LevelError, // Only send errors to Sentry
		AttrFromContext: []func(context.Context) []slog.Attr{aslog.SentryRoute},
	}.NewSentryHandler()

	slog.SetDefault(slog.New(aslog.ContextHandler{Handler: slogmulti.Fanout(jsonHandler, sentryHandler, buffer)}))
}
