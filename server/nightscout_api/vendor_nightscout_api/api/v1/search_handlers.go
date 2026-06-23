package v1

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func Count(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"count": len(records)})
	}

}

func Slice(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		q := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())
		q.StripImplicitDateFilters()
		searchQuery := q
		searchQuery.Limit = 1000
		records, err := dep.Store.Search(r.Context(), collection, searchQuery)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		filtered := nsutils.Slice(records, chi.URLParam(r, "field"), chi.URLParam(r, "type"), chi.URLParam(r, "prefix"), q.Limit)
		nsutils.WriteRecordsV1(w, filtered, nil)
	}
}

func Times(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix, expr, ok := ParseTimesWildcard(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		q := nsutils.ParseV1(r.URL.Query(), "date")
		q.StripImplicitDateFilters()
		searchQuery := q
		searchQuery.Limit = 1000
		records, err := dep.Store.Search(r.Context(), "entries", searchQuery)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		filtered, _ := nsutils.Times(records, prefix, expr, q.Limit)
		nsutils.WriteRecordsV1(w, filtered, nil)
	}

}

func TimesEcho(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix, expr, ok := ParseTimesWildcard(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, patterns := nsutils.Times(nil, prefix, expr, 0)
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"req":     map[string]any{"query": r.URL.Query()},
			"pattern": patterns,
		})
	}
}

func ParseTimesWildcard(r *http.Request) (prefix, expr string, ok bool) {
	wild := chi.URLParam(r, "*")
	parts := strings.SplitN(wild, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	prefix = parts[0]
	expr = strings.TrimSuffix(parts[1], ".json")
	if decoded, err := url.PathUnescape(prefix); err == nil {
		prefix = decoded
	}
	if decoded, err := url.PathUnescape(expr); err == nil {
		expr = decoded
	}
	if prefix == "" || expr == "" {
		return "", "", false
	}
	return prefix, expr, true
}
