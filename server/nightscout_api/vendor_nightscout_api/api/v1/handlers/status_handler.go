package handlers

import (
	"net/http"
	"slices"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/deps"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/httpx"
)

// StatusJSON serves the v1 JSON status payload.
func StatusJSON(dep deps.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"apiEnabled":        true,
			"careportalEnabled": slices.Contains(dep.Config.Enable, "careportal"),
			"settings": map[string]any{
				"enable": dep.Config.Enable,
			},
		})
	}
}
