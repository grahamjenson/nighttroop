package publicapi

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/apiutils"
)

// GETHealthHandler is whether the server is up and its control DB answers, for Docker's health check and deploys. A
// server is only up once every live site's DB is opened and migrated, so update.sh rolls back a release that cannot.
func GETHealthHandler(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if controlDB.SQLiteDB.PingContext(r.Context()) != nil {
			apiutils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
