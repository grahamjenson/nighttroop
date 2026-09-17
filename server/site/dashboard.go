package site

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/site_templ/site_dashboard_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := httputils.GetSiteContext(r.Context())

		w.Header().Set("Content-Type", "text/html")
		err := site_dashboard_templ.Dashboard(siteContext).Render(r.Context(), w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
