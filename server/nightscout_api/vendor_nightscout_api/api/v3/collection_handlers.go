package v3

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Read returns a single document from an API v3 collection.
func Read(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, _, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionRead)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		record, err := dep.Store.Get(r.Context(), collection, chi.URLParam(r, "identifier"))

		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		if modifiedSince, ok := nsutils.ParseIfModifiedSince(r); ok && time.UnixMilli(record.SrvModified).Before(modifiedSince.Add(time.Second)) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		query, _ := nsutils.ParseV3(r.URL.Query())

		nsutils.WriteRecordV3(w, record, query.Fields)

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

		query, err := nsutils.ParseV3(r.URL.Query())
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

		// The Actual Create
		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": "Bad or missing request body"})
			return
		}

		doc, err = collection.DefaultNormalizeValidateCreateDoc(doc)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		record, err := dep.Store.Create(r.Context(), collection, doc)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v3/%s/%s", collection, record.Identifier()))
		nsutils.WriteRecordV3(w, record, nil)

	}
}

// Update replaces a document in an API v3 collection and supports upsert semantics when allowed by the caller's permissions.
func Update(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, identity, err := nsutils.GetCollectionAndEnsurePermission(dep, r, nsutils.CollectionActionUpdate)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		canCreate := dep.Auth.HasPermission(*identity, fmt.Sprintf("api:%s:create", collection))
		identifier := chi.URLParam(r, "identifier")

		//
		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}

		// TODO cleanup folling logic make it understandable
		existing, err := dep.Store.Get(r.Context(), collection, identifier)
		existingNotFound := err == nsutils.ErrNotFound
		if err == nsutils.ErrGone {
			// Deleted Soft
			nsutils.WriteJSON(w, http.StatusGone, map[string]any{"status": http.StatusGone})
			return
		}
		if err != nil && !existingNotFound {
			nsutils.WriteError(w, err)
			return
		}

		// Cant find existing record, and user cant create a new one that is an error
		if existingNotFound && !canCreate {
			nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission api:%s:create", collection)})
			return
		}

		createNewRecord := existingNotFound && canCreate
		updateRecord := !existingNotFound

		if updateRecord {
			// We just check if the user has sent a header to not update if unmodified since
			if unmodifiedSince, ok := nsutils.ParseIfUnmodifiedSince(r); ok && time.UnixMilli(existing.SrvModified).After(unmodifiedSince) {
				nsutils.WriteJSON(w, http.StatusPreconditionFailed, map[string]any{"status": http.StatusPreconditionFailed})
				return
			}

		}

		// Set default values, normalize incoming values, and validate everything is good, then Create or Update

		var record nsutils.Record
		if createNewRecord {
			doc, err = collection.DefaultNormalizeValidateCreateDoc(doc)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			record, err = dep.Store.Create(r.Context(), collection, doc)
		} else {
			doc, err = collection.DefaultNormalizeValidateUpdateDoc(existing, doc)
			if err != nil {
				nsutils.WriteError(w, err)
				return
			}
			record, err = dep.Store.Update(r.Context(), collection, identifier, doc)
		}

		if err != nil || record.Identifier() == "" {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, nil)

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

		var doc map[string]any
		if err := nsutils.ReadJSON(r, &doc); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		existing, err := dep.Store.Get(r.Context(), collection, identifier)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		if unmodifiedSince, ok := nsutils.ParseIfUnmodifiedSince(r); ok && time.UnixMilli(existing.SrvModified).After(unmodifiedSince) {
			nsutils.WriteJSON(w, http.StatusPreconditionFailed, map[string]any{"status": http.StatusPreconditionFailed})
			return
		}
		doc, err = collection.DefaultNormalizeValidateUpdateDoc(existing, doc)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		record, err := dep.Store.Update(r.Context(), collection, identifier, doc)

		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordV3(w, record, nil)

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

		err = dep.Store.Delete(r.Context(), collection, chi.URLParam(r, "identifier"), permanent)
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

		query, err := nsutils.ParseV3(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		since, errr := strconv.ParseInt(chi.URLParam(r, "since"), 10, 64)
		if errr != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}

		query.Filters = append(query.Filters, nsutils.Filter{Field: "srvModified", Op: "gte", Value: strconv.FormatInt(since, 10)})

		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}

		nsutils.WriteRecordsV3(w, records, query.Fields)
	}
}
