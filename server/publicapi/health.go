package publicapi

import (
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/apiutils"
)

// GETHealthHandler is whether the server is up, its control DB answers and no live site's DB was left out at start, for
// Docker's health check and deploys, so update.sh rolls back a release that starts without a site. It never says which.
func GETHealthHandler(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if controlDB.SitesMissing() || controlDB.SQLiteDB.PingContext(r.Context()) != nil {
			apiutils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
