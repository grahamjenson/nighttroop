package settings

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
	"nighttroop.t1d.nz/models/nighttroop"

	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

func GetSyncSettingsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

		ctx := r.Context()

		if acceptType.API() {
			// TODO return JSON of current settings, devices, tokens, last seen, etc
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "This will return the current sync settings in JSON format."}`))
			return
		}

		nightscout, dexcom, err := GetAllSiteSyncSources(ctx, controlDB, siteContext.Subdomain)

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		syncSettings := nighttroop.SiteSyncSettings{}

		if nightscout != nil {
			syncSettings.NightscoutSyncSettings = &nighttroop.NightscoutSyncSettings{}

			syncSettings.NightscoutSyncSettings.URL, err = config.Security.Decrypt(nightscout.UrlOrUsername)
			if err != nil {
				slog.Error("Failed to decrypt Nightscout URL from DB", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			syncSettings.NightscoutSyncSettings.Enabled = nightscout.Enabled
			syncSettings.NightscoutSyncSettings.ID = nightscout.ID

		}

		if dexcom != nil {
			syncSettings.DexcomSyncSettings = &nighttroop.DexcomSyncSettings{}
			syncSettings.DexcomSyncSettings.Username, err = config.Security.Decrypt(dexcom.UrlOrUsername)
			if err != nil {
				slog.Error("Failed to decrypt Dexcom username from DB", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			syncSettings.DexcomSyncSettings.USARegion = dexcom.DexcomRegion == consts.DexcomRegionUS
			syncSettings.DexcomSyncSettings.Enabled = dexcom.Enabled
			syncSettings.DexcomSyncSettings.ID = dexcom.ID

		}

		if acceptType == httputils.AcceptTypeHTML {
			err = site_settings_templ.SyncSettingsFull(syncSettings).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTML template", "subdomain", siteContext.Subdomain)
				return
			}
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			err = site_settings_templ.SyncSettingsPartial(syncSettings).Render(r.Context(), w)
			if err != nil {
				errors.TemplateFailureError(w, r, err, "Failed to render sync settings HTMX template", "subdomain", siteContext.Subdomain)
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

func GetAllSiteSyncSources(ctx context.Context, db *db.ControlDB, subdomain string) (nightscout *control.SiteSyncSource, dexcom *control.SiteSyncSource, err error) {
	// TODO: This might become a problem we could make faster with a dirty flag or something
	sync_sources, err := db.Queries.GetAllSiteSyncSourcesForSite(ctx, control.GetAllSiteSyncSourcesForSiteParams{
		SiteID: subdomain,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error getting sync sources: %w", err)
	}

	for _, ss := range sync_sources {
		if ss.Type == consts.SyncSourceTypeNightscout {
			nightscout = ss
		}
		if ss.Type == consts.SyncSourceTypeDexcom {
			dexcom = ss
		}
	}

	return nightscout, dexcom, nil

}

func POSTSyncSettingsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowed, seconds := httputils.UserUpdateSyncWithinRateLimit(config, r); !allowed {
			http.Error(w, "Too Many Requests wait "+strconv.Itoa(seconds)+" seconds", http.StatusTooManyRequests)
			return
		}

		syncSourceTypeStr := chi.URLParam(r, "type")
		if syncSourceTypeStr == "" {
			http.Error(w, "Bad Request: Missing type", http.StatusBadRequest)
			return
		}

		syncSourceType, err := consts.NewSyncSourceTypeFromString(strings.ToUpper(syncSourceTypeStr))
		if err != nil {
			http.Error(w, "Bad Request: Invalid type", http.StatusBadRequest)
			return
		}

		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

		ctx := r.Context()

		// 1. Parse and validate form inputs
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		nightscout, dexcom, err := GetAllSiteSyncSources(ctx, controlDB, siteContext.Subdomain)

		if err != nil {
			slog.Error("Failed to get sync sources from DB", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if syncSourceType == consts.SyncSourceTypeNightscout {
			err = saveNightScoutSettings(ctx, siteContext.Subdomain, controlDB, config, nightscout, r.FormValue("ns_url"), r.FormValue("ns_password"), r.FormValue("ns_enabled") == "on")
		} else {
			err = saveDexcomSettings(ctx, siteContext.Subdomain, controlDB, config.Security, dexcom, r.FormValue("dex_username"), r.FormValue("dex_password"), consts.NewDexcomRegionFromString(r.FormValue("dex_region")), r.FormValue("dex_enabled") == "on")
		}

		if err != nil {
			slog.Error("Validation error for sync settings inputs", "error", err)
			site_templ.Fail().Render(r.Context(), w)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After saving, redirect back to the sync settings page
			http.Redirect(w, r, "/settings/sync", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			site_templ.Success().Render(r.Context(), w)
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

func DELETESyncSettingsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		siteContext := httputils.GetSiteContext(r.Context())

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
			SiteID: siteContext.Subdomain,
			Type:   type_,
		})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if acceptType == httputils.AcceptTypeHTML {
			// After deletion, redirect back to the sync settings page
			http.Redirect(w, r, "/settings/sync", http.StatusSeeOther)
			return
		}

		if acceptType == httputils.AcceptTypeHTMX {
			// Return a success message that HTMX can use to trigger a toast notification
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			w.Header().Set("HX-Redirect", "/settings/sync")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Default to JSON response for unsupported types
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

func GetSyncSettingsReportsHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

///////
// UTIL METHODS
////////

func saveNightScoutSettings(ctx context.Context, subdomain string, controlDB *db.ControlDB, config *config.ServerConfig, nightscout *control.SiteSyncSource, nsUrl string, nsPassword string, enabled bool) error {
	nsUrl = strings.TrimSpace(nsUrl)
	if nsUrl == "" {
		return fmt.Errorf("Nightscout URL cannot be empty. ")
	}

	security := config.Security
	var err error
	errMsg := ""

	nsUrl, err = nsURLNormalizeValidate(nsUrl, config)
	if err != nil {
		slog.Error("Failed to normalize and validate Nightscout URL", "error", err, "url", nsUrl)
		errMsg += fmt.Sprintf("Nightscout URL '%s' is not a valid URL. %s", nsUrl, err)
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
			SiteID:        subdomain,
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
			SiteID:        subdomain,
			Enabled:       enabled,
			UrlOrUsername: encUrl,
			UniqueHash:    uniqHash,
		})
	} else {
		err = controlDB.Queries.UpdateNightscoutSyncSource(ctx, control.UpdateNightscoutSyncSourceParams{
			SiteID:        subdomain,
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

func nsURLNormalizeValidate(rawUrl string, config *config.ServerConfig) (string, error) {
	errMsg := ""
	nsURL, err := url.Parse(rawUrl)
	if err != nil {
		return rawUrl, fmt.Errorf("Invalid URL: %w", err)
	}

	scheme := strings.ToLower(nsURL.Scheme)
	host := strings.ToLower(nsURL.Hostname())
	path := nsURL.Path
	params := nsURL.RawQuery
	port := nsURL.Port()

	if (port == "80" && scheme == "http") || (port == "443" && scheme == "https") {
		port = "" // Remove default ports
	}

	if scheme != "https" && !(config.IsDevOrTest() && scheme == "http") {
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
		return rawUrl, fmt.Errorf("%s", errMsg)
	}

	// Normalize the URL by removing default Ports

	normalizedUrl := fmt.Sprintf("%s://%s", scheme, host)
	// if the port is specified and is not the default port
	if port != "" {
		normalizedUrl = fmt.Sprintf("%s://%s:%s", scheme, host, port)
	}

	return normalizedUrl, nil
}

func saveDexcomSettings(ctx context.Context, subdomain string, controlDB *db.ControlDB, security *security.SecurityConfig, dexcom *control.SiteSyncSource, dexUsername string, dexPassword string, region consts.DexcomRegion, dexEnabled bool) error {
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
			SiteID:        subdomain,
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
			SiteID:        subdomain,
			Enabled:       dexEnabled,
			UrlOrUsername: encUsername,
			DexcomRegion:  region,
			UniqueHash:    uniqHash,
		})
	} else {
		err = controlDB.Queries.UpdateDexcomSyncSource(ctx, control.UpdateDexcomSyncSourceParams{
			SiteID:        subdomain,
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
