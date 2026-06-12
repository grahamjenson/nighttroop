package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Versions reports the API versions exposed by this server.
func Versions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nsutils.WriteJSON(w, http.StatusOK, []map[string]string{
			{"url": "/api/v1", "version": "1.0.0"},
			{"url": "/api/v2", "version": "2.0.0"},
			{"url": "/api/v3", "version": "3.0.4"},
		})
	}
}

// EchoRoute returns the parsed v1 query for debugging client requests.
func EchoRoute(_ di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		values := r.URL.Query()
		query := nsutils.ParseV1(values, "date")
		collection := chi.URLParam(r, "collection")

		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"storage": collection,
			"query":   query,
			"input": map[string]any{
				"find": values,
			},
		})
	}
}

// SliceRoute implements the v1 prefix-slice endpoint.
func SliceRoute(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		q := nsutils.ParseV1(r.URL.Query(), "date")
		q.StripImplicitDateFilters()
		searchQuery := q
		searchQuery.Limit = 1000
		records, err := dep.Store.Search(r.Context(), collection, searchQuery)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		filtered := nsutils.Slice(records, chi.URLParam(r, "field"), chi.URLParam(r, "type"), chi.URLParam(r, "prefix"), q.Limit)
		nsutils.WriteRecordsV1(w, filtered, nil)
	}
}

// TimesEchoRoute returns the expanded modal-time patterns used by TimesRoute.
func TimesEchoRoute(_ di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
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

// TimesRoute implements the v1 modal-time query endpoint.
func TimesRoute(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
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
