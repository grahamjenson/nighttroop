package control

import (
	"context"
	"encoding/json"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/control_templ"
	"nighttroop.t1d.nz/utils/config"
)

func GetDashboardHandler(controlDB *db.ControlDB, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/html")

		err := control_templ.Dashboard().Render(r.Context(), w)
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
