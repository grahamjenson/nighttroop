package main

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// Sentry reports a server's errors under its deployment's name, so each staging server's are apart
func TestSentryEnvironmentIsTheDeployment(t *testing.T) {
	t.Cleanup(func() { sentry.CurrentHub().BindClient(nil) })
	cfg := &config.ServerConfig{SentryDSN: "https://public@sentry.example.invalid/1", Deployment: "nighttroop-staging"}
	setupSentry(cfg)
	require.True(t, cfg.Status.SentryWorking)
	require.Equal(t, "nighttroop-staging", sentry.CurrentHub().Client().Options().Environment)
}
