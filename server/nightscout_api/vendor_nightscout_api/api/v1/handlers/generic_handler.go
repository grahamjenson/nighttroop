package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// GenericCollectionList lists records for a collection that follows the common
// Nightscout list conventions.
func GenericCollectionList(dep di.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *di.Identity) {

	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())
		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// GenericCollectionCreate creates one or more records in a generic collection.
func GenericCollectionCreate(dep di.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
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

// GenericCollectionDelete deletes records from a generic collection using the
// v1 query syntax.
func GenericCollectionDelete(dep di.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), collection.DefaultDateField())
		deleted, err := dep.Store.DeleteMatching(r.Context(), collection, query, true, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
