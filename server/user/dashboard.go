package user

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/templates"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		err := templates.UserDashboardHTML(w, subdomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
