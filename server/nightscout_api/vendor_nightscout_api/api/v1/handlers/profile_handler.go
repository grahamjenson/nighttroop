package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// ProfileList returns recent profile documents ordered by creation time.
func ProfileList(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		records, err := dep.Store.Search(r.Context(), "profile", &nsutils.Query{Limit: 100, SortField: "created_at", SortDesc: true})
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}

// ProfileCurrent returns only the most recent profile document.
func ProfileCurrent(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		records, err := dep.Store.Search(r.Context(), "profile", &nsutils.Query{Limit: 1, SortField: "created_at", SortDesc: true})
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if len(records) == 0 {
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, records[0])
	}
}
