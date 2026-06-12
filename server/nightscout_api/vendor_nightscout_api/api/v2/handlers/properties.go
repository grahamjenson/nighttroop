package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// PropertiesList lists properties using the Nightscout v1 query format.
func PropertiesList(dep di.Dependencies) func(http.ResponseWriter, *http.Request, *di.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *di.Identity) {
		query := nsutils.ParseV1(r.URL.Query(), "date")
		records, err := dep.Store.Search(r.Context(), "properties", query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}
