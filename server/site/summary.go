package site

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/models/nighttroop"
	"nighttroop.t1d.nz/server/components/site_templ/site_dashboard_templ"
	"nighttroop.t1d.nz/utils/config"
)

func GetSummaryHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/html")
		err := site_dashboard_templ.Dashboard(nighttroop.Dashboard{}).Render(r.Context(), w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
