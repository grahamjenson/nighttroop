package httputils

import (
	"log/slog"
	"net/http"
	"slices"

	"nighttroop.t1d.nz/utils/config"
)

// AllowedCountriesMiddleware refuses creating a site, signing in or recovering from outside the server's
// ALLOWED_COUNTRIES. With none set every country is allowed, as is an address whose country is not known, such as a
// private one. What is already signed in, and the Nightscout API that phones upload with, are not limited, so a
// person travelling keeps their data flowing.
func AllowedCountriesMiddleware(cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(cfg.AllowedCountries) > 0 {
				country := cfg.GeoIP.Locate(GetIPFromHTTPContext(r.Context())).Country
				if country != "" && !slices.Contains(cfg.AllowedCountries, country) {
					slog.InfoContext(r.Context(), "Refused a request from a country not allowed", "country", country, "path", r.URL.Path)
					http.Error(w, "Nighttroop is not available in your country", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
