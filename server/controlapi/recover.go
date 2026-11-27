package controlapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/control_templ/control_login_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/mail"
	"nighttroop.t1d.nz/utils/security"
)

// An admin who forgot their password is sent a link to their verified email, which gives them a new one. Each link
// works for an hour, and an admin is sent one an hour, by their ID. They still need their TOTP code to sign in.
var recoverLinks = mail.NewLinks(time.Hour, time.Hour)

// recoverLimit is the name the sign in limits count requests for a link under, longer than any admin's name can be
const recoverLimit = "recovering a password by an email"

// GETRecoverPage is where an admin types their verified email, on a server that sends email
func GETRecoverPage(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !cfg.EmailWorking() {
			http.NotFound(w, r)
			return
		}
		_ = control_login_templ.Recover().Render(r.Context(), w)
	}
}

// POSTRecover sends a link to recover the password of the admin whose verified email is the one given. It answers the
// same whether or not an admin has it, and sends it after answering, so it never tells whose email it is.
func POSTRecover(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		form, herr := apiutils.ExtractModelFromRequest[struct {
			Email string `json:"email"`
		}](httputils.GetContentType(r), r)
		email := strings.TrimSpace(form.Email)
		allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(cfg, r, recoverLimit)
		switch {
		case !cfg.EmailWorking():
			herr = apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.email_off"))
		case herr != nil:
		case !mail.Address(email):
			herr = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.site.email_bad"))
		case !allowed:
			herr = apiutils.ErrRateLimitExceeded(60)
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		if admin := verifiedAdmin(r.Context(), controlDB, cfg, email); admin != nil && recoverLinks.Wait(strconv.FormatInt(admin.ID, 10)) == 0 {
			owner, key := strconv.FormatInt(admin.ID, 10), mail.Key()
			recoverLinks.Sent(owner, key, *admin.EmailEnc)
			link := controlOrigin(cfg) + "/login/recover/password?id=" + owner + "&key=" + key
			subject := i18n.T(r.Context(), "control.login.recover.subject")
			text := i18n.T(r.Context(), "control.login.recover.text", i18n.Data{"Name": admin.Name, "Control": controlOrigin(cfg), "Link": link})
			go func(ctx context.Context) {
				if err := mail.New(cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL).Send(ctx, email, subject, text); err != nil {
					slog.ErrorContext(ctx, "cannot send a control admin's recover email", "control_identity_id", admin.ID, "error", err)
				}
			}(context.WithoutCancel(r.Context()))
		}
		controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, map[string]string{}, control_login_templ.RecoverSent())
	}
}

// verifiedAdmin is the admin whose verified email is email, exactly as typed, or nil. Control has few admins, so each
// is decrypted, as an encrypted email cannot be looked up.
func verifiedAdmin(ctx context.Context, controlDB *db.ControlDB, cfg *config.ServerConfig, email string) *controldb.ControlIdentity {
	admins, err := controlDB.Queries.AllControlIdentities(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "cannot list the control admins to recover a password", "error", err)
	}
	for _, admin := range admins {
		if admin.EmailEnc == nil || admin.EmailVerifiedAt == nil {
			continue
		}
		if verified, err := cfg.Security.DecryptFor(security.PurposeControlEmail, "", *admin.EmailEnc); err == nil && verified == email {
			return admin
		}
	}
	return nil
}

// GETNewPasswordPage is the link sent to recover a password. It changes nothing, so a mail scanner opening it does not
// use it up: its button does.
func GETNewPasswordPage(cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !cfg.EmailWorking() {
			http.NotFound(w, r)
			return
		}
		_ = control_login_templ.NewPassword(r.URL.Query().Get("id"), r.URL.Query().Get("key")).Render(r.Context(), w)
	}
}

// POSTNewPassword gives the admin of a recover link a new random password, shown once, if the link works and their
// email is still the verified one it was sent to. As when another admin resets it, their passkeys are removed.
func POSTNewPassword(controlDB *db.ControlDB, cfg *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acceptType := httputils.GetAcceptType(r)
		form, herr := apiutils.ExtractModelFromRequest[struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}](httputils.GetContentType(r), r)
		var admin *controldb.ControlIdentity
		var values *db.IdentityValues
		var err error
		if herr == nil {
			herr = apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.recover_link"))
			id, idErr := strconv.ParseInt(form.ID, 10, 64)
			emailEnc, valid := recoverLinks.Take(form.ID, form.Key)
			if idErr == nil && valid {
				admin, err = controlDB.Queries.GetControlIdentityByID(r.Context(), controldb.GetControlIdentityByIDParams{ID: id})
			}
			if admin != nil && err == nil && admin.EmailEnc != nil && *admin.EmailEnc == emailEnc && admin.EmailVerifiedAt != nil {
				values, err = controlDB.ResetControlIdentityPassword(r.Context(), admin, admin)
				herr = apiutils.ErrorFor(err, "control admins")
			}
		}
		if herr != nil {
			herr.Write(acceptType, w)
			return
		}
		credentials := adminCredentials(admin, values)
		w.Header().Set("Cache-Control", "no-store")
		controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, credentials, control_login_templ.Recovered(controlOrigin(cfg), credentials))
	}
}
