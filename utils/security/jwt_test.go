package security_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
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

// Each kind of JWT verifies as what it was made for, with its identity's credentials, and only by its own kind's Verify
func TestJWTTokensRoundTrip(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	admin := &controldb.ControlIdentity{ID: 3, CredentialsVersion: 2}
	identity := &controldb.SiteIdentity{ID: 7, CredentialsVersion: 5}
	adminCredentials := security.Credentials{IdentityID: 3, Version: 2}
	identityCredentials := security.Credentials{IdentityID: 7, Version: 5}

	control, expires, err := cfg.GenerateControlJWTToken(admin, "ctl-admin", "admin")
	if err != nil || time.Until(expires) < 59*time.Minute {
		t.Fatalf("control token: %v, expires %v", err, expires)
	}
	name, role, credentials, err := cfg.VerifyControlJWTToken(control)
	if err != nil || name != "ctl-admin" || role != "admin" || credentials != adminCredentials {
		t.Fatalf("control token verified as %q %q %v: %v", name, role, credentials, err)
	}

	user, _, err := cfg.GenerateUserJWTToken(identity, "abcdefgh", "admin", "device-read")
	if err != nil {
		t.Fatal(err)
	}
	subdomain, name, role, credentials, err := cfg.VerifyUserJWTToken(user)
	if err != nil || subdomain != "abcdefgh" || name != "admin" || role != "device-read" || credentials != identityCredentials {
		t.Fatalf("user token verified as %q %q %q %v: %v", subdomain, name, role, credentials, err)
	}

	preAuth, expires, err := cfg.GeneratePreAuthJWTToken(identity, "abcdefgh", "admin")
	if err != nil || time.Until(expires) > 5*time.Minute {
		t.Fatalf("pre-auth token: %v, expires %v", err, expires)
	}
	name, credentials, err = cfg.VerifyPreAuthJWTToken(preAuth, "abcdefgh")
	if err != nil || name != "admin" || credentials != identityCredentials {
		t.Fatalf("pre-auth token verified as %q %v: %v", name, credentials, err)
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

// A JWT's credentials are current only for the identity it was issued to, at the version of its credentials then, so a
// change to them, or the identity deleted and made again, refuses it
func TestCredentialsCheckCurrent(t *testing.T) {
	identity := &controldb.SiteIdentity{ID: 7, CredentialsVersion: 5}
	tests := []struct {
		name        string
		credentials security.Credentials
		current     bool
	}{
		{"current", security.Credentials{IdentityID: 7, Version: 5}, true},
		{"before a change", security.Credentials{IdentityID: 7, Version: 4}, false},
		{"another identity", security.Credentials{IdentityID: 8, Version: 5}, false},
		{"none", security.Credentials{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.credentials.CheckCurrent(identity); (err == nil) != tt.current {
				t.Fatalf("%v checked against %v: %v", tt.credentials, identity, err)
			}
		})
	}
}
