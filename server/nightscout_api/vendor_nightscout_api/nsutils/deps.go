package nsutils

import (
	"context"
	"net/http"
)

type Dependencies struct {
	Auth  AuthManager
	Store Store
}

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
type AuthManager interface {
	Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *Identity)) http.HandlerFunc
	HasPermission(identity Identity, permission string) bool
	AuthenticateExplicit(r *http.Request) (*Identity, *HTTPError)
}

// Store abstracts persistence for Nightscout collections.
//
// Applications are expected to provide their own implementation.
type Store interface {
	Create(ctx context.Context, collection Collection, data map[string]any, subject string) (Record, *HTTPError)
	Get(ctx context.Context, collection Collection, identifier string) (Record, *HTTPError)
	Search(ctx context.Context, collection Collection, query *Query) ([]Record, *HTTPError)
	Update(ctx context.Context, collection Collection, identifier string, patch map[string]any, subject string) (Record, *HTTPError)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool, subject string) *HTTPError
	DeleteMatching(ctx context.Context, collection Collection, query *Query, permanent bool, subject string) (int, *HTTPError)
}
