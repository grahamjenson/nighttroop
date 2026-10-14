package siteapi

import (
	"net/http"
	"strconv"

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
	w.WriteHeader(e.Status)
	// TODO: reply with appropriate content based HTMX and HTML
	switch acceptType {
	case httputils.AcceptTypeHTML:
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(e.Message))
	case httputils.AcceptTypeHTMX:
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Retarget", "#error-container")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Write([]byte(e.Message))
	case httputils.AcceptTypeJSON:
		writeJSON(w, map[string]any{
			"status":  e.Status,
			"message": e.Message,
		})
	default:
		w.Header().Set("Content-Type", "text/plain")
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
