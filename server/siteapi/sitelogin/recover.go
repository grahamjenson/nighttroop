package sitelogin

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// POSTRecover gives a site's owner its admin back: a body of {"backup_key": "..."}, the key shown when the site was
// made, answers with the admin's new name, password and TOTP secret, and the next backup key, as the key works once.
// Any failure is the same 401, so it says nothing of why.
func POSTRecover(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())
		// Each attempt is a bcrypt check, so a site's recoveries are limited as its logins are, by a name no identity has
		if allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(config, r, siteContext.SiteID, ":backup_key"); !allowed {
			http.Error(w, "Too many attempts, wait a bit", http.StatusTooManyRequests)
			return
		}

		var body struct {
			BackupKey string `json:"backup_key"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		var recovered *db.RecoveredAdmin
		err := json.NewDecoder(r.Body).Decode(&body)
		if err == nil {
			recovered, err = controlDB.RecoverSiteAdmin(r.Context(), siteContext.SiteID, body.BackupKey)
		}
		if err != nil {
			slog.InfoContext(r.Context(), "Site recovery failed", "error", err)
			http.Error(w, "Invalid backup key", http.StatusUnauthorized)
			return
		}

		// The answer is credentials, which nothing between should keep
		w.Header().Set("Cache-Control", "no-store")
		apiutils.WriteJSON(w, http.StatusOK, recovered)
	}
}
