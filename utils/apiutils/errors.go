package apiutils

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
)

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) Write(acceptType httputils.AcceptType, w http.ResponseWriter) string {
	// Headers set after WriteHeader are dropped, so each case sets them first.
	// HTMX forms choose where errors go with hx-status.
	switch acceptType {
	case httputils.AcceptTypeHTML, httputils.AcceptTypeHTMX:
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(e.Status)
		w.Write([]byte(e.Message))
	case httputils.AcceptTypeJSON:
		WriteJSON(w, e.Status, map[string]any{
			"status":  e.Status,
			"message": e.Message,
		})
	default:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(e.Status)
		w.Write([]byte(e.Message))
	}

	return e.Message
}

// Auth errors
var (
	ErrMissingCredentials = &HTTPError{Status: http.StatusUnauthorized, Message: "Missing or bad access token or JWT"}
	ErrMissingPermission  = func(permission string) *HTTPError {
		return &HTTPError{Status: http.StatusForbidden, Message: "Missing permission " + permission}
	}
	ErrNotImplemented = &HTTPError{Status: http.StatusNotImplemented, Message: "Not implemented"}
	ErrForbidden      = &HTTPError{Status: http.StatusForbidden, Message: "Forbidden"}
	ErrBadToken       = &HTTPError{Status: http.StatusUnauthorized, Message: "badToken"}
	ErrBadJWT         = &HTTPError{Status: http.StatusUnauthorized, Message: "bad jwt"}
)

// API errors
var (
	ErrBadRequest          = &HTTPError{Status: http.StatusBadRequest, Message: "Bad request"}
	ErrNotFound            = &HTTPError{Status: http.StatusNotFound, Message: "Not found"}
	ErrInternalServerError = &HTTPError{Status: http.StatusInternalServerError, Message: "Internal server error"}
)

// body errors
var (
	ErrUnknownContent  = &HTTPError{Status: http.StatusUnsupportedMediaType, Message: "Unknown content type"}
	ErrBadBody         = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing request body"}
	ErrBadBodyWithHint = func(hint string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing request body: " + hint}
	}
	ErrBadForm = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing form data"}
)

var (
	ErrMissingRequiredField         = &HTTPError{Status: http.StatusBadRequest, Message: "Missing required field"}
	ErrMissingRequiredFieldWithHint = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: "Missing required field " + field}
	}
	ErrConstraintCheckFailed         = &HTTPError{Status: http.StatusBadRequest, Message: "Constraint check failed"}
	ErrConstraintCheckFailedWithHint = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: "Constraint check failed: " + field}
	}
	ErrUniqueConstraintViolation = &HTTPError{Status: http.StatusConflict, Message: "Unique constraint violation"}
)

var (
	ErrRateLimitExceeded = func(seconds int) *HTTPError {
		return &HTTPError{Status: http.StatusTooManyRequests, Message: "Too Many Requests wait " + strconv.Itoa(seconds) + " seconds"}
	}
)

// ErrorFor is the response to an error reading or writing what, e.g. "glucose_readings": nil for none, its own
// HTTPError, not found for a missing row, or the database refusing a value says why; anything else is logged
func ErrorFor(err error, what string) *HTTPError {
	var herr *HTTPError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &herr):
		return herr
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	case utilerrors.IsSQLUniqueViolationError(err):
		return ErrUniqueConstraintViolation
	case utilerrors.IsSQLConstraintCheck(err):
		return ErrConstraintCheckFailedWithHint(err.Error())
	}
	slog.Error("Error writing "+what, "error", err)
	return ErrInternalServerError
}
