package v3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/types"
)

// Read returns a single document from an API v3 collection.
func Read(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		identifier := chi.URLParam(r, "identifier")
		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}

		query, _ := nsutils.ParseQuery(r.URL.Query())

		record, err := dep.Store.Get(r.Context(), collection, identifier, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, query.Fields, nsutils.Found)
	}
}

// Search implements collection search for API v3.
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

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteRecordsV3(w, records, query.Fields)
	}
}

// Create creates a document in an API v3 collection.
func Create(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var raw types.FlexibleBool
		rawError := json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw)
		if rawError != nil {
			raw = types.FlexibleBool(false)
		}

		// The Actual Create
		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": "Bad or missing request body"})
			return
		}

		record, duplicate, err := dep.Store.Create(r.Context(), collection, doc, raw.Bool())
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v3/%s/%s", collection, record.Identifier()))
		if duplicate {
			nsutils.WriteRecordV3(w, record, nil, nsutils.FoundDuplicate)
			return
		}
		nsutils.WriteRecordV3(w, record, nil, nsutils.Created)

	}
}

// Remove soft-deletes or permanently deletes an API v3 document.
func Remove(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionDelete)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		permanent := r.URL.Query().Get("permanent") == "true"
		identifier := chi.URLParam(r, "identifier")
		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}
		err = dep.Store.Delete(r.Context(), collection, identifier, permanent)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK})

	}
}

// History returns document revisions modified since the supplied timestamp.
func History(dep nsutils.Dependencies) http.HandlerFunc {
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

		since, errr := strconv.ParseInt(chi.URLParam(r, "since"), 10, 64)
		if errr != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}

		query.UpdatedGTE = ptr(since)

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordsV3(w, records, query.Fields)
	}
}

// Replace will replace (delete, then insert) if it exists, or create it if it doesnt
func Replace(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, identity, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionUpdate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		canCreate := nsutils.HasPermission(dep, collection, identity, nsutils.CollectionActionCreate)

		identifier := chi.URLParam(r, "identifier")
		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var raw types.FlexibleBool
		rawError := json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw)
		if rawError != nil {
			raw = types.FlexibleBool(false)
		}

		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		// TODO put If-Unmodified-Since into the update query to make it atomic
		record, created, err := dep.Store.Replace(r.Context(), collection, identifier, doc, canCreate, raw.Bool())

		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		if created {
			nsutils.WriteRecordV3(w, record, nil, nsutils.Created)
		} else {
			nsutils.WriteRecordV3(w, record, nil, nsutils.Replaced)
		}
	}
}

// Patch applies a partial update to an API v3 document.
func Patch(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionUpdate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		identifier := chi.URLParam(r, "identifier")
		if err := nsutils.ValidateIdentifierIsntBad(identifier); err != nil {
			nsutils.WriteError(w, err)
			return
		}

		var raw types.FlexibleBool
		rawError := json.Unmarshal([]byte(r.URL.Query().Get("raw")), &raw)
		if rawError != nil {
			raw = types.FlexibleBool(false)
		}

		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		// TODO put If-Unmodified-Since into the update query to make it atomic

		record, err := dep.Store.Patch(r.Context(), collection, identifier, doc, raw.Bool())

		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, nil, nsutils.Replaced)
	}
}
