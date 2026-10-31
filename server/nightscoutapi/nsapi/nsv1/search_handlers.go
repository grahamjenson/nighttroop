package nsv1

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

func Count(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, query, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}
		records, err := store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"count": len(records)})
	}

}

func Slice(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// collection/field/type/prefix
		collection, q, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}

		typeEq := chi.URLParam(r, "type")
		if typeEq != "" {
			q.TypeEq = new(typeEq)
			// HasFilters was worked out when the query was parsed, so it must be set for a filter added after
			q.HasFilters = true
		}

		prefix := chi.URLParam(r, "prefix")
		field := chi.URLParam(r, "field")

		records, err := store.Search(r.Context(), collection, q)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		filteredRecords := []nsutils.Record{}
		for _, record := range records {
			switch field {
			// The only example I have seen is someone using this for dateString
			// Can add more if there are other usecases
			case "dateString":
				if strings.HasPrefix(record.OccuredAtString(), prefix) {
					filteredRecords = append(filteredRecords, record)
				}
			default:
				filteredRecords = append(filteredRecords, record)
			}

		}

		nsutils.WriteRecordsV1(w, filteredRecords, q.Fields)
	}
}

func Times(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// The entries are searched as any query asks, then those not at a time with the prefix are left out, so a
		// page of the search can hold fewer than its limit
		collection, _, err := nsutils.GetCollectionFromStringAndEnsurePermission("entries", r, nsutils.CollectionActionRead)
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
		records, err := store.Search(r.Context(), collection, q)
		if err != nil {
			apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
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

func Echo(store nsutils.Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(r, nsutils.CollectionActionRead)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		q, err := nsutils.ParseQuery(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		identifierOrType := chi.URLParam(r, "identifier_or_type")
		if identifierOrType != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifierOrType) {
				q.TypeEq = new(identifierOrType)
			} else {
				q.IDEq = new(identifierOrType)
			}
		}

		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"input": map[string]any{
				"method": r.Method,
				"url":    r.URL.String(),
				"query":  r.URL.Query(),
				"header": r.Header,
			},
			"collection": collection,
			"storage":    collection,
			"query":      q,
		})
	}
}
