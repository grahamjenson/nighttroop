package publicapi

import (
	"log/slog"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/models/publicmodels"
	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GETRecoverSitePageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := public_templ.RecoverSite().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// POSTRecoverSiteHandler is the recover page's form: a site's ID and backup key give its admin a new password and TOTP,
// and the site a new backup key, shown on the site's own address as a new site's credentials are. Any failure is the
// same 401, so it says nothing of why, or whether the site exists.
func POSTRecoverSiteHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		siteID := r.FormValue("site_id")
		// Each attempt is a bcrypt check, so it is limited as the site's own recover is, by the site and the IP
		if allowed, _ := httputils.SiteLoginAttemptsWithinRateLimit(config, r, siteID, ":backup_key"); !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			public_templ.RecoverSiteFailed("Too many attempts, wait a bit and try again.").Render(r.Context(), w)
			return
		}

		// The site is found first, so its URL is known before the key is used up
		siteDB, _, err := controlDB.GetSite(r.Context(), siteID)
		var recovered *db.RecoveredAdmin
		if err == nil {
			recovered, err = controlDB.RecoverSiteAdmin(r.Context(), siteID, r.FormValue("backup_key"))
		}
		if err != nil {
			slog.InfoContext(r.Context(), "Site recovery failed", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			public_templ.RecoverSiteFailed("That site ID and backup key do not match a site.").Render(r.Context(), w)
			return
		}

		// The credentials are shown on the site's own address, as a new site's are, where its admin logs in and adds a
		// passkey in place of any the recovery removed
		w.Header().Set("HX-Redirect", welcomeURL(siteDB.Subdomain, publicmodels.NewSiteInfo{
			URL:             siteDB.URL(),
			AdminPassword:   recovered.Password,
			AdminTotpSecret: recovered.TotpSecret,
			BackupKey:       recovered.BackupKey,
			RecoverURL:      config.RecoverURL(),
			Passkey:         true,
		}))
		w.WriteHeader(http.StatusOK)
	}
}
