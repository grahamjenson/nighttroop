package nsv2

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
	"nighttroop.t1d.nz/utils/types"
)

// Loop refuses a bolus over 30 U and carbs over 250 g itself; a command beyond them is refused here before it is sent
const (
	maxRemoteBolus        = 30
	maxRemoteCarbs        = 250
	maxRemoteAbsorption   = 24    // hours
	maxOverrideMinutes    = 10080 // a week
	loopCommandExpiration = 5 * time.Minute
)

// loopCommand is a remote command for Loop, as NightscoutKit, LoopFollow and Nightscout's careportal post it to
// /api/v2/notifications/loop, NightscoutKit with every value a string
type loopCommand struct {
	EventType        string                 `json:"eventType"`
	Reason           string                 `json:"reason"`
	ReasonDisplay    string                 `json:"reasonDisplay"`
	Duration         *types.FlexibleFloat64 `json:"duration"`
	RemoteCarbs      *types.FlexibleFloat64 `json:"remoteCarbs"`
	RemoteAbsorption *types.FlexibleFloat64 `json:"remoteAbsorption"`
	RemoteBolus      *types.FlexibleFloat64 `json:"remoteBolus"`
	CreatedAt        *types.DateString      `json:"created_at"`
	OTP              types.ShortString      `json:"otp"`
	Notes            types.ShortString      `json:"notes"`
	EnteredBy        types.ShortString      `json:"enteredBy"`
}

// LoopNotification sends a remote command to Loop as an Apple push, as Nightscout does: an override started or
// cancelled, or carbs or a bolus with the one time password Loop checks. It goes to the Loop that last uploaded its
// profile, with the site's push credential for its bundle ID. Loop records what it does, so nothing is kept here.
func LoopNotification(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
	siteContext := httputils.GetSiteContext(r.Context())
	var command loopCommand
	if err := apiutils.ReadJSON(r, &command); err != nil {
		nsutils.WriteError(w, nsutils.BadDataBodyErrorWithHint(err.Error()))
		return
	}
	alert, err := loopCommandAlert(command, identity.Name, time.Now())
	if err != nil {
		nsutils.WriteError(w, nsutils.BadDataBodyErrorWithHint(err.Error()))
		return
	}
	q := siteContext.ControlDB.Queries
	registration, err := q.NewestPushRegistrationForApp(r.Context(), controldb.NewestPushRegistrationForAppParams{SiteID: siteContext.SiteID, App: consts.PushAppLoop})
	if errors.Is(err, sql.ErrNoRows) {
		nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusBadRequest, Message: "Loop notification failed: no Loop has uploaded its profile with its device token"})
		return
	}
	if err != nil {
		nsutils.WriteError(w, nsutils.ErrInternal)
		return
	}
	if allowed, _ := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SitePushLimiter(siteContext.SiteID)); !allowed {
		nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusTooManyRequests, Message: "Too many pushes: send it again shortly"})
		return
	}
	err = push.Send(r.Context(), q, siteContext.Config.Security, registration, alert)
	switch {
	case errors.Is(err, push.ErrNoCredential):
		nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusBadRequest, Message: "Loop notification failed: " + err.Error() + ": add Loop's APNs key in Push Certificates"})
	case errors.Is(err, push.ErrTokenGone):
		nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusGone, Message: "Loop notification failed: Loop is no longer registered on its device"})
	case err != nil:
		slog.WarnContext(r.Context(), "A Loop remote command push failed", "event_type", command.EventType, "error", err)
		nsutils.WriteError(w, &nsutils.HTTPError{Status: http.StatusBadGateway, Message: "Loop notification failed: " + err.Error()})
	default:
		// As Nightscout answers, and NightscoutKit only reads its status
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("OK"))
	}
}

// loopCommandAlert is the push Nightscout sends Loop for a command, with the fields Loop reads at its top, or why the
// command cannot be sent. Who sent it is its identity, in place of the IP address Nightscout sends.
func loopCommandAlert(c loopCommand, identityName string, now time.Time) (push.Alert, error) {
	enteredBy := c.EnteredBy.String()
	if enteredBy == "" {
		enteredBy = identityName
	}
	data := map[string]any{"remote-address": identityName, "entered-by": enteredBy,
		"sent-at": types.ISOString(now), "expiration": types.ISOString(now.Add(loopCommandExpiration))}
	if c.Notes != "" {
		data["notes"] = c.Notes.String()
	}
	if c.OTP != "" {
		data["otp"] = c.OTP.String()
	}
	var title, body string
	switch c.EventType {
	case "Temporary Override Cancel":
		data["cancel-temporary-override"] = "true"
		title = "Cancel Temporary Override"
	case "Temporary Override":
		if c.Reason == "" {
			return push.Alert{}, errors.New("a Temporary Override needs the override's name as its reason")
		}
		data["override-name"] = c.Reason
		// No duration, or 0, is until it is cancelled
		minutes := value(c.Duration)
		if !(minutes >= 0 && minutes <= maxOverrideMinutes) {
			return push.Alert{}, fmt.Errorf("an override's duration is 0 to %d minutes", maxOverrideMinutes)
		}
		if int64(minutes) > 0 {
			data["override-duration-minutes"] = int64(minutes)
		}
		title = cmp.Or(c.ReasonDisplay, c.Reason) + " Temporary Override"
	case "Remote Carbs Entry":
		carbs, absorption := value(c.RemoteCarbs), 3.0
		if !(carbs > 0 && carbs <= maxRemoteCarbs) {
			return push.Alert{}, fmt.Errorf("remoteCarbs must be more than 0 and at most %d grams", maxRemoteCarbs)
		}
		if value(c.RemoteAbsorption) > 0 {
			absorption = value(c.RemoteAbsorption)
		}
		if !(absorption <= maxRemoteAbsorption) {
			return push.Alert{}, fmt.Errorf("remoteAbsorption is at most %d hours", maxRemoteAbsorption)
		}
		data["carbs-entry"], data["absorption-time"] = carbs, absorption
		if c.CreatedAt != nil {
			data["start-time"] = types.ISOString(c.CreatedAt.Time())
		}
		title, body = fmt.Sprintf("Remote Carbs Entry: %g grams", carbs), fmt.Sprintf("Absorption Time: %g hours", absorption)
	case "Remote Bolus Entry":
		bolus := value(c.RemoteBolus)
		if !(bolus > 0 && bolus <= maxRemoteBolus) {
			return push.Alert{}, fmt.Errorf("remoteBolus must be more than 0 and at most %d U", maxRemoteBolus)
		}
		data["bolus-entry"] = bolus
		title = fmt.Sprintf("Remote Bolus Entry: %g U", bolus)
	default:
		return push.Alert{}, fmt.Errorf("unknown eventType %q", c.EventType)
	}
	for _, part := range []string{c.Notes.String(), enteredBy} {
		if part != "" {
			body = strings.TrimPrefix(body+" - "+part, " - ")
		}
	}
	return push.Alert{Title: title, Body: body, Data: data}, nil
}

// value is a number sent, or 0 when it is not
func value(f *types.FlexibleFloat64) float64 {
	if f == nil {
		return 0
	}
	return f.Float64()
}
