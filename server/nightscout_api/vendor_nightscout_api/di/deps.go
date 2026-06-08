package di

import (
	"errors"
	"net/http"
)

type Dependencies struct {
	Auth  Manager
	Store Store
}

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

// Identity is the resolved authorization context for a request.
type Identity struct {
	Name        string
	Permissions []string
}

// Manager defines the authentication and authorization hooks required by the
// API routers in this module.
//
// The library intentionally depends on this interface instead of shipping a
// production auth implementation, so applications can map it to their own
// user, token, and permission model.
type Manager interface {
	Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *Identity)) http.HandlerFunc
	HasPermission(identity Identity, permission string) bool
	AuthenticateExplicit(r *http.Request) (*Identity, error)
}
