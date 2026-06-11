package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// StatusJSON serves the v1 JSON status payload.
func StatusJSON(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"apiEnabled":        true,
			"careportalEnabled": true,
			"settings": map[string]any{
				"enable": false,
			},
		})
	}
}
