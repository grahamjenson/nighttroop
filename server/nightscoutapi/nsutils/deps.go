package nsutils

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
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
	Create(ctx context.Context, collection Collection, data map[string]any) (Document, bool, *HTTPError)

	// SplitV3 makes a document API v3 gives as more than one that many, before API v3 changes or deletes one of them
	SplitV3(ctx context.Context, collection Collection, identifier string) *HTTPError

	// Get will retrieve a record from the collection by its identifier, or return an error if it fails.
	Get(ctx context.Context, collection Collection, identifier string, query nsquery.Query) (Document, *HTTPError)

	// Search gets a query that will return a list of records that satifsy the query
	Search(ctx context.Context, collection Collection, query nsquery.Query) ([]Document, *HTTPError)

	// History is what changed since, deleted too, the oldest change first
	History(ctx context.Context, collection Collection, since int64, query nsquery.Query) ([]Document, *HTTPError)

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
	Replace(ctx context.Context, collection Collection, identifier string, data map[string]any, canCreate bool) (Document, bool, *HTTPError)

	// patch PATCH
	// Patch will 404 if no record is founf of that id
	Patch(ctx context.Context, collection Collection, identifier string, patch map[string]any) (Document, *HTTPError)
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

// deleteFilters are the query's keys a delete takes, and whether only a treatments delete takes it, as a reading's
// search does not filter on it: those xDrip4iOS sends to delete readings, Trio to delete readings, carbs, insulin and
// overrides, and iAPS to delete carbs, insulin, manual glucose, overrides and announcements
var deleteFilters = map[string]bool{"token": false, "count": false, "find[type]": false, "find[date][$gte]": false,
	"find[date][$lte]": false, "find[date][$in][]": false, "find[id][$eq]": false, "find[dateString][$eq]": false,
	"find[created_at][$eq]": false, "find[created_at][$gte]": false, "find[created_at][$lte]": false,
	"find[eventType][$eq]": false, "find[eventType]": false, "find[enteredBy]": true, "find[carbs][$exists]": true,
	"find[bolus][$exists]": true, "find[glucose][$exists]": true, "find[creation_date][$eq]": true}

const (
	// maxDeleteAlternatives is the most find[$or] alternatives a delete has, as each is a search
	maxDeleteAlternatives = 10
	// maxDeleteCount is the most a delete's count deletes, the newest first, as iAPS deletes its newest override
	maxDeleteCount = 10
	// maxDeleteRangeMs is the longest a delete's date range from $gte to $lte is, 2 days
	maxDeleteRangeMs = 2 * 24 * 60 * 60 * 1000
)

// ParseDeleteQueries is a delete's query for each of its find[$or] alternatives, or the one query without them, each
// as ParseDeleteQuery takes it
func ParseDeleteQueries(values url.Values, collection Collection) ([]nsquery.Query, *HTTPError) {
	alternatives := nsquery.OrBranches(values)
	if len(alternatives) > maxDeleteAlternatives {
		return nil, ErrBadQueryError("a delete has at most 10 find[$or] alternatives")
	}
	queries := make([]nsquery.Query, 0, len(alternatives))
	for _, alternative := range alternatives {
		query, err := ParseDeleteQuery(alternative, collection)
		if err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}
	return queries, nil
}

// ParseDeleteQuery is the query of a delete of the collection, which deletes what a search with it finds. Any other
// key is refused, so a filter not understood is never ignored to delete more than was asked. So is a query without a
// date or an id, unless it has a type and a count, a count of more than 10, and a date range of more than 2 days.
func ParseDeleteQuery(values url.Values, collection Collection) (nsquery.Query, *HTTPError) {
	for key := range values {
		if treatmentsOnly, ok := deleteFilters[key]; !ok || treatmentsOnly && collection != Treatments {
			return nsquery.Query{}, ErrBadQueryError(key + " cannot be used to delete")
		}
	}
	query, err := ParseQuery(values)
	dated := query.DateGTE != nil || query.DateLTE != nil || query.DateIn != nil || query.DateEQ != nil || query.IDEq != nil || query.CreationDateEQ != nil
	switch {
	case err != nil:
	case query.CountGiven && (query.CountAsked < 1 || query.CountAsked > maxDeleteCount):
		err = ErrBadQueryError("a delete's count is 1 to 10")
	case query.DateGTE != nil && query.DateLTE != nil && *query.DateLTE-*query.DateGTE > maxDeleteRangeMs:
		err = ErrBadQueryError("a delete's find[...][$gte] and find[...][$lte] are at most 2 days apart")
	case !dated && (!query.CountGiven || query.TypeEq == nil):
		err = ErrBadQueryError("a delete needs find[date][$gte], find[date][$lte], find[date][$in][], a date's $eq, find[id][$eq], or find[eventType] and a count")
	}
	return query, err
}

// ReadCollectionAndQuery is the request's collection, which its identity must be able to read, and its query.
// When either is bad it writes the error and ok is false.
func ReadCollectionAndQuery(w http.ResponseWriter, r *http.Request) (collection Collection, query nsquery.Query, ok bool) {
	collection, _, err := GetCollectionAndEnsurePermission(r, CollectionActionRead)
	if err == nil {
		if query, err = ParseQuery(r.URL.Query()); err == nil {
			// Readings are given up to a day of 1 minute readings, twice over, as xDrip+'s follower asks since its newest
			if collection == Entries {
				query.Limit = max(query.Limit, min(query.CountAsked, apiquery.MaxEntriesQueryLimit))
			}
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
		validate := ValidateIdentifierIsntBad
		if collection == Settings {
			validate = ValidateSettingKeyIsntBad
		}
		if err = validate(identifier); err == nil {
			return collection, identity, identifier, true
		}
	}
	WriteError(w, err)
	return collection, identity, identifier, false
}
