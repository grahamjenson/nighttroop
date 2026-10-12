package public

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/models/public"
	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
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
			w.WriteHeader(http.StatusTooManyRequests)
			if httputils.GetAcceptType(r).API() {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Too many requests",
				})
				return
			} else {
				public_templ.CreateSiteRateLimited(remainingTime+1).Render(r.Context(), w)
				return
			}
		}

		// Check to see if there is 10k sites and return error if there is
		if controlDB.GetNSitesDBs() > config.MaxSites {
			// TODO make this a template like above
			http.Error(w, "Full: Too many sites", http.StatusTooManyRequests)
			return
		}

		siteCreateResult, err := controlDB.CreateSite()
		if err != nil {
			slog.Error("Unable to create site", "error", err)
			http.Error(w, "Unable to Create Site", http.StatusInternalServerError)
			return
		}

		responseModel := CreateUserResponse{
			SiteSubdomain:   siteCreateResult.SiteDB.Subdomain,
			APIKey:          siteCreateResult.AdminPassword,
			BackupKey:       siteCreateResult.BackupPassword,
			TOTPSecret:      siteCreateResult.TotpSecret,
			PWDNSToken:      siteCreateResult.PwdNSToken,
			ReadOnlyNSToken: siteCreateResult.ReadOnlyNSToken,
		}

		slog.Info("Number of Sites", "count", controlDB.GetNSitesDBs(), "routines", runtime.NumGoroutine())

		if httputils.GetAcceptType(r).API() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(responseModel)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		err = public_templ.CreateSiteInfo(public.NewSiteInfo{
			URL:             siteCreateResult.SiteDB.URL(),
			AdminPassword:   siteCreateResult.AdminPassword,
			AdminTotpSecret: siteCreateResult.TotpSecret,
			PwdNSToken:      siteCreateResult.PwdNSToken,
			PwdNSURL:        siteCreateResult.SiteDB.NSURL(),
			BackupKey:       siteCreateResult.BackupPassword,
		}).Render(r.Context(), w)
		if err != nil {
			slog.Error("Unable to create site", "error", err)
			http.Error(w, "Unable to Create Site", http.StatusInternalServerError)
		}

	}
}
