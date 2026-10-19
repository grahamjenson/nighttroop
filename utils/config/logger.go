package config

import (
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"nighttroop.t1d.nz/utils/consts"
)

// setupLogging makes slog.Default() the single logger for the server. Everything logs through it,
// including request logs and the std log package. Use slog.With(...) for sub loggers with context.
func setupLogging(mode consts.ServerMode) {
	if mode == consts.Test {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Set minimum log level
		})))
		return
	}

	// 1. Create a JSON handler writing to standard output
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum log level
	})

	sentryHandler := slogsentry.Option{
		Level: slog.LevelError, // Only send errors to Sentry
	}.NewSentryHandler()

	slog.SetDefault(slog.New(slogmulti.Fanout(jsonHandler, sentryHandler)))
}
