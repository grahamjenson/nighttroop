package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/types"
)

func Search(dep nsutils.Dependencies) http.HandlerFunc {
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
		identifierOrType := chi.URLParam(r, "identifier_or_type")
		if identifierOrType != "" {
			// If the URL has an identifier or type segment, treat it as a filter on the default id or type field.
			if collection.IsTypeString(identifierOrType) {
				query.TypeEq = ptr(identifierOrType)
			} else {
				query.IDEq = ptr(identifierOrType)
			}
		}
		records := []nsutils.Record{}
		if query.IDEq != nil {
			record, err := dep.Store.Get(r.Context(), collection, *query.IDEq, query)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			records = append(records, record)
		} else {
			records, err = dep.Store.Search(r.Context(), collection, query)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
		}

		nsutils.WriteRecordsV1(w, records, query.Fields)
	}
}

func Create(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		var created []nsutils.Record

		// Deal with one record being sent or an array being sent
		arrayOfModels := []any{}
		switch typed := body.(type) {
		case map[string]any:
			arrayOfModels = append(arrayOfModels, typed)
		case []any:
			arrayOfModels = typed
		default:
			slog.Error("Collection Create JSON type error", "type", typed)
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		var raw types.FlexibleBool
		rawError := json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw)
		if rawError != nil {
			raw = types.FlexibleBool(false)
		}

		// Loop over the array of models sent
		for _, item := range arrayOfModels {
			doc, ok := item.(map[string]any)

			if !ok {
				nsutils.WriteError(w, nsutils.BadDataBodyError)
				return
			}

			record, _, err := dep.Store.Create(r.Context(), collection, doc, raw.Bool())
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			created = append(created, record)
		}

		nsutils.WriteRecordsV1(w, created, nil)
	}
}

func Current(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
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
		// Force limit 1 and sort decending by date
		query.Limit = 1
		query.SortOrder = nil

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteRecordsV1(w, records, query.Fields)
	}

}

func Delete(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionDelete)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		permanent := r.URL.Query().Get("permanent") == "true"

		identifierOrType := chi.URLParam(r, "identifier_or_type")
		if identifierOrType != "" {
			if collection.IsTypeString(identifierOrType) {
				amount, err := dep.Store.DeleteType(r.Context(), collection, identifierOrType, permanent)
				if err != nil {
					// IGNORE THE ERROR AND JUST RETURN 0 DELETED COUNT
					nsutils.WriteJSON(w, http.StatusOK, map[string]any{
						"acknowledged": true,
						"deletedCount": 0,
					})
					return
				}
				nsutils.WriteJSON(w, http.StatusOK, map[string]any{
					"acknowledged": true,
					"deletedCount": amount,
				})
				return
			} else {
				err = dep.Store.Delete(r.Context(), collection, identifierOrType, permanent)
				if err != nil {
					// IGNORE THE ERROR AND JUST RETURN 0 DELETED COUNT
					nsutils.WriteJSON(w, http.StatusOK, map[string]any{
						"acknowledged": true,
						"deletedCount": 0,
					})
					return
				}

				nsutils.WriteJSON(w, http.StatusOK, map[string]any{
					"acknowledged": true,
					"deletedCount": 1,
				})
				return
			}
		}

	}
}

// Replace If not found will return 404
func Replace(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, identity, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		canCreate := nsutils.HasPermission(dep, collection, identity, nsutils.CollectionActionCreate)

		var body map[string]any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteError(w, nsutils.BadDataBodyError)
			return
		}

		// GET THE IDENTIFIER FROM THE URL OR BODY
		identifier := chi.URLParam(r, "identifier")
		if identifier == "" {
			if uid, ok := body["_id"].(string); ok {
				identifier = uid
			}
			if identifier == "" {
				if id, ok := body["identifier"].(string); ok {
					identifier = id
				}
			}
		}

		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var raw types.FlexibleBool
		rawError := json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw)
		if rawError != nil {
			raw = types.FlexibleBool(false)
		}

		updated, _, herr := dep.Store.Replace(r.Context(), collection, identifier, body, canCreate, raw.Bool())
		if herr != nil {
			nsutils.WriteError(w, herr)
			return
		}
		nsutils.WriteRecordV1(w, updated, nil)
	}
}
