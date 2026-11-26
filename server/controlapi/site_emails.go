package controlapi

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sites_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/mail"
	"nighttroop.t1d.nz/utils/security"
)

// siteEmailEvery is how often a site may be sent an email by control, so a site's admin is not sent many
const siteEmailEvery = 24 * time.Hour

// siteEmailed is when each site was last sent an email by control, by site ID. It is in memory, as the keys of the
// verification links are, so a restart of the server forgets it.
var siteEmailed = struct {
	sync.Mutex
	at map[string]time.Time
}{at: map[string]time.Time{}}

// siteEmailTexts are the ids of each pre-built email's subject and text
var siteEmailTexts = map[string][2]string{
	"archive": {"control.site_emails.archive.subject", "control.site_emails.archive.text"},
	"storage": {"control.site_emails.storage.subject", "control.site_emails.storage.text"},
	"contact": {"control.site_emails.contact.subject", "control.site_emails.contact.text"},
}

// SendSiteEmailHandler sends one of the pre-built emails to a site's verified email. Only admins can, only where the
// server sends email, and a site is sent one a day. The admin gives the email's name and no text.
func SendSiteEmailHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	cfg, siteID := controlContext.ControlDB.ServerConfig, chi.URLParam(r, "id")
	form, herr := apiutils.ExtractModelFromRequest[struct {
		Email string `json:"email"`
	}](httputils.GetContentType(r), r)
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	texts, known := siteEmailTexts[form.Email]
	switch {
	case !cfg.EmailWorking():
		apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.email_off")).Write(acceptType, w)
		return
	case !known:
		apiutils.ErrBadBodyWithHint(i18n.NewError("errors.control.email_which")).Write(acceptType, w)
		return
	}
	site, err := controlContext.ControlDB.Queries.GetSite(r.Context(), controldb.GetSiteParams{ID: siteID})
	if err != nil || site.EmailEnc == nil || site.EmailVerifiedAt == nil {
		apiutils.ErrNotFound.Write(acceptType, w)
		return
	}

	// The time is taken before the email is sent, so two requests at once send one email
	siteEmailed.Lock()
	last, sent := siteEmailed.at[siteID]
	wait := siteEmailEvery - time.Since(last)
	if !sent || wait <= 0 {
		siteEmailed.at[siteID] = time.Now()
	}
	siteEmailed.Unlock()
	if sent && wait > 0 {
		apiutils.ErrRateLimitExceeded(int(wait.Seconds()) + 1).Write(acceptType, w)
		return
	}

	address, err := cfg.Security.DecryptFor(security.PurposeSiteEmail, siteID, *site.EmailEnc)
	if err == nil {
		err = mail.New(cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL).Send(r.Context(), address,
			i18n.T(r.Context(), texts[0]), i18n.T(r.Context(), texts[1], i18n.Data{"Site": cfg.SiteURL(siteID)}))
	}
	if err != nil {
		// The address is not logged. An email that was not sent does not count, so it can be sent again at once.
		slog.ErrorContext(r.Context(), "cannot send an email to a site", "site_id", siteID, "error", err)
		siteEmailed.Lock()
		if sent {
			siteEmailed.at[siteID] = last
		} else {
			delete(siteEmailed.at, siteID)
		}
		siteEmailed.Unlock()
		apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.email_not_sent")).Write(acceptType, w)
		return
	}
	controlContext.ControlDB.Logs.SiteEmailedEvent(r.Context(), siteID, controlContext.IdentityModel.Name, form.Email)
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, &controlmodels.SiteEmailSent{SiteID: siteID, Email: form.Email}, control_sites_templ.SiteEmailSent())
}
