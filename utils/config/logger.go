package config

import (
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
)

func setupLogging() *slog.Logger {
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
