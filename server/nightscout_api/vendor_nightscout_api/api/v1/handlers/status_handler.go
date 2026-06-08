package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/di"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/httpx"
)

// StatusJSON serves the v1 JSON status payload.
func StatusJSON(dep di.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"apiEnabled":        true,
			"careportalEnabled": true,
			"settings": map[string]any{
				"enable": false,
			},
		})
	}
}
