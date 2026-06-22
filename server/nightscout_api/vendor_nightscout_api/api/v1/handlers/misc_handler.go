package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Versions reports the API versions exposed by this server.
func Versions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nsutils.WriteJSON(w, http.StatusOK, []map[string]string{
			{"url": "/api/v1", "version": "1.0.0"},
			{"url": "/api/v2", "version": "2.0.0"},
			{"url": "/api/v3", "version": "3.0.4"},
		})
	}
}

// VerifyAuth implements the v1 verifyauth endpoint.
func VerifyAuth(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := dep.Auth.AuthenticateExplicit(r)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusOK, map[string]any{
				"status": "ok",
				"message": map[string]any{
					"message": "UNAUTHORIZED",
					"isAdmin": false,
				},
			})
			return
		}
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"message": map[string]any{
				"message": "OK",
				"isAdmin": dep.Auth.HasPermission(*identity, "api:*:admin") || dep.Auth.HasPermission(*identity, "api:entries:create"),
			},
		})
	}
}

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
