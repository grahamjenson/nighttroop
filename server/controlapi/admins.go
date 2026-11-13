package controlapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_admins_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
)

// maxControlAdmins is the most admins control can have, far more than a server needs, as each is made with bcrypt
const maxControlAdmins = 20

var (
	errYourself  = &apiutils.HTTPError{Status: http.StatusForbidden, Message: "You cannot change or delete yourself here"}
	errLastAdmin = &apiutils.HTTPError{Status: http.StatusConflict, Message: "The last admin cannot be deleted or made read only"}
)

// controlAdminForm is an admin to add, its name and role, or the role to give one
type controlAdminForm struct {
	Name string             `json:"name"`
	Role consts.ControlRole `json:"role"`
}

// ControlAdminsHandler is control's admins, newest first, never with their secrets
func ControlAdminsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	rows, err := controlContext.ControlDB.Queries.AllControlIdentities(r.Context())
	admins, herr := apiutils.Searched(rows, err, "control admins", controlmodels.ControlAdminFromDB)
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, admins, control_admins_templ.Admins(admins))
}

// CreateControlAdminHandler adds an admin, of a name no other has and a role, showing its password and TOTP secret
// once. Only admins can.
func CreateControlAdminHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	form, err := adminForm(r)
	var admins []*controldb.ControlIdentity
	if err == nil {
		form.Name = strings.TrimSpace(form.Name)
		admins, err = controlContext.ControlDB.Queries.AllControlIdentities(r.Context())
	}
	if err == nil && (form.Name == "" || utf8.RuneCountInString(form.Name) > consts.MAX_IDENTITY_NAME_LENGTH) {
		err = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.admin_name"))
	}
	if err == nil && len(admins) >= maxControlAdmins {
		err = apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.max_admins", i18n.Data{"Max": maxControlAdmins}))
	}
	var values *db.IdentityValues
	var created *controldb.ControlIdentity
	if err == nil {
		values, err = controlContext.ControlDB.CreateControlIdentity(r.Context(), form.Name, form.Role, controlContext.IdentityModel)
	}
	if err == nil {
		created, err = controlContext.ControlDB.Queries.GetControlIdentity(r.Context(), controldb.GetControlIdentityParams{Name: form.Name})
	}
	if herr := apiutils.ErrorFor(err, "control admins"); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	slog.InfoContext(r.Context(), "Control admin added", "admin", created.Name, "role", created.Role, "by", controlContext.IdentityModel.Name)
	admin, credentials := controlmodels.ControlAdminFromDB(created), adminCredentials(created, values)
	w.Header().Set("Cache-Control", "no-store")
	controlcollections.WriteResponse(w, r, acceptType, http.StatusCreated, credentials,
		control_admins_templ.Added(controlContext.ControlDB.ServerConfig.LinkURL("control"), admin, credentials))
}

// SetControlAdminRoleHandler gives another admin a role, which signs it out. Only admins can.
func SetControlAdminRoleHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	identity, err := othersControlAdmin(r)
	var form *controlAdminForm
	if err == nil {
		form, err = adminForm(r)
	}
	if err == nil {
		err = controlContext.ControlDB.SetControlIdentityRole(r.Context(), identity, form.Role, controlContext.IdentityModel)
	}
	if herr := adminError(err); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	identity.Role = form.Role
	slog.InfoContext(r.Context(), "Control admin's role changed", "admin", identity.Name, "role", identity.Role, "by", controlContext.IdentityModel.Name)
	admin := controlmodels.ControlAdminFromDB(identity)
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, admin, control_admins_templ.RoleChanged(admin))
}

// DeleteControlAdminHandler deletes another admin, with its passkeys, which signs it out. Only admins can.
func DeleteControlAdminHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	identity, err := othersControlAdmin(r)
	if err == nil {
		err = controlContext.ControlDB.DeleteControlIdentity(r.Context(), identity)
	}
	if herr := adminError(err); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	slog.InfoContext(r.Context(), "Control admin deleted", "admin", identity.Name, "by", controlContext.IdentityModel.Name)
	admin := controlmodels.ControlAdminFromDB(identity)
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, admin, control_admins_templ.Deleted(admin))
}

// ResetControlAdminPasswordHandler gives another admin a new random password, shown once. Only admins can.
func ResetControlAdminPasswordHandler(w http.ResponseWriter, r *http.Request) {
	resetControlAdmin(w, r, (*db.ControlDB).ResetControlIdentityPassword)
}

// ResetControlAdminTOTPHandler gives another admin a new TOTP secret, shown once. Only admins can.
func ResetControlAdminTOTPHandler(w http.ResponseWriter, r *http.Request) {
	resetControlAdmin(w, r, (*db.ControlDB).ResetControlIdentityTOTP)
}

func resetControlAdmin(w http.ResponseWriter, r *http.Request, reset func(*db.ControlDB, context.Context, *controldb.ControlIdentity, *controldb.ControlIdentity) (*db.IdentityValues, error)) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	identity, err := othersControlAdmin(r)
	var values *db.IdentityValues
	if err == nil {
		values, err = reset(controlContext.ControlDB, r.Context(), identity, controlContext.IdentityModel)
	}
	if herr := apiutils.ErrorFor(err, "control admins"); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	credentials := adminCredentials(identity, values)
	w.Header().Set("Cache-Control", "no-store")
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, credentials,
		control_admins_templ.Credentials(controlContext.ControlDB.ServerConfig.LinkURL("control"), credentials))
}

// othersControlAdmin is the admin in the URL, for an admin to manage: never the one signed in, so control always has
// an admin who can sign in
func othersControlAdmin(r *http.Request) (*controldb.ControlIdentity, error) {
	controlContext := httputils.GetControlContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	if id == controlContext.IdentityModel.ID {
		return nil, errYourself
	}
	return controlContext.ControlDB.Queries.GetControlIdentityByID(r.Context(), controldb.GetControlIdentityByIDParams{ID: id})
}

// adminForm is the admin to add, or the role to give one, from JSON or the page's form
func adminForm(r *http.Request) (*controlAdminForm, error) {
	form, herr := apiutils.ExtractModelFromRequest[controlAdminForm](httputils.GetContentType(r), r)
	if herr != nil {
		return nil, herr
	}
	if !slices.Contains(control_admins_templ.Roles, form.Role) {
		return nil, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.admin_role"))
	}
	return form, nil
}

// adminError is the response to an error managing an admin, saying when it was refused as control's last admin
func adminError(err error) *apiutils.HTTPError {
	if errors.Is(err, db.ErrLastControlAdmin) {
		return errLastAdmin
	}
	return apiutils.ErrorFor(err, "control admins")
}

// adminCredentials are an admin's new password, TOTP secret or both, to show once
func adminCredentials(identity *controldb.ControlIdentity, values *db.IdentityValues) *controlmodels.SiteIdentityCredentials {
	credentials := &controlmodels.SiteIdentityCredentials{Name: identity.Name, TotpSecret: values.TotpSecret}
	if values.Password != "" {
		credentials.Password = &values.Password
	}
	return credentials
}
