package nsutils

import (
	"context"
	"net/http"
	"net/url"

	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
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

type AuthManager interface {
	Require(permission string, allowDefault bool, next func(http.ResponseWriter, *http.Request, *Identity)) http.HandlerFunc
	HasPermission(identity Identity, permission string) bool
	AuthenticateExplicit(r *http.Request) (*Identity, *HTTPError)
}

// Store abstracts persistence for Nightscout collections.
type Store interface {
	Create(ctx context.Context, collection Collection, data map[string]any) (Record, *HTTPError)
	Get(ctx context.Context, collection Collection, identifier string) (Record, *HTTPError)
	Search(ctx context.Context, collection Collection, query nsquery.Query) ([]Record, *HTTPError)
	Update(ctx context.Context, collection Collection, identifier string, patch map[string]any) (Record, *HTTPError)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool) *HTTPError
	DeleteType(ctx context.Context, collection Collection, typeString string, permanent bool) (int, *HTTPError)
}

// ParseQuery translates Nightscout v1 query parameters into the storage query
// model used by this library.
func ParseQuery(values url.Values) (nsquery.Query, *HTTPError) {
	qp, err := nsquery.ParseQueryValues(values)
	if err != nil {
		return nsquery.Query{}, ErrBadQueryError(err.Error())
	}

	q := nsquery.QueryValueToQuery(qp)

	return q, nil
}

func ptr[T any](v T) *T {
	return &v
}
