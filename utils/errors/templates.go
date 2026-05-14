package errors

import (
	"log/slog"
	"net/http"
)

func TemplateFailureError(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Error("Template Rendering Failed", allArgs...)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
