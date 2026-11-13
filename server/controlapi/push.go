package controlapi

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_push_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// maxPushDevices is the most push registrations the page lists, the newest
const maxPushDevices = 500

// PushDevicesHandler is every site's apps registered for pushes, or with ?site= one site's, newest first, with whose
// they are, and for an admin a button to test each
func PushDevicesHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	var siteID *string
	if site := r.URL.Query().Get("site"); site != "" {
		siteID = &site
	}
	rows, err := controlContext.ControlDB.Queries.SearchAllPushRegistrations(r.Context(), controldb.SearchAllPushRegistrationsParams{SiteID: siteID, Limit: maxPushDevices})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list push registrations", "error", err)
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	devices := make([]*controlmodels.PushRegistration, 0, len(rows))
	for _, row := range rows {
		registration := sitemodels.PushRegistrationFromDB(&row.PushRegistration)
		registration.IdentityName = row.IdentityName
		devices = append(devices, &controlmodels.PushRegistration{PushRegistration: registration, SiteID: row.PushRegistration.SiteID})
	}
	admin := controlContext.IdentityModel.Role == consts.ControlRoleAdmin
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, devices, control_push_templ.PushDevices(devices, admin))
}

// PushTestHandler sends a test push to any site's registration, within its site's push limit, answering with the
// registration as it is after. Only admins can.
func PushTestHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		apiutils.ErrNotFound.Write(acceptType, w)
		return
	}
	q, config := controlContext.ControlDB.Queries, controlContext.ControlDB.ServerConfig
	registration, err := q.GetAnySitePushRegistration(r.Context(), controldb.GetAnySitePushRegistrationParams{ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		apiutils.ErrNotFound.Write(acceptType, w)
		return
	}
	if err != nil {
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	if allowed, seconds := httputils.WithinRateLimitAll(config.RateLimiters.SitePushLimiter(registration.SiteID)); !allowed {
		apiutils.ErrRateLimitExceeded(seconds).Write(acceptType, w)
		return
	}
	err = push.Send(r.Context(), q, config.Security, registration, push.Alert{Title: "Nighttroop", Body: "A test push from the server's control site"})
	switch {
	case errors.Is(err, push.ErrNoCredential):
		apiutils.ErrConstraintCheckFailedWithHint(err).Write(acceptType, w)
		return
	case errors.Is(err, push.ErrTokenGone):
		(&apiutils.HTTPError{Status: http.StatusGone, Message: "The app is no longer registered on this device, so its registration is deleted"}).Write(acceptType, w)
		return
	case err != nil:
		slog.WarnContext(r.Context(), "A control test push failed", "site_id", registration.SiteID, "app", registration.App, "error", err)
		(&apiutils.HTTPError{Status: http.StatusBadGateway, Message: err.Error()}).Write(acceptType, w)
		return
	}
	if registration, err = q.GetAnySitePushRegistration(r.Context(), controldb.GetAnySitePushRegistrationParams{ID: id}); err != nil {
		apiutils.ErrInternalServerError.Write(acceptType, w)
		return
	}
	tested := sitemodels.PushRegistrationFromDB(registration)
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, tested, control_push_templ.PushDeviceTested(tested))
}
