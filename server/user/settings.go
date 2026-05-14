package user

import (
	"log/slog"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
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

		userdb, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())
		nightscout, dexcom, err := userdb.GetAllSyncSources(r.Context())

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ns_url := ""
		ns_password_garbage := ""
		ns_enabled := false
		if nightscout != nil {
			if nightscout.Url != nil {
				ns_url = *nightscout.Url
			}
			ns_password_garbage = ""
			ns_enabled = nightscout.Enabled
		}

		dex_username := ""
		dex_password_garbage := ""
		dex_usa_region := false
		dex_enabled := false
		if dexcom != nil {
			if dexcom.Username != nil {
				dex_username = *dexcom.Username
			}
			dex_password_garbage = ""
			dex_usa_region = dexcom.DexcomRegion == consts.DexcomRegionUS
			dex_enabled = dexcom.Enabled
		}

		if acceptType == httputils.AcceptTypeHTML {
			err = templates.UserSettingsSyncHTML(w, subdomain, ns_url, ns_password_garbage, ns_enabled, dex_username, dex_password_garbage, dex_usa_region, dex_enabled)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTML template", "subdomain", subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = templates.UserSettingsSyncHTMX(w, ns_url, ns_password_garbage, ns_enabled, dex_username, dex_password_garbage, dex_usa_region, dex_enabled)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTMX template", "subdomain", subdomain)
				return
			}
			return
		}

		slog.Error("Unsupported Accept Type", "accept_type", acceptType)

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
