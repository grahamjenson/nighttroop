package site

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/utils/config"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		//siteContext := httputils.GetSiteContext(r.Context())
		err := site_templ.Dashboard().Render(r.Context(), w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
