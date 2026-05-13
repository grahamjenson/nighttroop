package user

import (
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/templates"
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

func POSTSyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse and validate form inputs
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 2. Extract form values
		nsEnabled := r.FormValue("ns_enabled") == "on"
		nsUrl := strings.TrimSpace(r.FormValue("ns_url"))
		nsPassword := r.FormValue("ns_password")

		dexEnabled := r.FormValue("dex_enabled") == "on"
		dexUsername := strings.TrimSpace(r.FormValue("dex_username"))
		dexPassword := r.FormValue("dex_password")
		dexRegion := r.FormValue("dex_region") // "US" or "INT"

		// 3. Strict Server-Side Validation for Nightscout URL (No sub-paths/trailing slashes)
		if nsEnabled {
			if nsUrl == "" || nsPassword == "" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				templates.UserToastsFailureHTMX(w)
				return
			}

			// TODO enforce valid URL that is not in the blocked list
		}

		if dexEnabled && (dexUsername == "" || dexPassword == "" || dexRegion == "") {
			w.WriteHeader(http.StatusUnprocessableEntity)
			templates.UserToastsFailureHTMX(w)

			return
		}

		// 4. TODO: Execute your database/encrypted storage persistence layer here
		// Example: err := controlDB.SaveSyncSettings(r.Context(), ...)

		// 5. Build combined response (In-line success message + OOB Toast wrapper)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		templates.UserToastsSuccessHTMX(w)
	}
}
