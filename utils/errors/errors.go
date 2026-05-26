package errors

import (
	"log/slog"
	"net/http"
)

func HandleInternalError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Info(message, allArgs...)
	http.Error(w, "Internal Error", statusCode)
}
