package controlapi

import (
	"net/http"
	"time"

	"nighttroop.t1d.nz/server/components/control_templ/control_rate_limits_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/httputils"
)

// RateLimitsHandler is how close each kind of rate limiter is to refusing requests, which its page asks for every 10
// seconds
func RateLimitsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	stats := controlContext.ControlDB.ServerConfig.RateLimiters.Stats(time.Now())
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, stats, control_rate_limits_templ.RateLimits(stats))
}
