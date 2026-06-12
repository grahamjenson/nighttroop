package di

import (
	"context"
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
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
	AuthenticateExplicit(r *http.Request) (*Identity, error)
}

// Store abstracts persistence for Nightscout collections.
//
// Applications are expected to provide their own implementation.
type Store interface {
	Create(ctx context.Context, collection nsutils.Collection, data map[string]any, subject string) (nsutils.Record, error)
	Get(ctx context.Context, collection nsutils.Collection, identifier string) (nsutils.Record, error)
	Search(ctx context.Context, collection nsutils.Collection, query *nsutils.Query) ([]nsutils.Record, error)
	Update(ctx context.Context, collection nsutils.Collection, identifier string, patch map[string]any, subject string) (nsutils.Record, error)
	Delete(ctx context.Context, collection nsutils.Collection, identifier string, permanent bool, subject string) error
	DeleteMatching(ctx context.Context, collection nsutils.Collection, query *nsutils.Query, permanent bool, subject string) (int, error)
}
