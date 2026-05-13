package user

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/assets/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetSyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		acceptType := httputils.GetAcceptType(r)
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())

		if acceptType == httputils.AcceptTypeJSON {
			// TODO return JSON of current settings, devices, tokens, last seen, etc
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "This will return the current sync settings in JSON format."}`))
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			templates.UserSyncSettingsHTML(w, subdomain)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			templates.UserSyncSettingsHTMX(w)
			return

		}

	}
}
