package nsutils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/types"
)

// Identity is the resolved authorization context for a request.
type Identity struct {
	Name        string
	Permissions []string
}

// Store abstracts persistence for Nightscout collections.
type Store interface {
	// Create is obvious it will create a new record in the collection and return it, or an error if it fails.
	// If duplicate it will return the existing record and a boolean indicating it was a duplicate
	Create(ctx context.Context, collection Collection, data map[string]any, raw bool) (Document, bool, *HTTPError)

	// Get will retrieve a record from the collection by its identifier, or return an error if it fails.
	Get(ctx context.Context, collection Collection, identifier string, query nsquery.Query) (Document, *HTTPError)

	// Search gets a query that will return a list of records that satifsy the query
	Search(ctx context.Context, collection Collection, query nsquery.Query) ([]Document, *HTTPError)

	// Delete deletes a single record (permanent is soft delete or hard delete)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool) *HTTPError

	// DeleteType deletes all records of a given type (type can mean different things to different collections)
	DeleteType(ctx context.Context, collection Collection, typeString string, permanent bool) (int, *HTTPError)

	// DeleteQuery soft deletes what a search with the query finds, at most a search's worth, and says how many
	DeleteQuery(ctx context.Context, collection Collection, query nsquery.Query) (int, *HTTPError)

	// Replace
	// Cant change "app" OR "event_type" fields , need to be passed and you need to check they are the same
	// need created_at still, as well
	// Delete old record and insert new one with the same identifier
	// return    v1 return the recrod as body v3 "body": { "lastModified": 1785972708836, "status": 200 }, on success
	// v1 if id is not found it is just created, no checks or anything
	Replace(ctx context.Context, collection Collection, identifier string, data map[string]any, canCreate bool, raw bool) (Document, bool, *HTTPError)

	// patch PATCH
	// Patch will 404 if no record is founf of that id
	Patch(ctx context.Context, collection Collection, identifier string, patch map[string]any, raw bool) (Document, *HTTPError)
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

// deleteFilters are the query's keys a delete takes, those xDrip4iOS sends to delete readings
var deleteFilters = map[string]bool{"token": true, "find[type]": true, "find[date][$gte]": true, "find[date][$lte]": true, "find[date][$in][]": true}

// ParseDeleteQuery is the query of a delete, which deletes what a search with it finds. Any other key is refused, as is
// a query without a date, so a filter not understood is never ignored to delete more than was asked.
func ParseDeleteQuery(values url.Values) (nsquery.Query, *HTTPError) {
	for key := range values {
		if !deleteFilters[key] {
			return nsquery.Query{}, ErrBadQueryError(key + " cannot be used to delete")
		}
	}
	query, err := ParseQuery(values)
	if err == nil && query.DateGTE == nil && query.DateLTE == nil && query.DateIn == nil {
		err = ErrBadQueryError("a delete needs find[date][$gte], find[date][$lte] or find[date][$in][]")
	}
	return query, err
}

// ReadCollectionAndQuery is the request's collection, which its identity must be able to read, and its query.
// When either is bad it writes the error and ok is false.
func ReadCollectionAndQuery(w http.ResponseWriter, r *http.Request) (collection Collection, query nsquery.Query, ok bool) {
	collection, _, err := GetCollectionAndEnsurePermission(r, CollectionActionRead)
	if err == nil {
		if query, err = ParseQuery(r.URL.Query()); err == nil {
			return collection, query, true
		}
	}
	WriteError(w, err)
	return collection, query, false
}

// ReadCollectionAndIdentifier is the request's collection, which its identity must be able to act on, the identity,
// and the identifier in its URL. When either is bad it writes the error and ok is false.
func ReadCollectionAndIdentifier(w http.ResponseWriter, r *http.Request, action CollectionAction) (collection Collection, identity *Identity, identifier string, ok bool) {
	collection, identity, err := GetCollectionAndEnsurePermission(r, action)
	if err == nil {
		identifier = chi.URLParam(r, "identifier")
		if err = ValidateIdentifierIsntBad(identifier); err == nil {
			return collection, identity, identifier, true
		}
	}
	WriteError(w, err)
	return collection, identity, identifier, false
}

// RawParam is the request's raw parameter, false when it is missing or not a bool
func RawParam(r *http.Request) bool {
	var raw types.FlexibleBool
	if json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw) != nil {
		return false
	}
	return raw.Bool()
}
