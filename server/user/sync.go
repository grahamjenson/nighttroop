package user

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/server/components/shared_templ"
	"nighttroop.t1d.nz/server/components/user_templ"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func GetSyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		acceptType := httputils.GetAcceptType(r)
		userContext := httputils.GetUserContext(r.Context())

		ctx := r.Context()

		if acceptType == httputils.AcceptTypeJSON {
			// TODO return JSON of current settings, devices, tokens, last seen, etc
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "This will return the current sync settings in JSON format."}`))
			return
		}

		nightscout, dexcom, err := controlDB.GetAllUserSyncSources(ctx, userContext.Subdomain)

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var ns_id *int64 = nil
		ns_url := ""
		ns_password_garbage := ""
		ns_enabled := true
		if nightscout != nil {
			ns_url, err = config.Security.Decrypt(nightscout.UrlOrUsername)
			if err != nil {
				slog.Error("Failed to decrypt Nightscout URL from DB", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ns_enabled = nightscout.Enabled
			ns_id = &nightscout.ID
			ns_password_garbage = consts.PASSWORD_GARBAGE_VALUE
		}

		dex_username := ""
		dex_password_garbage := ""
		dex_usa_region := false
		dex_enabled := true
		var dex_id *int64 = nil
		if dexcom != nil {
			dex_username, err = config.Security.Decrypt(dexcom.UrlOrUsername)
			if err != nil {
				slog.Error("Failed to decrypt Dexcom username from DB", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			dex_usa_region = dexcom.DexcomRegion == consts.DexcomRegionUS
			dex_enabled = dexcom.Enabled
			dex_id = &dexcom.ID
			dex_password_garbage = consts.PASSWORD_GARBAGE_VALUE
		}

		if acceptType == httputils.AcceptTypeHTML {
			err = user_templ.SyncSettings(ns_url, ns_password_garbage, ns_enabled, ns_id, dex_username, dex_password_garbage, dex_usa_region, dex_enabled, dex_id).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTML template", "subdomain", userContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = user_templ.SyncSettingsForm(ns_url, ns_password_garbage, ns_enabled, ns_id, dex_username, dex_password_garbage, dex_usa_region, dex_enabled, dex_id).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTMX template", "subdomain", userContext.Subdomain)
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

func POSTSyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig, syncSourceType consts.SyncSourceType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowed, seconds := httputils.UserUpdateSyncAllowed(config, r); !allowed {
			http.Error(w, "Too Many Requests wait "+strconv.Itoa(seconds)+" seconds", http.StatusTooManyRequests)
			return
		}

		acceptType := httputils.GetAcceptType(r)
		userContext := httputils.GetUserContext(r.Context())

		ctx := r.Context()

		// 1. Parse and validate form inputs
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		nightscout, dexcom, err := controlDB.GetAllUserSyncSources(ctx, userContext.Subdomain)

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if syncSourceType == consts.SyncSourceTypeNightscout {
			err = saveNightScoutSettings(ctx, userContext.Subdomain, controlDB, config, nightscout, r.FormValue("ns_url"), r.FormValue("ns_password"), r.FormValue("ns_enabled") == "on")
		} else {
			err = saveDexcomSettings(ctx, userContext.Subdomain, controlDB, config.Security, dexcom, r.FormValue("dex_username"), r.FormValue("dex_password"), consts.NewDexcomRegionFromString(r.FormValue("dex_region")), r.FormValue("dex_enabled") == "on")
		}

		if err != nil {
			slog.Error("Validation error for sync settings inputs", "error", err)
			shared_templ.Fail().Render(r.Context(), w)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After saving, redirect back to the sync settings page
			http.Redirect(w, r, "/sync_settings", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			shared_templ.Success().Render(r.Context(), w)
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonResponce := []byte(`{"message": "Sync settings saved successfully!"}`)
		_, err = w.Write(jsonResponce)
		if err != nil {
			slog.Error("Failed to write JSON response for successful sync settings save", "error", err)
		}
	}
}

func DELETESyncSettingsHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		userContext := httputils.GetUserContext(r.Context())

		// 1. Extract ID from URL path
		typeStr := chi.URLParam(r, "type")
		if typeStr == "" {
			http.Error(w, "Bad Request: Missing type", http.StatusBadRequest)
			return
		}

		type_, err := consts.NewSyncSourceTypeFromString(strings.ToUpper(typeStr))
		if err != nil {
			http.Error(w, "Bad Request: Invalid type", http.StatusBadRequest)
			return
		}

		err = controlDB.Queries.DeleteSyncSource(r.Context(), control.DeleteSyncSourceParams{
			Subdomain: userContext.Subdomain,
			Type:      type_,
		})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After deletion, redirect back to the sync settings page
			http.Redirect(w, r, "/sync_settings", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			// Return a success message that HTMX can use to trigger a toast notification
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			w.Header().Set("HX-Redirect", "/sync_settings")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Defaule to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

///////
// UTIL METHODS
////////

func saveNightScoutSettings(ctx context.Context, subdomain string, controlDB *db.ControlDB, config config.ServerConfig, nightscout *control.UserSyncSource, nsUrl string, nsPassword string, enabled bool) error {

	if nsUrl == "" {
		return fmt.Errorf("Nightscout URL cannot be empty. ")
	}

	security := config.Security
	var err error
	errMsg := ""

	nsUrl, err = nsURLNormalizeValidate(nsUrl, config)
	if err != nil {
		errMsg += "Nightscout URL is not a valid URL. "
	}

	if nsPassword == "" {
		errMsg += "Nightscout password cannot be empty. "

	}

	if errMsg != "" {
		return fmt.Errorf("%s", errMsg)
	}

	encPassword, err := security.Encrypt(nsPassword)
	if err != nil {
		return fmt.Errorf("Failed to encrypt Nightscout password: %w", err)
	}

	encUrl, err := security.Encrypt(nsUrl)
	if err != nil {
		return fmt.Errorf("Failed to encrypt Nightscout URL: %w", err)
	}

	uniqHash, err := generateUniqueHash(security, nsUrl, nsPassword, consts.SyncSourceTypeNightscout, "")
	if err != nil {
		return fmt.Errorf("Failed to generate unique hash for Nightscout sync source: %w", err)
	}

	if nightscout == nil { // NEW
		err = controlDB.Queries.CreateNightscoutSyncSource(ctx, control.CreateNightscoutSyncSourceParams{
			Subdomain:     subdomain,
			Password:      encPassword,
			Enabled:       enabled,
			UrlOrUsername: encUrl,
			UniqueHash:    uniqHash,
		})
		if err != nil {
			return fmt.Errorf("Failed to create Nightscout sync source: %w", err)
		}
		return nil
	}

	if nsPassword == consts.PASSWORD_GARBAGE_VALUE {
		err = controlDB.Queries.UpdateNightscoutSyncSourceWithoutPassword(ctx, control.UpdateNightscoutSyncSourceWithoutPasswordParams{
			Subdomain:     subdomain,
			Enabled:       enabled,
			UrlOrUsername: encUrl,
			UniqueHash:    uniqHash,
		})
	} else {
		err = controlDB.Queries.UpdateNightscoutSyncSource(ctx, control.UpdateNightscoutSyncSourceParams{
			Subdomain:     subdomain,
			Password:      encPassword,
			Enabled:       enabled,
			UrlOrUsername: encUrl,
			UniqueHash:    uniqHash,
		})
	}

	if err != nil {
		return fmt.Errorf("Failed to update Nightscout sync source: %w", err)
	}

	return nil
}

func nsURLNormalizeValidate(rawUrl string, config config.ServerConfig) (string, error) {
	errMsg := ""
	nsURL, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid URL: %w", err)
	}

	scheme := strings.ToLower(nsURL.Scheme)
	host := strings.ToLower(nsURL.Hostname())
	path := nsURL.Path
	params := nsURL.RawQuery
	port := nsURL.Port()

	if (port == "80" && scheme == "http") || (port == "443" && scheme == "https") {
		port = "" // Remove default ports
	}

	if scheme != "https" && !(config.IsDev() && scheme == "http") {
		errMsg += "Nightscout URL must use HTTPS."
	}
	if path != "" && path != "/" {
		errMsg += "Nightscout URL cannot have a path. "
	}
	if params != "" {
		errMsg += "Nightscout URL cannot have query parameters. "
	}

	err = httputils.IsNightscoutURLBlocked(host, config)
	if err != nil {
		errMsg += "Nightscout URL is blocked." + err.Error()
	}

	if errMsg != "" {
		return "", fmt.Errorf("%s", errMsg)
	}

	// Normalize the URL by removing default Ports

	normalizedUrl := fmt.Sprintf("%s://%s", scheme, host)
	// if the port is specified and is not the default port
	if port != "" {
		normalizedUrl = fmt.Sprintf("%s://%s:%s", scheme, host, port)
	}

	return normalizedUrl, nil
}

func saveDexcomSettings(ctx context.Context, subdomain string, controlDB *db.ControlDB, security *security.SecurityConfig, dexcom *control.UserSyncSource, dexUsername string, dexPassword string, region consts.DexcomRegion, dexEnabled bool) error {
	var err error
	errMsg := ""
	if dexUsername == "" {
		errMsg += "Dexcom username cannot be empty. "
	}
	if dexPassword == "" {
		errMsg += "Dexcom password cannot be empty. "
	}

	if errMsg != "" {
		return fmt.Errorf("%s", errMsg)
	}

	encPassword, err := security.Encrypt(dexPassword)
	if err != nil {
		return fmt.Errorf("Failed to encrypt Dexcom password: %w", err)
	}

	encUsername, err := security.Encrypt(dexUsername)
	if err != nil {
		return fmt.Errorf("Failed to encrypt Dexcom username: %w", err)
	}

	uniqHash, err := generateUniqueHash(security, dexUsername, dexPassword, consts.SyncSourceTypeDexcom, string(region))
	if err != nil {
		return fmt.Errorf("Failed to generate unique hash for Dexcom sync source: %w", err)
	}

	if dexcom == nil { // NEW

		err = controlDB.Queries.CreateDexcomSyncSource(ctx, control.CreateDexcomSyncSourceParams{
			Subdomain:     subdomain,
			Password:      encPassword,
			Enabled:       dexEnabled,
			UrlOrUsername: encUsername,
			DexcomRegion:  region,
			UniqueHash:    uniqHash,
		})
		if err != nil {
			return fmt.Errorf("Failed to create Dexcom sync source: %w", err)
		}
		return nil
	}

	if dexPassword == consts.PASSWORD_GARBAGE_VALUE {
		err = controlDB.Queries.UpdateDexcomSyncSourceWithoutPassword(ctx, control.UpdateDexcomSyncSourceWithoutPasswordParams{
			Subdomain:     subdomain,
			Enabled:       dexEnabled,
			UrlOrUsername: encUsername,
			DexcomRegion:  region,
			UniqueHash:    uniqHash,
		})
	} else {
		err = controlDB.Queries.UpdateDexcomSyncSource(ctx, control.UpdateDexcomSyncSourceParams{
			Subdomain:     subdomain,
			Password:      encPassword,
			Enabled:       dexEnabled,
			UrlOrUsername: encUsername,
			DexcomRegion:  region,
			UniqueHash:    uniqHash,
		})
	}

	if err != nil {
		return fmt.Errorf("Failed to update Dexcom sync source: %w", err)
	}

	return nil
}

func generateUniqueHash(security *security.SecurityConfig, plainTextURLOrUsername string, plaintTextPassword string, syncSourceType consts.SyncSourceType, dexcomRegion string) (string, error) {

	// Any two users with this constalations will be making the exact same request to the same URL.
	// We are deduping these requests to prevent multiple calls
	uniqString := fmt.Sprintf("%s:%s:%s", syncSourceType, plainTextURLOrUsername, plaintTextPassword)
	if syncSourceType == consts.SyncSourceTypeDexcom {
		uniqString = fmt.Sprintf("%s:%s:%s:%s", syncSourceType, plainTextURLOrUsername, plaintTextPassword, dexcomRegion)
	}

	uniqHash, err := security.GenerateHash(uniqString)
	if err != nil {
		return "", fmt.Errorf("error generating hash for sync job: %w", err)
	}

	return uniqHash, nil
}
