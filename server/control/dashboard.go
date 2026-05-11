package control

import (
	"context"
	"encoding/json"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	controlTemplates "nighttroop.t1d.nz/server/assets/templates/control"
	"nighttroop.t1d.nz/utils/config"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/html")
		err := controlTemplates.DashboardHtml(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetUsersHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := controlDB.Stmts.UserAll.Query(context.TODO())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
