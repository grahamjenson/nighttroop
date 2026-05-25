package httputils

import (
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"nighttroop.t1d.nz/utils/config"
)

func AddStandardMiddleware(router chi.Router, config config.ServerConfig) {
	router.Use(middleware.RequestID)

	if config.SentryDSN != "" {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		})
		router.Use(sentryMiddleware.Handle)
	}

	router.Use(slogchi.New(config.Logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestSize(500 * 1024)) // 500KB max body size, should be enough for any request and prevents abuse
	router.Use(middleware.Timeout(30 * time.Second))
}
