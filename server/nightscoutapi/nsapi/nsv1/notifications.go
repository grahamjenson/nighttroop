package nsv1

import (
	"net/http"
	"strconv"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// AckNotifications is Nightscout's GET /api/v1/notifications/ack?level=&group=&time=, which a follower acks an alarm
// with: it snoozes the identity's alarms in the group up to the level for time milliseconds, as the /alarm socket's
// ack does, and answers OK
func AckNotifications() http.HandlerFunc {
	return nsutils.Require("api:*:read", false, func(w http.ResponseWriter, r *http.Request, _ *nsutils.Identity) {
		query := r.URL.Query()
		level, _ := strconv.Atoi(query.Get("level"))
		silenceMs, _ := strconv.ParseInt(query.Get("time"), 10, 64)
		siteContext := httputils.GetSiteContext(r.Context())
		if !nsutils.SnoozeAlarms(siteContext, siteContext.IdentityModel.ID, alarms.Level(level), query.Get("group"), silenceMs) {
			nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusTooManyRequests, Message: "Too many requests"})
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK"))
	})
}

// AdminNotifies is Nightscout's notices for the site's admin, such as a weak API secret. Nighttroop has none of its
// settings to warn about, so there are none.
func AdminNotifies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK, "message": map[string]any{"notifies": []any{}, "notifyCount": 0}})
	}
}
