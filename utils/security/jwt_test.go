package security_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

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

// Each kind of JWT verifies as what it was made for, and only by its own kind's Verify
func TestJWTTokensRoundTrip(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	before := time.Now().UnixMilli()

	control, expires, err := cfg.GenerateControlJWTToken("ctl-admin", "admin")
	if err != nil || time.Until(expires) < 59*time.Minute {
		t.Fatalf("control token: %v, expires %v", err, expires)
	}
	name, role, issuedAtMs, err := cfg.VerifyControlJWTToken(control)
	if err != nil || name != "ctl-admin" || role != "admin" || issuedAtMs < before {
		t.Fatalf("control token verified as %q %q %d: %v", name, role, issuedAtMs, err)
	}

	user, _, err := cfg.GenerateUserJWTToken("abcdefgh", "admin", "device-read")
	if err != nil {
		t.Fatal(err)
	}
	subdomain, name, role, issuedAtMs, err := cfg.VerifyUserJWTToken(user)
	if err != nil || subdomain != "abcdefgh" || name != "admin" || role != "device-read" || issuedAtMs < before {
		t.Fatalf("user token verified as %q %q %q %d: %v", subdomain, name, role, issuedAtMs, err)
	}

	preAuth, expires, err := cfg.GeneratePreAuthJWTToken("abcdefgh", "admin")
	if err != nil || time.Until(expires) > 5*time.Minute {
		t.Fatalf("pre-auth token: %v, expires %v", err, expires)
	}
	name, issuedAtMs, err = cfg.VerifyPreAuthJWTToken(preAuth, "abcdefgh")
	if err != nil || name != "admin" || issuedAtMs < before {
		t.Fatalf("pre-auth token verified as %q %d: %v", name, issuedAtMs, err)
	}
	if _, _, err := cfg.VerifyPreAuthJWTToken(preAuth, "otherste"); err == nil {
		t.Fatal("a pre-auth token verified for another site")
	}

	if _, _, _, err := cfg.VerifyControlJWTToken(user); err == nil {
		t.Fatal("a user token verified as a control token")
	}
	if _, _, _, _, err := cfg.VerifyUserJWTToken(control); err == nil {
		t.Fatal("a control token verified as a user token")
	}
	if _, _, err := cfg.VerifyPreAuthJWTToken(user, "abcdefgh"); err == nil {
		t.Fatal("a user token verified as a pre-auth token")
	}
	if _, _, _, _, err := cfg.VerifyUserJWTToken(preAuth); err == nil {
		t.Fatal("a pre-auth token verified as a user token")
	}
}
