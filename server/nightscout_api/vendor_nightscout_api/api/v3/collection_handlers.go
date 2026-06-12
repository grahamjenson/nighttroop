package v3

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Search implements collection search for API v3.
func Search(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			query, err := nsutils.ParseV3(r.URL.Query())
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}

			records, err := dep.Store.Search(r.Context(), collection, query)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
				return
			}
			nsutils.WriteRecordsV3(w, records, query.Fields)
		}).ServeHTTP(w, r)
	}
}

// Create creates a document in an API v3 collection.
func Create(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		identity, err := dep.Auth.AuthenticateExplicit(r)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusUnauthorized, map[string]any{"status": http.StatusUnauthorized, "message": "Missing or bad access token or JWT"})
			return
		}
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

		permission := fmt.Sprintf("api:%s:create", collection)

		if !dep.Auth.HasPermission(*identity, permission) {
			nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission %s", permission)})
			return
		}
		record, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v3/%s/%s", collection, record.Identifier()))
		nsutils.WriteRecordV3(w, record, nil)

	}
}

// Read returns a single document from an API v3 collection.
func Read(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}

		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			record, err := dep.Store.Get(r.Context(), collection, chi.URLParam(r, "identifier"))

			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			if modifiedSince, ok := nsutils.ParseIfModifiedSince(r); ok && time.UnixMilli(record.SrvModified).Before(modifiedSince.Add(time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			query, _ := nsutils.ParseV3(r.URL.Query())

			nsutils.WriteRecordV3(w, record, query.Fields)
		}).ServeHTTP(w, r)
	}
}

// Update replaces a document in an API v3 collection and supports upsert semantics when allowed by the caller's permissions.
func Update(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:update", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			canCreate := dep.Auth.HasPermission(*identity, fmt.Sprintf("api:%s:create", collection))
			var doc map[string]any
			if err := nsutils.ReadJSON(r, &doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			identifier := chi.URLParam(r, "identifier")
			// TODO cleanup folling logic make it understandable
			existing, err := dep.Store.Get(r.Context(), collection, identifier)
			existingNotFound := err == nsutils.ErrNotFound
			if err == nsutils.ErrGone {
				// Deleted Soft
				nsutils.WriteJSON(w, http.StatusGone, map[string]any{"status": http.StatusGone})
				return
			}
			if err != nil && !existingNotFound {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
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
				record, err = dep.Store.Create(r.Context(), collection, doc, identity.Name)
			} else {
				if err := validateReplaceDoc(existing, doc); err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
				record, err = dep.Store.Update(r.Context(), collection, identifier, doc, identity.Name)
			}

			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}

			nsutils.WriteRecordV3(w, record, nil)
		}).ServeHTTP(w, r)
	}
}

// Patch applies a partial update to an API v3 document.
func Patch(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}

		identifier := chi.URLParam(r, "identifier")

		permission := fmt.Sprintf("api:%s:update", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			var doc map[string]any
			if err := nsutils.ReadJSON(r, &doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			existing, err := dep.Store.Get(r.Context(), collection, identifier)
			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			if unmodifiedSince, ok := nsutils.ParseIfUnmodifiedSince(r); ok && time.UnixMilli(existing.SrvModified).After(unmodifiedSince) {
				nsutils.WriteJSON(w, http.StatusPreconditionFailed, map[string]any{"status": http.StatusPreconditionFailed})
				return
			}
			if err := validatePatchDoc(existing, doc); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			record, err := dep.Store.Update(r.Context(), collection, identifier, doc, identity.Name)

			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}

			nsutils.WriteRecordV3(w, record, nil)
		}).ServeHTTP(w, r)
	}
}

// Remove soft-deletes or permanently deletes an API v3 document.
func Remove(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:delete", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			permanent := r.URL.Query().Get("permanent") == "true"
			err := dep.Store.Delete(r.Context(), collection, chi.URLParam(r, "identifier"), permanent, identity.Name)
			if err != nil {
				status, body := nsutils.RequireRecord(err)
				nsutils.WriteJSON(w, status, body)
				return
			}
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK})
		}).ServeHTTP(w, r)
	}
}

// History returns document revisions modified since the supplied timestamp.
func History(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			query, err := nsutils.ParseV3(r.URL.Query())

			since, err := strconv.ParseInt(chi.URLParam(r, "since"), 10, 64)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
				return
			}
			query.Filters = append(query.Filters, nsutils.Filter{Field: "srvModified", Op: "gte", Value: strconv.FormatInt(since, 10)})

			records, err := dep.Store.Search(r.Context(), collection, query)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
				return
			}
			nsutils.WriteRecordsV3(w, records, query.Fields)
		}).ServeHTTP(w, r)
	}
}
