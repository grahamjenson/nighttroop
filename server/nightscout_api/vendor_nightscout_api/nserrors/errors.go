package nserrors

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingCredentials indicates that a request did not include any usable
	// authentication material.
	ErrMissingCredentials = errors.New("missing credentials")
	// ErrBadToken indicates that an access token or API secret was provided but
	// did not match a known credential.
	ErrBadToken = errors.New("bad token")
	// ErrBadJWT indicates that a bearer token looked like a JWT but could not be
	// parsed or verified by the authentication backend.
	ErrBadJWT = errors.New("bad jwt")
)

var (
	ErrNotFound = errors.New("record not found")
	ErrGone     = errors.New("record deleted")
)

var (
	BadDataBodyError   = fmt.Errorf("Bad or missing body")
	BadDateFieldError  = fmt.Errorf("Bad or missing date field")
	BadUTCTOffsetError = fmt.Errorf("Bad or missing utcOffset field")
	BadCreatedAtError  = fmt.Errorf("Bad or missing created_at field")
	BadDateStringError = fmt.Errorf("Bad or missing dateString field")
	BadOffsetError     = fmt.Errorf("Bad or missing utcOffset field")
)
