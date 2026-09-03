package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
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
		query, err := nsutils.ParseQuery(r.URL.Query())
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

		q, err := nsutils.ParseQuery(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		q.Fields = []string{chi.URLParam(r, "field")}
		typeEq := chi.URLParam(r, "type")
		if typeEq != "" {
			q.TypeEq = ptr(typeEq)
		}

		records, err := dep.Store.Search(r.Context(), collection, q)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, q.Fields)
	}
}

func Times(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO this just lists everything at this point
		collection, _, err := nsutils.GetCollectionFromStringAndEnsurePermission(dep, "entries", r, nsutils.CollectionActionRead)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		prefix := chi.URLParam(r, "prefix")
		q, err := nsutils.ParseQuery(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		records, err := dep.Store.Search(r.Context(), collection, q)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		filteredRecords := []nsutils.Record{}
		for _, record := range records {
			if strings.HasPrefix(record.OccuredAtString(), prefix) {
				filteredRecords = append(filteredRecords, record)
			}
		}
		nsutils.WriteRecordsV1(w, filteredRecords, q.Fields)
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
		q, err := nsutils.ParseQuery(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		qJSON, _ := json.MarshalIndent(q, "", "  ")

		identifierOrType := chi.URLParam(r, "identifierOrType")
		if identifierOrType != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifierOrType) {
				q.TypeEq = ptr(identifierOrType)
			} else {
				q.IDEq = ptr(identifierOrType)
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
