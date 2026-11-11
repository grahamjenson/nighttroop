package controlapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"nighttroop.t1d.nz/server/components/control_templ/control_rate_limits_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// RateLimitsHandler is how close each kind of rate limiter is to refusing requests, which its page asks for every 10
// seconds
func RateLimitsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	limiters := controlContext.ControlDB.ServerConfig.RateLimiters
	stats := limiters.Stats(time.Now())
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, stats, control_rate_limits_templ.RateLimits(stats, limiters.Settings(), controlContext.IdentityModel.Role == consts.ControlRoleAdmin))
}

// SetRateLimitsHandler sets every rate limit, from JSON, {name: {per_minute, burst}}, or the page's form, with
// <name>-per_minute and <name>-burst, in place of the env's: one left out, or the env's own, uses the env's. Only
// admins can, and a rate that would refuse every request is refused.
func SetRateLimitsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	limiters := controlContext.ControlDB.ServerConfig.RateLimiters
	rates, herr := ratesFromRequest(r, limiters.Settings())
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	if err := controlContext.ControlDB.SetRateLimits(r.Context(), rates, controlContext.IdentityModel.Name); err != nil {
		slog.ErrorContext(r.Context(), "Unable to set the rate limits", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	settings := limiters.Settings()
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, settings, control_rate_limits_templ.RateSettingsSaved(settings, limiters.Stats(time.Now())))
}

func ratesFromRequest(r *http.Request, settings []config.RateSetting) (map[string]config.Rate, *apiutils.HTTPError) {
	rates := map[string]config.Rate{}
	if httputils.GetContentType(r) == httputils.ContentTypeJSON {
		sent, herr := apiutils.ExtractModelFromRequest[map[string]config.Rate](httputils.ContentTypeJSON, r)
		if herr != nil {
			return nil, herr
		}
		rates = *sent
	} else if err := r.ParseForm(); err != nil {
		return nil, apiutils.ErrBadForm
	} else {
		// Both blank uses the env's
		for _, setting := range settings {
			perMinuteText, burstText := r.Form.Get(setting.Name+"-per_minute"), r.Form.Get(setting.Name+"-burst")
			if perMinuteText == "" && burstText == "" {
				continue
			}
			perMinute, err := strconv.ParseFloat(perMinuteText, 64)
			burst, burstErr := strconv.Atoi(burstText)
			if err != nil || burstErr != nil {
				return nil, apiutils.ErrBadBodyWithHint(setting.Name + " needs a number for its rate and a whole number for its burst")
			}
			rates[setting.Name] = config.Rate{PerMinute: perMinute, Burst: burst}
		}
	}
	for _, setting := range settings {
		if rate, ok := rates[setting.Name]; ok && !rate.Valid() {
			return nil, apiutils.ErrBadBodyWithHint(setting.Name + " needs a rate above 0 and a burst of at least 1")
		}
	}
	return rates, nil
}
