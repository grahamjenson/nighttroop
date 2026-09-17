package control

import (
	"context"
	"encoding/json"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/control_templ/control_dashboard_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlContext := httputils.GetControlContext(r.Context())

		w.Header().Set("Content-Type", "text/html")

		err := control_dashboard_templ.Dashboard(controlContext).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetUsersHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := controlDB.Queries.GetAllSites(context.TODO())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
