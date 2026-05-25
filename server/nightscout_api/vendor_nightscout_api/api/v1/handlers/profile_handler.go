package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/auth"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/httpx"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/store"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/util"
)

// ProfileList returns recent profile documents ordered by creation time.
func ProfileList(dep deps.Dependencies) func(http.ResponseWriter, *http.Request, *auth.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *auth.Identity) {
		records, err := dep.Store.Search(r.Context(), "profile", store.Query{Limit: 100, SortField: "created_at", SortDesc: true})
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		util.WriteRecords(w, records, nil)
	}
}

// ProfileCurrent returns only the most recent profile document.
func ProfileCurrent(dep deps.Dependencies) func(http.ResponseWriter, *http.Request, *auth.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *auth.Identity) {
		records, err := dep.Store.Search(r.Context(), "profile", store.Query{Limit: 1, SortField: "created_at", SortDesc: true})
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if len(records) == 0 {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, records[0])
	}
}
