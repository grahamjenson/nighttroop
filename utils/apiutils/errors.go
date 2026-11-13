package apiutils

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
)

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	// Hint is the part of the message written for users, which a page shows in its language without the message's start
	Hint *i18n.Error `json:"-"`
	// Internal is a message in the server's own words, such as the database's, which a page does not show its user
	Internal bool `json:"-"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// UserMessage is what a page shows its user: the hint in the context's language, or else the message
func (e *HTTPError) UserMessage(ctx context.Context) string {
	if e.Hint == nil {
		return e.Message
	}
	return e.Hint.In(ctx)
}

// withHint is the error whose message starts with start then says the hint in English. A hint not written for
// people, as an error from a library, is shown as it is.
func withHint(status int, start string, hint error) *HTTPError {
	var message *i18n.Error
	if !errors.As(hint, &message) {
		message = i18n.NewError("errors.as_written", i18n.Data{"Text": hint.Error()})
	}
	return &HTTPError{Status: status, Message: start + message.Error(), Hint: message}
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
	ErrUnknownContent = &HTTPError{Status: http.StatusUnsupportedMediaType, Message: "Unknown content type"}
	ErrBadBody        = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing request body"}
	// ErrBadBodyWithHint says why the body is refused: an *i18n.Error, so a page shows it in its language
	ErrBadBodyWithHint = func(hint error) *HTTPError {
		return withHint(http.StatusBadRequest, "Bad or missing request body: ", hint)
	}
	ErrBadForm = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing form data"}
)

var (
	ErrMissingRequiredField = &HTTPError{Status: http.StatusBadRequest, Message: "Missing required field"}
	// ErrMissingRequiredFieldWithHint names the fields missing, by the id of a message saying so, e.g. "Missing
	// required field name"
	ErrMissingRequiredFieldWithHint = func(id string) *HTTPError {
		return withHint(http.StatusBadRequest, "", i18n.NewError(id))
	}
	ErrConstraintCheckFailed = &HTTPError{Status: http.StatusBadRequest, Message: "Constraint check failed"}
	// ErrConstraintCheckFailedWithHint says which value is refused and why: an *i18n.Error, so a page shows it in its
	// language
	ErrConstraintCheckFailedWithHint = func(hint error) *HTTPError {
		return withHint(http.StatusBadRequest, "Constraint check failed: ", hint)
	}
	// ErrDatabaseRefused is the database refusing a value, in its own words, which are kept from users
	ErrDatabaseRefused = func(err error) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: "Constraint check failed: " + err.Error(), Internal: true}
	}
	ErrUniqueConstraintViolation = &HTTPError{Status: http.StatusConflict, Message: "Unique constraint violation"}
	// ErrClientClosedRequest is a request its client stopped waiting for, as htmx does a request a newer one replaces,
	// with nginx's status for it: no one reads it, and it is not the server's error
	ErrClientClosedRequest = &HTTPError{Status: 499, Message: "Client Closed Request"}
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
		return ErrDatabaseRefused(err)
	case errors.Is(err, context.Canceled):
		slog.Info("Request cancelled reading or writing "+what, "error", err)
		return ErrClientClosedRequest
	}
	slog.Error("Error writing "+what, "error", err)
	return ErrInternalServerError
}
