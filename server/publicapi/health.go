package publicapi

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/apiutils"
)

// GETHealthHandler is whether the server is up and its control DB answers, for Docker's health check and deploys
func GETHealthHandler(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := controlDB.SQLiteDB.PingContext(r.Context()); err != nil {
			apiutils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
