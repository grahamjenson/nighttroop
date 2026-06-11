package handlers

import (
	"net/http"
	"time"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// ensureV1Timestamp fills in date/created_at when missing. Nightscout v1
// clients (Trio, xDrip, etc.) routinely POST treatments, devicestatus, and
// food without a timestamp and expect the server to default to now.
func ensureV1Timestamp(doc map[string]any) {
	if doc == nil {
		return
	}
	hasDate := false
	if _, ok := doc["date"]; ok && doc["date"] != nil {
		hasDate = true
	}
	if value, ok := nsutils.StringField(doc, "created_at"); ok && value != "" {
		hasDate = true
	}
	if hasDate {
		return
	}
	now := time.Now().UTC()
	doc["created_at"] = now.Format("2006-01-02T15:04:05.000Z")
	doc["date"] = now.UnixMilli()
}

// GenericCollectionList lists records for a collection that follows the common
// Nightscout list conventions.
func GenericCollectionList(dep nsutils.Dependencies, collection nsutils.Collection, defaultDateField string) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {

	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), defaultDateField)
		records, err := dep.Store.Search(r.Context(), collection, query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecords(w, records, nil)
	}
}

// GenericCollectionCreate creates one or more records in a generic collection.
func GenericCollectionCreate(dep nsutils.Dependencies, collection nsutils.Collection) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		var created []nsutils.Record
		switch typed := body.(type) {
		case map[string]any:
			ensureV1Timestamp(typed)
			record, _, err := dep.Store.Create(r.Context(), collection, typed, identity.Name)
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
				ensureV1Timestamp(doc)
				record, _, err := dep.Store.Create(r.Context(), collection, doc, identity.Name)
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
		nsutils.WriteRecords(w, created, nil)
	}
}

// GenericCollectionDelete deletes records from a generic collection using the
// v1 query syntax.
func GenericCollectionDelete(dep nsutils.Dependencies, collection nsutils.Collection, defaultDateField string) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), defaultDateField)
		deleted, err := dep.Store.DeleteMatching(r.Context(), collection, query, true, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
