package nsutils

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	// Column is the column a value refused was for, as a document refused for a value out of its column's limits says
	Column string `json:"-"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func WriteError(w http.ResponseWriter, err *HTTPError) {
	if err == nil {
		err = ErrBadRequest
	}
	apiutils.WriteJSON(w, err.Status, map[string]any{"status": err.Status, "message": err.Message})
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
	ErrNotFound = &HTTPError{Status: http.StatusNotFound, Message: "record not found"}

	ErrGone = &HTTPError{Status: http.StatusGone, Message: "record deleted"}
	// ErrModifiedSince is a replace or patch of a record changed after the request's If-Unmodified-Since
	ErrModifiedSince = &HTTPError{Status: http.StatusPreconditionFailed, Message: "record modified since If-Unmodified-Since"}
	ErrInternal      = &HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
	ErrCannotModify  = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Field %s cannot be modified by the client", field)}
	}
	ErrRecordReadOnly = &HTTPError{Status: http.StatusForbidden, Message: "record is read-only (because soft deleted)"}
)

// Bad Body Data Errors
var (
	ErrBadQueryError = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad query parameters: %s", field)}
	}
	ErrBadRequest            = &HTTPError{Status: http.StatusBadRequest, Message: "Bad request"}
	ErrInvalidCollection     = &HTTPError{Status: http.StatusBadRequest, Message: "invalid collection"}
	BadDataBodyError         = &HTTPError{Status: http.StatusBadRequest, Message: "Bad or missing body"}
	BadDataBodyErrorWithHint = func(hint string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad or missing body: %s", hint)}
	}
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
	BadStringFieldError = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad or missing string field %s", field)}
	}
	BadBoolFieldError = func(field string) *HTTPError {
		return &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Bad or missing bool field %s", field)}
	}
)

// Too Many Rows Errors
var (
	ErrTooManyRecords  = &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Too many records: at most %d per request", consts.MaxV1BatchRecords)}
	// ErrTooManyBatchWrites is a batch that would take the site or identity past its limit on writes
	ErrTooManyBatchWrites = &HTTPError{Status: http.StatusTooManyRequests, Message: "Too many records written: send them again shortly"}
	ErrTooManyProfiles = &HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Too many profiles: at most %d in a store", consts.MaxProfilesPerStore)}
)

// Bad Updates
var (
	ErrCannotChangeApp        = &HTTPError{Status: http.StatusBadRequest, Message: "Field app cannot be modified by the client"}
	ErrCannotChangeEventType  = &HTTPError{Status: http.StatusBadRequest, Message: "Field eventType cannot be modified by the client"}
	ErrCannotChangeName       = &HTTPError{Status: http.StatusBadRequest, Message: "Field name cannot be modified by the client"}
	ErrCannotChangeType       = &HTTPError{Status: http.StatusBadRequest, Message: "Field type cannot be modified by the client"}
	ErrCannotChangeSubject    = &HTTPError{Status: http.StatusBadRequest, Message: "Field subject cannot be modified by the client"}
	ErrCannotChangeDate       = &HTTPError{Status: http.StatusBadRequest, Message: "Field date cannot be modified by the client"}
	ErrCannotChangeDevice     = &HTTPError{Status: http.StatusBadRequest, Message: "Field device cannot be modified by the client"}
	ErrCannotChangeIdentifier = &HTTPError{Status: http.StatusBadRequest, Message: "Field identifier cannot be modified by the client"}
	ErrCannotChangeUTCOffset  = &HTTPError{Status: http.StatusBadRequest, Message: "Field utcOffset cannot be modified by the client"}
	ErrCannotChangeCreatedAt  = &HTTPError{Status: http.StatusBadRequest, Message: "Field srvCreated cannot be modified by the client"}
	ErrCannotChangeModifiedAt = &HTTPError{Status: http.StatusBadRequest, Message: "Field srvModified cannot be modified by the client"}
	ErrCannotChangeModifiedBy = &HTTPError{Status: http.StatusBadRequest, Message: "Field modifiedBy cannot be modified by the client"}
)

// GetSQLTranslatedError is the response to the database refusing a value, or nil for any other error. The driver's
// words are logged, not sent, as they say how the data is kept: the client is told the column and why.
func GetSQLTranslatedError(err error) *HTTPError {
	reason := ""
	switch {
	case utilerrors.IsSQLUniqueViolationError(err):
		reason = "is already used"
	case utilerrors.IsSQLConstraintNotNull(err):
		reason = "is required"
	case utilerrors.IsSQLConstraintCheck(err) && strings.Contains(err.Error(), "length("):
		reason = "is too long"
	case utilerrors.IsSQLConstraintCheck(err):
		reason = "is out of range"
	case utilerrors.IsSQLConstraintForeignKey(err):
		reason = "refers to a record that does not exist"
	default:
		return nil
	}
	slog.Warn("database refused a value", "error", err)
	return BadDataBodyErrorWithHint(constraintColumn(err.Error()) + " " + reason)
}

var (
	sqlFunctionOrTable = regexp.MustCompile(`\w+[(.]`)
	sqlColumn          = regexp.MustCompile(`[a-z_][a-z0-9_]*`)
)

// constraintColumn is the first column named by the constraint the driver says failed, e.g. dose for "constraint
// failed: CHECK constraint failed: dose BETWEEN -500 AND 500 (275)" or "NOT NULL constraint failed: insulin_doses.dose",
// or "a value" when it names none, as a foreign key's does not
func constraintColumn(message string) string {
	const failed = "constraint failed"
	i := strings.LastIndex(message, failed)
	if i < 0 {
		return "a value"
	}
	if column := sqlColumn.FindString(sqlFunctionOrTable.ReplaceAllString(message[i+len(failed):], "")); column != "" {
		return column
	}
	return "a value"
}
