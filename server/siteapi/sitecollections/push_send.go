package sitecollections

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/push"
)

// maxPushDevices is the most push registrations a page lists, far more than a site's identities register
const maxPushDevices = 500

// PushDevicesHandler lists the apps on devices that pushes are sent to, with whose they are, and a button to test each:
// an admin sees every identity's on the site, and anyone else their own
func PushDevicesHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitesettings.PushRegistrationsCollection{}).GetName(), siteapiutils.ActionSearch)
	if siteContext == nil {
		return
	}
	rows, err := siteContext.ControlDB.Queries.SearchPushRegistrationsWithIdentity(r.Context(), controldb.SearchPushRegistrationsWithIdentityParams{
		SiteID: siteContext.SiteID, IdentityID: sitesettings.PushIdentity(siteContext), Limit: maxPushDevices})
	if err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	registrations := make([]*sitemodels.PushRegistration, 0, len(rows))
	for _, row := range rows {
		registration := sitemodels.PushRegistrationFromDB(&row.PushRegistration)
		registration.IdentityName = row.IdentityName
		registrations = append(registrations, registration)
	}
	admin := siteContext.IdentityModel.Role == consts.IdentityRoleAdmin
	WriteResponse(w, r, siteContext, http.StatusOK, registrations, site_settings_templ.PushDevices(registrations, admin), true)
}

// PushTestHandler sends a test notification to a push registration the signed in identity may see, see sitesettings.PushIdentity,
// so they can see pushes reach it, answering with the registration as it is after
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
	get := controldb.GetPushRegistrationParams{ID: id, SiteID: siteContext.SiteID, IdentityID: sitesettings.PushIdentity(siteContext)}
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
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.no_push_credential", i18n.Data{"App": registration.App})))
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
	tested := sitemodels.PushRegistrationFromDB(registration)
	WriteResponse(w, r, siteContext, http.StatusOK, tested, site_settings_templ.PushDeviceTested(tested), false)
}
