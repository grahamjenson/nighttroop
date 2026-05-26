package errors

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
)

// We have a very generic error to not give any hints about how far through it was rejected
func GenericAdminLoginError(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Info("Control User Login Failed", allArgs...)
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)

}

// We have a very generic error to not give any hints about how far through it was rejected
func GenericIdentityLoginError(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Info("Identity Login Failed", allArgs...)
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func localStack() string {
	const depth = 32
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)

	frames := runtime.CallersFrames(pcs[:n])

	var out []string

	for {
		f, more := frames.Next()

		if strings.Contains(f.File, "nighttroop") {
			out = append(out, fmt.Sprintf("%s:%d", f.File, f.Line))
		}

		if !more {
			break
		}
	}

	return strings.Join(out, " → ")
}
