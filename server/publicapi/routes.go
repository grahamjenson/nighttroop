package publicapi

import (
	"log/slog"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/publicmodels"
	"nighttroop.t1d.nz/server/assets/docs"
	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

type CreateUserResponse struct {
	SiteSubdomain   string `json:"site_subdomain"`
	APIKey          string `json:"api_key"`
	TOTPSecret      string `json:"totp_secret"`
	BackupKey       string `json:"backup_key"`
	PWDNSToken      string `json:"pwd_nstoken"`
	ReadOnlyNSToken string `json:"read_only_nstoken"`
}

func GETMainPageHandler(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := public_templ.CreateSite().Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func POSTCreateSiteHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed, remainingTime := httputils.CreateSiteWithinRateLimit(config, r)
		if !allowed {
			if httputils.GetAcceptType(r).API() {
				apiutils.WriteJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests",
				})
				return
			}
			w.WriteHeader(http.StatusTooManyRequests)
			public_templ.CreateSiteRateLimited(remainingTime+1).Render(r.Context(), w)
			return
		}

		// Check to see if there is 10k sites and return error if there is
		if controlDB.GetNSitesDBs() > config.MaxSites {
			// Plain text, as there is no page for a full server yet, unlike the rate limit's above
			http.Error(w, "Full: Too many sites", http.StatusTooManyRequests)
			return
		}

		siteCreateResult, err := controlDB.CreateSite(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "Unable to create site", "error", err)
			http.Error(w, "Unable to Create Site", http.StatusInternalServerError)
			return
		}

		// The site starts in the time zone and language of the browser that created it, with its country's formats
		params := &controldb.UpdateSitesSettingsParams{ID: siteCreateResult.SiteDB.Subdomain}
		settings.SetSettings(params, settings.BrowserSettings(r.FormValue("timezone"), r.FormValue("language")))
		if err := controlDB.Queries.UpdateSitesSettings(r.Context(), *params); err != nil {
			slog.ErrorContext(r.Context(), "Unable to set the new site's settings", "error", err)
		}

		responseModel := CreateUserResponse{
			SiteSubdomain:   siteCreateResult.SiteDB.Subdomain,
			APIKey:          siteCreateResult.AdminPassword,
			BackupKey:       siteCreateResult.BackupPassword,
			TOTPSecret:      siteCreateResult.TotpSecret,
			PWDNSToken:      siteCreateResult.PwdNSToken,
			ReadOnlyNSToken: siteCreateResult.ReadOnlyNSToken,
		}

		slog.InfoContext(r.Context(), "Number of Sites", "count", controlDB.GetNSitesDBs(), "routines", runtime.NumGoroutine())

		if httputils.GetAcceptType(r).API() {
			apiutils.WriteJSON(w, http.StatusOK, responseModel)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		err = public_templ.CreateSiteInfo(publicmodels.NewSiteInfo{
			URL:             siteCreateResult.SiteDB.URL(),
			AdminPassword:   siteCreateResult.AdminPassword,
			AdminTotpSecret: siteCreateResult.TotpSecret,
			PwdNSToken:      siteCreateResult.PwdNSToken,
			PwdNSURL:        siteCreateResult.SiteDB.NSURL(),
			ReadOnlyNSToken: siteCreateResult.ReadOnlyNSToken,
			BackupKey:       siteCreateResult.BackupPassword,
			RecoverURL:      config.RecoverURL(),
		}).Render(r.Context(), w)
		if err != nil {
			slog.ErrorContext(r.Context(), "Unable to create site", "error", err)
			http.Error(w, "Unable to Create Site", http.StatusInternalServerError)
		}

	}
}

// LegalRoutes serves the terms of service and privacy policy, with the server's support email, so every host's footer
// links to its own /terms and /privacy
func LegalRoutes(r chi.Router) {
	for path, document := range map[string]string{"/terms": docs.TermsOfService, "/privacy": docs.PrivacyPolicy} {
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := public_templ.Legal(docs.WithSupportEmail(document, httputils.SupportEmail(r.Context()))).Render(r.Context(), w); err != nil {
				slog.ErrorContext(r.Context(), "failed to render a legal document", "path", path, "error", err)
			}
		})
	}
}
