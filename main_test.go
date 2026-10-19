package main

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"nighttroop.t1d.nz/utils/config"
)

func TestSetupSentryReportsStatus(t *testing.T) {
	// Reset the global Sentry client so later tests are unaffected
	t.Cleanup(func() { _ = sentry.Init(sentry.ClientOptions{}) })

	tests := []struct {
		name string
		dsn  string
		want bool
	}{
		{"not set up", "", false},
		{"invalid dsn", "not a dsn", false},
		{"set up", "https://public@sentry.example.com/1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{SentryDSN: tt.dsn}
			setupSentry(cfg)
			assert.Equal(t, tt.want, cfg.Status.SentryWorking)
		})
	}
}
