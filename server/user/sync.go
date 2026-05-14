package user

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/user"
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
		userdb, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())
		ctx := r.Context()

		if acceptType == httputils.AcceptTypeJSON {
			// TODO return JSON of current settings, devices, tokens, last seen, etc
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "This will return the current sync settings in JSON format."}`))
			return
		}

		nightscout, dexcom, err := userdb.GetAllSyncSources(ctx)

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var ns_id *int64 = nil
		ns_url := ""
		ns_password_garbage := ""
		ns_enabled := false
		if nightscout != nil {
			if nightscout.Url != nil {
				ns_url = *nightscout.Url
			}
			ns_password_garbage = ""
			ns_enabled = nightscout.Enabled
			ns_id = &nightscout.ID
		}

		dex_username := ""
		dex_password_garbage := ""
		dex_usa_region := false
		dex_enabled := false
		var dex_id *int64 = nil
		if dexcom != nil {
			if dexcom.Username != nil {
				dex_username = *dexcom.Username
			}
			dex_password_garbage = ""
			dex_usa_region = dexcom.DexcomRegion == consts.DexcomRegionUS
			dex_enabled = dexcom.Enabled
			dex_id = &dexcom.ID
		}

		if acceptType == httputils.AcceptTypeHTML {
			err = templates.UserSettingsSyncHTML(w, subdomain, ns_url, ns_password_garbage, ns_enabled, ns_id, dex_username, dex_password_garbage, dex_usa_region, dex_enabled, dex_id)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTML template", "subdomain", subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = templates.UserSettingsSyncHTMX(w, ns_url, ns_password_garbage, ns_enabled, ns_id, dex_username, dex_password_garbage, dex_usa_region, dex_enabled, dex_id)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTMX template", "subdomain", subdomain)
				return
			}
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

func POSTSyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		userdb, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())
		ctx := r.Context()

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
		dexRegionStr := r.FormValue("dex_region") // "US" or "INT"
		dexRegion := consts.NewDexcomRegionFromString(dexRegionStr)

		saveNS, saveDex, err := validateSyncSettingInputs(nsUrl, nsPassword, dexUsername, dexPassword)

		if err != nil {
			slog.Error("Validation error for sync settings inputs", "error", err)
			templates.UserFormResponsesFailureHTMX(w, err.Error())
			return
		}

		tx, err := userdb.SQLiteDB.BeginTx(ctx, nil)
		if err != nil {
			slog.Error("Failed to begin transaction for sync source updates", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		userdbWithTx := userdb.Queries.WithTx(tx)

		if saveNS {
			err = userdbWithTx.UpsertSyncSource(ctx, user.UpsertSyncSourceParams{
				Type:         consts.SyncSourceTypeNightscout,
				Password:     nsPassword,
				Enabled:      nsEnabled,
				Url:          &nsUrl,
				Username:     nil,
				DexcomRegion: consts.DexcomRegionOther,
			})
			if err != nil {
				slog.Error("Failed to upsert Nightscout sync source", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		if saveDex {
			err = userdbWithTx.UpsertSyncSource(ctx, user.UpsertSyncSourceParams{
				Type:         consts.SyncSourceTypeDexcom,
				Password:     dexPassword,
				Enabled:      dexEnabled,
				Url:          nil,
				Username:     &dexUsername,
				DexcomRegion: dexRegion,
			})
			if err != nil {
				slog.Error("Failed to upsert Dexcom sync source", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()

		if err != nil {
			slog.Error("Failed to commit sync source updates", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After saving, redirect back to the sync settings page
			http.Redirect(w, r, "/sync-settings", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			templates.UserFormResponsesSuccessHTMX(w, "Sync settings saved successfully!")
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Sync settings saved successfully!"}`))
	}
}

func DELETESyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		acceptType := httputils.GetAcceptType(r)
		userdb, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())

		// 1. Extract ID from URL path
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			http.Error(w, "Bad Request: Missing ID", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Bad Request: Invalid ID", http.StatusBadRequest)
			return
		}

		// 2. TODO: Execute your database deletion logic here using the extracted ID
		err = userdb.Queries.DeleteSyncSource(r.Context(), id)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After deletion, redirect back to the sync settings page
			http.Redirect(w, r, "/sync-settings", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			// Return a success message that HTMX can use to trigger a toast notification
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("HX-Redirect", "/sync-settings")
			return
		}

		// Defaule to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

func validateSyncSettingInputs(nsUrl string, nsPassword string, dexUsername string, dexPassword string) (bool, bool, error) {
	saveNS := false
	saveDex := false
	errMsg := ""

	if nsUrl != "" {
		saveNS = true
	}

	if dexUsername != "" {
		saveDex = true
	}

	// Validations
	if saveNS {
		// URL cannot have a path or query
		nsUrlURl, err := url.Parse(nsUrl)
		if err != nil {
			errMsg += "Nightscout URL is not a valid URL. "
		} else {
			if nsUrlURl.Path != "" && nsUrlURl.Path != "/" {
				errMsg += "Nightscout URL cannot have a path. "
			}
			if nsUrlURl.RawQuery != "" {
				errMsg += "Nightscout URL cannot have query parameters. "
			}
		}
		if nsPassword == "" {
			errMsg += "Nightscout password cannot be empty. "

		}
	}

	if saveDex {
		if dexUsername == "" {
			errMsg += "Dexcom username cannot be empty. "
		}

		if dexPassword == "" {
			errMsg += "Dexcom password cannot be empty. "
		}
	}

	if errMsg != "" {
		return saveNS, saveDex, fmt.Errorf(errMsg)
	}
	return saveNS, saveDex, nil
}
