package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

func Search(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collection, err := nsutils.ParseCollection(chi.URLParam(r, "collection"))
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collection"})
			return
		}
		permission := fmt.Sprintf("api:%s:read", collection)
		dep.Auth.Require(permission, false, func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
			query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())

			records, err := dep.Store.Search(r.Context(), collection, query)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
				return
			}
			nsutils.WriteRecordsV1(w, records, query.Fields)
		}).ServeHTTP(w, r)
	}
}

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

		permission := fmt.Sprintf("api:%s:create", collection)

		if !dep.Auth.HasPermission(*identity, permission) {
			nsutils.WriteJSON(w, http.StatusForbidden, map[string]any{"status": http.StatusForbidden, "message": fmt.Sprintf("Missing permission %s", permission)})
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

		record, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v3/%s/%s", collection, record.Identifier()))
		nsutils.WriteRecordsV1(w, []nsutils.Record{record}, nil)

	}
}

// CollectionCreate creates one or more records in a generic collection.
func CollectionCreate(dep nsutils.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		var created []nsutils.Record
		switch typed := body.(type) {
		case map[string]any:
			doc := typed
			doc, err := collection.DefaultNormalizeValidateCreateDoc(doc)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			record, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			created = append(created, record)
		case []any:
			for _, item := range typed {
				doc, ok := item.(map[string]any)
				if !ok {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
					return
				}
				doc, err := collection.DefaultNormalizeValidateCreateDoc(doc)
				if err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
				record, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
				if err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
				created = append(created, record)
			}
		default:
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		nsutils.WriteRecordsV1(w, created, nil)
	}
}

// CollectionDelete deletes records from a generic collection using the
// v1 query syntax.
func CollectionDelete(dep nsutils.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())
		deleted, err := dep.Store.DeleteMatching(r.Context(), collection, query, true, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
