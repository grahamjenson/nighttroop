package v3

import (
	"net/http"
	"time"

	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nsutils"
)

// Version reports the application and API version metadata for API v3.
func Version(dep nsutils.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"version":    "3.0.4",
				"apiVersion": "3.0.4",
				"srvDate":    time.Now().UnixMilli(),
			},
		})
	}
}

// Status returns a lightweight authenticated health payload for API v3.
func Status(_ nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate": time.Now().UnixMilli(),
			},
		})
	}
}

// Test is a minimal authenticated endpoint used for API v3 connectivity checks.
func Test(_ nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		nsutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "result": "ok"})
	}
}

// LastModified returns the newest modification timestamp for each known
// collection.
func LastModified(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		collections := []nsutils.Collection{nsutils.Entries, nsutils.Treatments, nsutils.Devicestatus, nsutils.Profile, nsutils.Food}
		lastModifiedMap := make(map[nsutils.Collection]int64)
		for _, collection := range collections {
			record, err := dep.Store.Search(r.Context(), collection, &nsutils.Query{Limit: ptr(nightscout.FlexibleInt64(1)), SortDesc: ptr(nightscout.FlexibleBool(true)), Sort: ptr(nightscout.ShortString("srv_modified"))})
			if err != nil || record == nil || len(record) == 0 {
				continue
			}
			lastModifiedMap[collection] = record[0].SrvModified
		}

		nsutils.WriteJSON(w, http.StatusOK, map[string]any{
			"status": http.StatusOK,
			"result": map[string]any{
				"srvDate":     time.Now().UnixMilli(),
				"collections": lastModifiedMap,
			},
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
