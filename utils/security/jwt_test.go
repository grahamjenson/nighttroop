package security_test

import (
	"context"
	"log/slog"
	"testing"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// levelHandler records the levels of the records it is given
type levelHandler struct {
	slog.Handler
	levels *[]slog.Level
}

func (h levelHandler) Handle(_ context.Context, r slog.Record) error {
	*h.levels = append(*h.levels, r.Level)
	return nil
}

func (h levelHandler) Enabled(context.Context, slog.Level) bool { return true }

// A garbage token is client-caused and must not be logged at Error, which setupLogging fans out to Sentry
func TestVerifyUserJWTTokenGarbageIsNotLoggedAtError(t *testing.T) {
	levels := []slog.Level{}
	previous := slog.Default()
	slog.SetDefault(slog.New(levelHandler{Handler: previous.Handler(), levels: &levels}))
	defer slog.SetDefault(previous)

	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	if _, _, _, _, err := cfg.VerifyUserJWTToken("not.a.jwt"); err == nil {
		t.Fatal("expected a garbage token to fail verification")
	}
	if len(levels) == 0 {
		t.Fatal("expected the failed verification to be logged")
	}
	for _, level := range levels {
		if level >= slog.LevelError {
			t.Fatalf("garbage token was logged at %v", level)
		}
	}
}
