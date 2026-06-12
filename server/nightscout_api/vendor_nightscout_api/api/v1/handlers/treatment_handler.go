package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// TreatmentsList lists treatments using the Nightscout v1 query format.
func TreatmentsList(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), "created_at")
		records, err := dep.Store.Search(r.Context(), "treatments", query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// TreatmentsCreate creates one or more treatments.
func TreatmentsCreate(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		var body any
		if err := nsutils.ReadJSON(r, &body); err != nil {
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		switch typed := body.(type) {
		case map[string]any:
			doc := typed
			doc, err := nsutils.Treatments.DefaultNormalizeValidateCreateDoc(doc)
			if err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
			if _, err := dep.Store.Create(r.Context(), "treatments", doc, identity.Name); err != nil {
				nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
				return
			}
		case []any:
			for _, item := range typed {
				doc, ok := item.(map[string]any)
				if !ok {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
					return
				}
				doc, err := nsutils.Treatments.DefaultNormalizeValidateCreateDoc(doc)
				if err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
				if _, err := dep.Store.Create(r.Context(), "treatments", doc, identity.Name); err != nil {
					nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest, "message": err.Error()})
					return
				}
			}
		default:
			nsutils.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": http.StatusBadRequest})
			return
		}
		records, _ := dep.Store.Search(r.Context(), "treatments", nsutils.DefaultQuery())
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// TreatmentsDelete deletes treatments matching the supplied v1 query.
func TreatmentsDelete(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), "created_at")
		deleted, err := dep.Store.DeleteMatching(r.Context(), "treatments", query, true, identity.Name)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"status": http.StatusInternalServerError, "message": err.Error()})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	}
}
