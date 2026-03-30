package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	controlTemplates "nighttroop.t1d.nz/server/assets/templates/control"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Render Dashboard")
		w.Header().Set("Content-Type", "text/html")
		err := controlTemplates.Dashboard(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetUsersHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := controlDB.UsersAll.Query(context.TODO())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func GETPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userDB := httputils.GetUserFromHTTPContext(r.Context())

		if userDB == nil {
			fmt.Println("User NOT found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		accept := r.Header.Get("Accept")

		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "{\"ping\": \"pong\"}"))
			return
		}

		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Pong"))
			return
		}

	}
}
