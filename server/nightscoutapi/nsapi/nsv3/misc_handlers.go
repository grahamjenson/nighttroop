package nsv3

import (
	"log/slog"
	"net/http"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// Version reports the application and API version metadata for API v3.
func Version(store nsutils.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"version":    "3.0.4",
				"apiVersion": "3.0.4",
				"srvDate":    time.Now().UnixMilli(),
			},
		})
	}
}

//	{
//	    "result": {
//	      "apiPermissions": {
//	        "devicestatus": "crud",
//	        "entries": "crud",
//	        "food": "crud",
//	        "profile": "crud",
//	        "settings": "crud",
//	        "treatments": "crud"
//	      },
//	      "apiVersion": "3.0.3-alpha",
//	      "srvDate": 1787540907702,
//	      "storage": {
//	        "storage": "mongodb",
//	        "version": "6.0.28"
//	      },
//	      "version": "15.0.7"
//	    },
//	    "status": 200
//	}
//
// Status returns a lightweight authenticated health payload for API v3.
func Status(_ nsutils.Store) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate":    time.Now().UnixMilli(),
				"apiVersion": "0.0.1",
				"version":    "0.0.1",
				"storage": map[string]string{
					"storage": "sqlite",
					"version": "0.0.1",
				},
				"apiPermissions": map[string]string{
					"devicestatus": "crud",
					"entries":      "crud",
					"food":         "crud",
					"profile":      "crud",
					"settings":     "crud",
					"treatments":   "crud",
				},
			},
		})
	}
}

// Test is a minimal authenticated endpoint used for API v3 connectivity checks.
func Test(_ nsutils.Store) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": "ok"})
	}
}

// LastModified returns the newest modification timestamp for each known
// collection.
func LastModified(store nsutils.Store) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {

		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext == nil {
			slog.Error("site context not available")
			http.Error(w, nsutils.ErrMissingCredentials.Error(), http.StatusUnauthorized)
			return
		}

		lasModified, err := siteContext.SiteDB.Queries.LastModifiedCollections(r.Context())
		if err != nil {
			slog.Error("failed to query last modified timestamps from database", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate": time.Now().UnixMilli(),
				"collections": map[string]int64{
					"entries":      lasModified.GlucoseReadingUpdatedAt,
					"treatments":   lasModified.EventGroupUpdatedAt,
					"devicestatus": lasModified.DeviceReportUpdatedAt,
					"profile":      lasModified.ProfilePresetUpdatedAt,
					"food":         lasModified.FoodPresetUpdatedAt,
					"settings":     0, // TODO: I guess save the last time someone changed the settings
				},
			},
		})
	}
}
