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
	// Create is obvious it will create a new record in the collection and return it, or an error if it fails.
	Create(ctx context.Context, collection Collection, data map[string]any) (Record, *HTTPError)

	// Get will retrieve a record from the collection by its identifier, or return an error if it fails.
	Get(ctx context.Context, collection Collection, identifier string) (Record, *HTTPError)

	// Search gets a query that will return a list of records that satifsy the query
	Search(ctx context.Context, collection Collection, query nsquery.Query) ([]Record, *HTTPError)

	// Delete deletes a single record (permanent is soft delete or hard delete)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool) *HTTPError

	// DeleteType deletes all records of a given type (type can mean different things to different collections)
	DeleteType(ctx context.Context, collection Collection, typeString string, permanent bool) (int, *HTTPError)

	// Replace
	// Cant change "app" OR "event_type" fields , need to be passed and you need to check they are the same
	// need created_at still, as well
	// Delete old record and insert new one with the same identifier
	// return    v1 return the recrod as body v3 "body": { "lastModified": 1785972708836, "status": 200 }, on success
	// v1 if id is not found it is just created, no checks or anything
	Replace(ctx context.Context, collection Collection, identifier string, data map[string]any) (Record, *HTTPError)

	// patch PATCH
	// Patch will 404 if no record is founf of that id
	Patch(ctx context.Context, collection Collection, identifier string, patch map[string]any) (Record, *HTTPError)
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
