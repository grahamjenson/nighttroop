package errors

import (
	"errors"
	"log/slog"
	"net/http"

	sql "github.com/mattn/go-sqlite3"
)

func HandleInternalError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Info(message, allArgs...)
	http.Error(w, "Internal Error", statusCode)
}

func IsUniqueViolationError(err error) bool {
	var sqliteErr sql.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.ExtendedCode {
		case sql.ErrConstraintUnique,
			sql.ErrConstraintPrimaryKey:
			return true
		}

	}
	return false
}
