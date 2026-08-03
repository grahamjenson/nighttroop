package config

import (
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"nighttroop.t1d.nz/utils/consts"
)

func setupLogging(mode consts.ServerMode) *slog.Logger {
	if mode == consts.Test {
		// In test mode, use a no-op logger to avoid cluttering test output.
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Set minimum log level
		}))
	}

	// 1. Create a JSON handler writing to standard output
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum log level
	})

	sentryHandler := slogsentry.Option{
		Level: slog.LevelError, // Only send errors to Sentry
	}.NewSentryHandler()

	logger := slog.New(slogmulti.Fanout(jsonHandler, sentryHandler))

	slog.SetDefault(logger)
	return logger
}
