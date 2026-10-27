package nsv1

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// Versions reports the API versions exposed by this server.
func Versions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiutils.WriteJSON(w, http.StatusOK, []map[string]string{
			{"url": "/api/v1", "version": "1.0.0"},
			{"url": "/api/v2", "version": "2.0.0"},
			{"url": "/api/v3", "version": "3.0.4"},
		})
	}
}

// VerifyAuth implements the v1 verifyauth endpoint.
func VerifyAuth(dep nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := nsutils.AuthenticateExplicit(r)
		if err != nil {
			apiutils.WriteJSON(w, http.StatusOK, map[string]any{
				"status": "ok",
				"message": map[string]any{
					"message": "UNAUTHORIZED",
					"isAdmin": false,
				},
			})
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"message": map[string]any{
				"message": "OK",
				"isAdmin": nsutils.HasPermission(*identity, "api:*:admin") || nsutils.HasPermission(*identity, "api:entries:create"),
			},
		})
	}
}

// StatusJSON serves the v1 JSON status payload.
func StatusJSON(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiutils.WriteJSON(w, http.StatusOK, Status())
	}
}

// Status is the v1 status payload, which a websocket authorize asking for status gets too
func Status() map[string]any {
	return map[string]any{
		"status":            "ok",
		"apiEnabled":        true,
		"careportalEnabled": true,
		"settings": map[string]any{
			"enable": false,
		},
	}
}
