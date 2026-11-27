package controlapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/mail"
	"nighttroop.t1d.nz/utils/security"
)

// A control admin may give an email of their own on their Account page. It is theirs once they open the link sent to
// it, which works for a day. Each admin is sent one an hour, by their ID.
var accountEmailLinks = mail.NewLinks(24*time.Hour, time.Hour)

// accountEmail is the admin's email as their Account page shows it, or nil when the page has no email card, as the
// server sends no email
func accountEmail(cfg *config.ServerConfig, identity *controldb.ControlIdentity) *site_settings_templ.SiteEmail {
	if !cfg.EmailWorking() {
		return nil
	}
	shown := &site_settings_templ.SiteEmail{Control: true}
	if identity.EmailEnc != nil {
		email, err := cfg.Security.DecryptFor(security.PurposeControlEmail, "", *identity.EmailEnc)
		if err != nil {
			slog.Error("cannot read a control admin's email", "control_identity_id", identity.ID, "error", err)
			return shown
		}
		shown.Email, shown.Verified = email, identity.EmailVerifiedAt != nil
	}
	return shown
}

// POSTAccountEmail keeps the email sent as the admin's, not yet verified, and sends it a link that verifies it
func POSTAccountEmail(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, acceptType := httputils.GetControlContext(r.Context()).IdentityModel, httputils.GetAcceptType(r)
		owner := strconv.FormatInt(identity.ID, 10)
		form, herr := apiutils.ExtractModelFromRequest[struct {
			Email string `json:"email"`
		}](httputils.GetContentType(r), r)
		email := strings.TrimSpace(form.Email)
		switch {
		case !cfg.EmailWorking():
			herr = apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.email_off"))
		case herr != nil:
		case !mail.Address(email):
			herr = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.site.email_bad"))
		case accountEmailLinks.Wait(owner) > 0:
			herr = apiutils.ErrRateLimitExceeded(int(accountEmailLinks.Wait(owner).Seconds()) + 1)
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}

		key := mail.Key()
		emailEnc, err := cfg.Security.EncryptFor(security.PurposeControlEmail, "", email)
		if err == nil {
			err = controlDB.Queries.UpdateControlIdentityEmail(r.Context(), controldb.UpdateControlIdentityEmailParams{ID: identity.ID, EmailEnc: &emailEnc})
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "cannot keep a control admin's email", "error", err)
			apiutils.ErrInternalServerError.Write(acceptType, w)
			return
		}
		link := controlOrigin(cfg) + "/account/email/verify?id=" + owner + "&key=" + key
		err = mail.New(cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL).Send(r.Context(), email, i18n.T(r.Context(), "control.account.email.verify_subject"),
			i18n.T(r.Context(), "control.account.email.verify_text", i18n.Data{"Name": identity.Name, "Control": controlOrigin(cfg), "Link": link}))
		if err != nil {
			// The address is not logged. The email stays, not verified, and the link can be sent again at once.
			slog.ErrorContext(r.Context(), "cannot send a control admin's verification email", "error", err)
			apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.email_not_sent")).Write(acceptType, w)
			return
		}
		accountEmailLinks.Sent(owner, key, emailEnc)
		shown := &site_settings_templ.SiteEmail{Email: email, Control: true}
		controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, shown, site_settings_templ.EmailSaved(shown))
	}
}

// DELETEAccountEmail takes the admin's email away, and the link sent for it no longer works
func DELETEAccountEmail(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, acceptType := httputils.GetControlContext(r.Context()).IdentityModel, httputils.GetAcceptType(r)
		if !cfg.EmailWorking() {
			apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.email_off")).Write(acceptType, w)
			return
		}
		if err := controlDB.Queries.UpdateControlIdentityEmail(r.Context(), controldb.UpdateControlIdentityEmailParams{ID: identity.ID}); err != nil {
			slog.ErrorContext(r.Context(), "cannot remove a control admin's email", "error", err)
			apiutils.ErrInternalServerError.Write(acceptType, w)
			return
		}
		accountEmailLinks.Forget(strconv.FormatInt(identity.ID, 10))
		shown := &site_settings_templ.SiteEmail{Control: true}
		controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, shown, site_settings_templ.EmailRemoved(shown))
	}
}

// GETVerifyAccountEmail is the link sent to an admin's email: with the link's key, within a day, it verifies the email
// the link was sent to, once, and opens Account. It needs no sign in, as whoever has the link has the email.
func GETVerifyAccountEmail(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		emailEnc, valid := accountEmailLinks.Take(strconv.FormatInt(id, 10), r.URL.Query().Get("key"))
		var verified int64
		if err == nil && valid {
			verified, err = controlDB.Queries.VerifyControlIdentityEmail(r.Context(), controldb.VerifyControlIdentityEmailParams{ID: id, EmailEnc: &emailEnc})
		}
		if err != nil || !valid || verified != 1 {
			http.Error(w, i18n.T(r.Context(), "errors.control.email_link"), http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}
