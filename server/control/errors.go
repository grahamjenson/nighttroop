package control

import (
	"log/slog"
	"net/http"
)

// We have a very generic error to not give any hints about how far through it was rejected
func genericLoginError(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	slog.Info("Control User Login Failed", allArgs...)
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)

}
