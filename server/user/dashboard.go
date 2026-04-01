package user

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	userTemplates "nighttroop.t1d.nz/server/assets/templates/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func GetDashboardHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Render Dashboard")
		w.Header().Set("Content-Type", "text/html")
		subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
		err := userTemplates.DashboardHtml(w, subdomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GETPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userDB, _ := httputils.GetUserDBAndModelFromHTTPContext(r.Context())

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
