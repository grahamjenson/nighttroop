package handlers

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

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
