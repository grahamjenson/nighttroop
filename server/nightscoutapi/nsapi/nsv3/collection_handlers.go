package nsv3

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// Read returns a single document from an API v3 collection.
func Read(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, identifier, ok := nsutils.ReadCollectionAndIdentifier(w, r, nsutils.CollectionActionRead)
		if !ok {
			return
		}

		query, _ := nsutils.ParseQuery(r.URL.Query())

		record, err := store.Get(r.Context(), collection, identifier, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, query.Fields, nsutils.Found)
	}
}

// Search implements collection search for API v3.
func Search(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, query, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}

		records, err := store.Search(r.Context(), collection, query)
		if err != nil {
			apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteRecordsV3(w, records, query.Fields)
	}
}

// Create creates a document in an API v3 collection.
func Create(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(r, nsutils.CollectionActionCreate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		// The Actual Create
		var doc map[string]any
		if err := apiutils.ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			apiutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": "Bad or missing request body"})
			return
		}

		record, duplicate, err := store.Create(r.Context(), collection, doc, nsutils.RawParam(r))
		if err != nil {
			apiutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
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
func Remove(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, identifier, ok := nsutils.ReadCollectionAndIdentifier(w, r, nsutils.CollectionActionDelete)
		if !ok {
			return
		}

		permanent := r.URL.Query().Get("permanent") == "true"
		err := store.Delete(r.Context(), collection, identifier, permanent)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK})

	}
}

// History returns document revisions modified since the supplied timestamp.
func History(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, query, ok := nsutils.ReadCollectionAndQuery(w, r)
		if !ok {
			return
		}

		since, errr := strconv.ParseInt(chi.URLParam(r, "since"), 10, 64)
		if errr != nil {
			apiutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}

		query.UpdatedGTE = new(since)
		// HasFilters was worked out when the query was parsed, so it must be set for a filter added after
		query.HasFilters = true

		records, err := store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordsV3(w, records, query.Fields)
	}
}

// Replace will replace (delete, then insert) if it exists, or create it if it doesnt
func Replace(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, identity, identifier, ok := nsutils.ReadCollectionAndIdentifier(w, r, nsutils.CollectionActionUpdate)
		if !ok {
			return
		}
		canCreate := nsutils.CollectionActionHasPermission(collection, identity, nsutils.CollectionActionCreate)

		doc, ok := readDocument(w, r)
		if !ok {
			return
		}
		// TODO put If-Unmodified-Since into the update query to make it atomic
		record, created, err := store.Replace(r.Context(), collection, identifier, doc, canCreate, nsutils.RawParam(r))

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
func Patch(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, identifier, ok := nsutils.ReadCollectionAndIdentifier(w, r, nsutils.CollectionActionUpdate)
		if !ok {
			return
		}

		doc, ok := readDocument(w, r)
		if !ok {
			return
		}
		// TODO put If-Unmodified-Since into the update query to make it atomic

		record, err := store.Patch(r.Context(), collection, identifier, doc, nsutils.RawParam(r))

		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, nil, nsutils.Replaced)
	}
}

// readDocument is the request's JSON body. When it is bad it writes a 400 and ok is false.
func readDocument(w http.ResponseWriter, r *http.Request) (doc map[string]any, ok bool) {
	if err := apiutils.ReadJSON(r, &doc); err != nil {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
		return nil, false
	}
	return doc, true
}
