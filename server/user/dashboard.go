package user

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	userTemplates "nighttroop.t1d.nz/server/assets/templates/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		err := userTemplates.DashboardHtml(w, subdomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
