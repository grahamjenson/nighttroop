package sitecollections

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// PushTestHandler sends a test notification to one of the signed in identity's push registrations, so they can see
// pushes reach it, answering with the registration as it is after
func PushTestHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.PushRegistrationsCollection{}).GetName(), siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	q := siteContext.ControlDB.Queries
	get := controldb.GetPushRegistrationParams{ID: id, SiteID: siteContext.SiteID, IdentityID: siteContext.IdentityModel.ID}
	registration, err := q.GetPushRegistration(r.Context(), get)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrNotFound)
		return
	}
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	if allowed, seconds := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SitePushLimiter(siteContext.SiteID)); !allowed {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(seconds))
		return
	}

	err = push.Send(r.Context(), q, siteContext.Config.Security, registration, push.Alert{Title: "Nighttroop",
		Body: "Pushes from " + siteContext.SiteDB.URL() + " reach this device"})
	switch {
	case errors.Is(err, push.ErrNoCredential):
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(err.Error()+": add one in Push Notifications"))
		return
	case errors.Is(err, push.ErrTokenGone):
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusGone, Message: "The app is no longer registered on this device, so its registration is deleted"})
		return
	case err != nil:
		slog.WarnContext(r.Context(), "A test push failed", "app", registration.App, "error", err)
		writeError(w, r, siteContext.AcceptType, &apiutils.HTTPError{Status: http.StatusBadGateway, Message: err.Error()})
		return
	}
	if registration, err = q.GetPushRegistration(r.Context(), get); err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	WriteResponse(w, r, siteContext, http.StatusOK, sitemodels.PushRegistrationFromDB(registration), templ.NopComponent, false)
}
