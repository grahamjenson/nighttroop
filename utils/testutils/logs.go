package testutils

import (
	"log/slog"
	"testing"
)

// UseLogHandler sends all slog output to handler until the test or fuzz target ends,
// then restores the previous default so other tests are unaffected
func UseLogHandler(t testing.TB, handler slog.Handler) {
	t.Helper()
	previous := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(previous) })
}
