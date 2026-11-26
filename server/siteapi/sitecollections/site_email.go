package sitecollections

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	sendmail "nighttroop.t1d.nz/utils/mail"
	"nighttroop.t1d.nz/utils/security"
)

// A site's admin may give the site an email, to be told about the site by. It is theirs once they open the link sent
// to it. The link's key is kept here, in memory, so a restart of the server means sending the link again.
const (
	// emailLinkLasts is how long a verification link works, and emailEvery how often a site may be sent one, so a
	// site cannot be used to send emails to an address that did not ask for them
	emailLinkLasts = 24 * time.Hour
	emailEvery     = time.Hour
	// maxEmailLength is the longest address an email standard allows
	maxEmailLength = 254
)

// emailLink is the verification link last sent for a site
type emailLink struct {
	// keyHash is the SHA-256 of the link's key, and emailEnc the email it was sent to, as the site's row has it
	keyHash  [sha256.Size]byte
	emailEnc string
	sentAt   time.Time
}

// emailLinks are the link last sent for each site, by site ID: at most one a site, and one a site an hour
var emailLinks = struct {
	sync.Mutex
	bySite map[string]emailLink
}{bySite: map[string]emailLink{}}

// SiteEmailFor is the site's email as its admin's My Account page shows it, or nil when the page has no email card:
// the server sends no email, or the identity is not an admin
func SiteEmailFor(siteContext *httputils.SiteContext) *site_settings_templ.SiteEmail {
	if !siteContext.Config.EmailWorking() || siteContext.IdentityModel == nil || !can.IdentityAdministerSite(siteContext.IdentityModel.Role) {
		return nil
	}
	shown := &site_settings_templ.SiteEmail{}
	if site := siteContext.SiteModel; site != nil && site.EmailEnc != nil {
		email, err := siteContext.Config.Security.DecryptFor(security.PurposeSiteEmail, siteContext.SiteID, *site.EmailEnc)
		if err != nil {
			slog.Error("cannot read a site's email", "site_id", siteContext.SiteID, "error", err)
			return shown
		}
		shown.Email, shown.Verified = email, site.EmailVerifiedAt != nil
	}
	return shown
}

// emailAdmin is the context of a request to change the site's email, or nil after refusing it: only an admin may, and
// only where the server sends email
func emailAdmin(w http.ResponseWriter, r *http.Request) *httputils.SiteContext {
	siteContext := httputils.GetSiteContext(r.Context())
	switch {
	case siteContext.IdentityModel == nil || !can.IdentityAdministerSite(siteContext.IdentityModel.Role):
		writeError(w, r, siteContext.AcceptType, apiutils.ErrForbidden)
	case !siteContext.Config.EmailWorking():
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.email_off")))
	default:
		return siteContext
	}
	return nil
}

// SaveSiteEmailHandler keeps the email sent as the site's, not yet verified, and sends it a link that verifies it.
// A site is sent one link an hour.
func SaveSiteEmailHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := emailAdmin(w, r)
	if siteContext == nil {
		return
	}
	form, herr := apiutils.ExtractModelFromRequest[struct {
		Email string `json:"email"`
	}](siteContext.ContentType, r)
	if herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}
	email := strings.TrimSpace(form.Email)
	if address, err := mail.ParseAddress(email); err != nil || address.Address != email || len(email) > maxEmailLength {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrBadBodyWithHint(i18n.NewError("errors.site.email_bad")))
		return
	}

	emailLinks.Lock()
	last, sent := emailLinks.bySite[siteContext.SiteID]
	emailLinks.Unlock()
	if wait := emailEvery - time.Since(last.sentAt); sent && wait > 0 {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(int(wait.Seconds())+1))
		return
	}

	key := make([]byte, 32)
	_, _ = rand.Read(key)
	emailEnc, err := siteContext.Config.Security.EncryptFor(security.PurposeSiteEmail, siteContext.SiteID, email)
	if err == nil {
		err = siteContext.ControlDB.Queries.UpdateSiteEmail(r.Context(), controldb.UpdateSiteEmailParams{ID: siteContext.SiteID, EmailEnc: &emailEnc})
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot keep a site's email", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	siteContext.SiteModel.EmailEnc, siteContext.SiteModel.EmailVerifiedAt = &emailEnc, nil

	link := siteContext.SiteDB.URL() + "/settings/account/email/verify?key=" + hex.EncodeToString(key)
	mailer := sendmail.New(siteContext.Config.ResendAPIKey, siteContext.Config.ResendFrom, siteContext.Config.ResendAPIURL)
	err = mailer.Send(r.Context(), email, i18n.T(r.Context(), "site.settings.account.email.verify_subject"),
		i18n.T(r.Context(), "site.settings.account.email.verify_text", i18n.Data{"Site": siteContext.SiteDB.URL(), "Link": link}))
	if err != nil {
		// The address is not logged. The email stays, not verified, and the link can be sent again at once.
		slog.ErrorContext(r.Context(), "cannot send a site's verification email", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.site.email_not_sent")))
		return
	}
	emailLinks.Lock()
	emailLinks.bySite[siteContext.SiteID] = emailLink{keyHash: sha256.Sum256(key), emailEnc: emailEnc, sentAt: time.Now()}
	emailLinks.Unlock()

	shown := &site_settings_templ.SiteEmail{Email: email}
	WriteResponse(w, r, siteContext, http.StatusOK, shown, site_settings_templ.EmailSaved(shown), false)
}

// RemoveSiteEmailHandler takes the site's email away, and the link sent for it no longer works
func RemoveSiteEmailHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := emailAdmin(w, r)
	if siteContext == nil {
		return
	}
	if err := siteContext.ControlDB.Queries.UpdateSiteEmail(r.Context(), controldb.UpdateSiteEmailParams{ID: siteContext.SiteID}); err != nil {
		slog.ErrorContext(r.Context(), "cannot remove a site's email", "error", err)
		writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
		return
	}
	// The link goes, but when it was sent stays, so removing an email does not let the site send another at once
	emailLinks.Lock()
	if last, ok := emailLinks.bySite[siteContext.SiteID]; ok {
		emailLinks.bySite[siteContext.SiteID] = emailLink{sentAt: last.sentAt}
	}
	emailLinks.Unlock()
	shown := &site_settings_templ.SiteEmail{}
	WriteResponse(w, r, siteContext, http.StatusOK, shown, site_settings_templ.EmailRemoved(shown), false)
}

// VerifySiteEmailHandler is the link sent to a site's email: with the link's key, within a day, it verifies the email
// the link was sent to, once, and opens My Account. It needs no sign in, as whoever has the link has the email.
func VerifySiteEmailHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := httputils.GetSiteContext(r.Context())
	key, err := hex.DecodeString(r.URL.Query().Get("key"))
	keyHash := sha256.Sum256(key)

	emailLinks.Lock()
	last, sent := emailLinks.bySite[siteContext.SiteID]
	valid := err == nil && sent && last.emailEnc != "" && time.Since(last.sentAt) < emailLinkLasts &&
		subtle.ConstantTimeCompare(keyHash[:], last.keyHash[:]) == 1
	if valid {
		emailLinks.bySite[siteContext.SiteID] = emailLink{sentAt: last.sentAt}
	}
	emailLinks.Unlock()

	var verified int64
	if valid {
		verified, err = siteContext.ControlDB.Queries.VerifySiteEmail(r.Context(), controldb.VerifySiteEmailParams{ID: siteContext.SiteID, EmailEnc: &last.emailEnc})
	}
	if !valid || err != nil || verified != 1 {
		http.Error(w, i18n.T(r.Context(), "errors.site.email_link"), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/settings/account", http.StatusSeeOther)
}
