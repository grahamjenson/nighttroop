package controlapi

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/control_templ/control_setup_templ"
	"nighttroop.t1d.nz/server/components/errortoast"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// GETSetupPageHandler asks for the setup code a new server printed to its logs
func GETSetupPageHandler() http.HandlerFunc {
	return templ.Handler(control_setup_templ.Setup()).ServeHTTP
}

// POSTSetupHandler shows the control admin's login to save, given the setup code
func POSTSetupHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupWithinRateLimit(w, r, config) {
			return
		}
		admin, err := controlDB.BeginSetup(r.FormValue("code"))
		if err != nil {
			slog.InfoContext(r.Context(), "Setup code refused", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			errortoast.Toast("", "That is not the setup code in the server's logs.").Render(r.Context(), w)
			return
		}
		// The answer is credentials, which nothing between should keep
		w.Header().Set("Cache-Control", "no-store")
		control_setup_templ.SetupCredentials(admin.Name, admin.Password, admin.TotpSecret, security.TOTPURI(admin.Name, config.SetupURL(), admin.TotpSecret), r.FormValue("code")).Render(r.Context(), w)
	}
}

// POSTSetupConfirmHandler saves the control admin once a code from its TOTP secret shows it was saved
func POSTSetupConfirmHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupWithinRateLimit(w, r, config) {
			return
		}
		if err := controlDB.CompleteSetup(r.Context(), r.FormValue("code"), r.FormValue("totp")); err != nil {
			slog.InfoContext(r.Context(), "Setup confirmation refused", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			errortoast.Toast("", "That code is not from the TOTP secret shown.").Render(r.Context(), w)
			return
		}
		slog.InfoContext(r.Context(), "Nighttroop is set up, with its control admin")
		control_setup_templ.SetupComplete().Render(r.Context(), w)
	}
}

// setupWithinRateLimit limits guesses at the setup code as logins are limited, by the IP
func setupWithinRateLimit(w http.ResponseWriter, r *http.Request, config *config.ServerConfig) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(config, r, ":setup"); !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
		errortoast.Toast("", "Too many attempts, wait a bit and try again.").Render(r.Context(), w)
		return false
	}
	return true
}
