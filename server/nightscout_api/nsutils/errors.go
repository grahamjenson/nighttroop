package nsutils

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func WriteError(w http.ResponseWriter, err *HTTPError) {
	if err == nil {
		err = ErrBadRequest
	}
	WriteJSON(w, err.Status, map[string]any{"status": err.Status, "message": err.Message})
}

//////
// ERRORS
//////

// Permission and Login Errors

var (
	ErrMissingCredentials = &HTTPError{Status: http.StatusUnauthorized, Message: "Missing or bad access token or JWT"}
	ErrMissingPermission  = func(permission string) *HTTPError {
		return &HTTPError{Status: http.StatusForbidden, Message: "Missing permission " + permission}
	}
	ErrNotImplemented = &HTTPError{Status: http.StatusNotImplemented, Message: "Not implemented"}
	ErrBadToken       = &HTTPError{Status: http.StatusUnauthorized, Message: "badToken"}
	ErrBadJWT         = &HTTPError{Status: http.StatusUnauthorized, Message: "bad jwt"}
)

// Query Errors
var (
	ErrSortParamsConflict  = &HTTPError{Status: http.StatusBadRequest, Message: "Parameters sort and sort_desc cannot be combined"}
	ErrLimitOutOfTolerance = &HTTPError{Status: http.StatusBadRequest, Message: "Parameter limit out of tolerance"}
	ErrSkipOutOfTolerance  = &HTTPError{Status: http.StatusBadRequest, Message: "Parameter skip out of tolerance"}
)

// General Errors

var (
	ErrNotFound     = &HTTPError{Status: http.StatusNotFound, Message: "record not found"}
	ErrGone         = &HTTPError{Status: http.StatusGone, Message: "record deleted"}
	ErrInternal     = &HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
	ErrCannotModify = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Field %s cannot be modified by the client", field)}
	}
)

// Bad Body Data Errors
var (
	ErrBadQueryError = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad query parameters: %s", field)}
	}
	ErrBadRequest        = &HTTPError{Status: http.StatusBadRequest, Message: "Bad request"}
	ErrInvalidCollection = &HTTPError{Status: http.StatusBadRequest, Message: "invalid collection"}
	BadDataBodyError     = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing body"}
	BadIdentifierError   = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing identifier field"}
	BadDateFieldError    = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing date field"}
	BadUTCTOffsetError   = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing utcOffset field"}
	BadCreatedAtError    = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing created_at field"}
	BadDateStringError   = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing dateString field"}
	BadOffsetError       = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing utcOffset field"}
	BadAppFieldError     = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing app field"}
	BadNumericFieldError = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad or missing numeric field %s", field)}
	}
)

// Bad Updates
var (
	ErrCannotChangeApp        = &HTTPError{Status: http.StatusBadRequest, Message: "Field app cannot be modified by the client"}
	ErrCannotChangeEventType  = &HTTPError{Status: http.StatusBadRequest, Message: "Field eventType cannot be modified by the client"}
	ErrCannotChangeType       = &HTTPError{Status: http.StatusBadRequest, Message: "Field type cannot be modified by the client"}
	ErrCannotChangeSubject    = &HTTPError{Status: http.StatusBadRequest, Message: "Field subject cannot be modified by the client"}
	ErrCannotChangeDate       = &HTTPError{Status: http.StatusBadRequest, Message: "Field date cannot be modified by the client"}
	ErrCannotChangeDevice     = &HTTPError{Status: http.StatusBadRequest, Message: "Field device cannot be modified by the client"}
	ErrCannotChangeIdentifier = &HTTPError{Status: http.StatusBadRequest, Message: "Field identifier cannot be modified by the client"}
	ErrCannotChangeUTCTOffset = &HTTPError{Status: http.StatusBadRequest, Message: "Field utcOffset cannot be modified by the client"}
	ErrCannotChangeCreatedAt  = &HTTPError{Status: http.StatusBadRequest, Message: "Field srvCreated cannot be modified by the client"}
	ErrCannotChangeModifiedAt = &HTTPError{Status: http.StatusBadRequest, Message: "Field srvModified cannot be modified by the client"}
	ErrCannotChangeModifiedBy = &HTTPError{Status: http.StatusBadRequest, Message: "Field modifiedBy cannot be modified by the client"}
	ErrCannotChangeIsValid    = &HTTPError{Status: http.StatusBadRequest, Message: "Field isValid cannot be modified by the client"}
)
