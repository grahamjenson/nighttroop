package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func ptr[T any](v T) *T {
	return &v
}

func Count(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		query, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
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

		q, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		q.Fields = &nsutils.QueryStringArray{chi.URLParam(r, "field")}
		typeEq := chi.URLParam(r, "type")
		if typeEq != "" {
			q.TypeEq = ptr(nightscout.ShortString(typeEq))
		}

		records, err := dep.Store.Search(r.Context(), collection, q)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, q.Fields.ToArray())
	}
}

func Times(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Unsure of what this endpoint actually does
		collection, _, err := nsutils.GetCollectionFromStringAndEnsurePermission(dep, "entries", r, nsutils.CollectionActionRead)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		q, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		records, err := dep.Store.Search(r.Context(), collection, q)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, q.Fields.ToArray())
	}

}

func Echo(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			collection, _, err = nsutils.GetCollectionFromStringAndEnsurePermission(dep, "entries", r, nsutils.CollectionActionRead)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		q, err := nsutils.ParseV1(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		qJSON, _ := json.MarshalIndent(q, "", "  ")

		identifierOrType := chi.URLParam(r, "identifierOrType")
		if identifierOrType != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifierOrType) {
				q.TypeEq = ptr(nightscout.ShortString(identifierOrType))
			} else {
				q.IDEq = ptr(nightscout.ShortString(identifierOrType))
			}
		}

		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"req": map[string]any{
				"method": r.Method,
				"url":    r.URL.String(),
				"query":  r.URL.Query(),
				"header": r.Header,
			},
			"collection": collection,
			"query":      string(qJSON),
		})
	}
}
