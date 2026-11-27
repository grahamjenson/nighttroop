package controlapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_sent_emails_templ"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/mail"
	"nighttroop.t1d.nz/utils/security"
)

// sentEmailsShown is how many of the last emails the page lists
const sentEmailsShown = 50

// sentAtLayouts are the ways Resend writes when an email was made
var sentAtLayouts = []string{"2006-01-02 15:04:05.999999999-07", "2006-01-02 15:04:05.999999999-07:00", time.RFC3339Nano}

// SentEmailsHandler is the emails the server last sent, as Resend has them, the newest first: the verification
// links and the pre-built emails. Each has its site, when a live site has its address, and its address mostly hidden.
func SentEmailsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	cfg, emails := controlContext.ControlDB.ServerConfig, []*controlmodels.SentEmail{}
	if cfg.EmailWorking() {
		sent, err := mailerOf(cfg).List(r.Context(), sentEmailsShown)
		if err != nil {
			slog.ErrorContext(r.Context(), "cannot read the sent emails", "error", err)
			controlcollections.WriteError(w, r, acceptType, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.emails_not_read")))
			return
		}
		sites := sitesByEmail(r.Context(), controlContext)
		for _, email := range sent {
			model := &controlmodels.SentEmail{ID: email.ID, Subject: email.Subject, Status: email.LastEvent}
			for _, layout := range sentAtLayouts {
				if at, err := time.Parse(layout, email.CreatedAt); err == nil {
					model.SentAt = at.UnixMilli()
					break
				}
			}
			if len(email.To) > 0 {
				model.To = maskEmail(email.To[0])
				if siteID, ok := sites[strings.ToLower(email.To[0])]; ok {
					model.SiteID = &siteID
				}
			}
			emails = append(emails, model)
		}
	}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, emails, control_sent_emails_templ.SentEmails(emails, cfg.EmailWorking()))
}

// SentEmailAddressHandler is the whole address a sent email was to, read from Resend. Only an admin can ask, and
// that they did is logged, without the address.
func SentEmailAddressHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	id := chi.URLParam(r, "id")
	sent, err := mailerOf(controlContext.ControlDB.ServerConfig).Get(r.Context(), id)
	if err != nil || len(sent.To) == 0 {
		controlcollections.WriteError(w, r, acceptType, apiutils.ErrNotFound)
		return
	}
	slog.InfoContext(r.Context(), "a control admin read the address of a sent email", "by", controlContext.IdentityModel.Name, "email_id", id)
	address := &controlmodels.SentEmailAddress{To: sent.To[0]}
	controlcollections.WriteResponse(w, r, acceptType, http.StatusOK, address, control_sent_emails_templ.Address(address.To))
}

// mailerOf is the mailer of the server's Resend account
func mailerOf(cfg *config.ServerConfig) *mail.Mailer {
	return mail.New(cfg.ResendAPIKey, cfg.ResendFrom, cfg.ResendAPIURL)
}

// sitesByEmail is the ID of each live site by its email, in lower case
func sitesByEmail(ctx context.Context, controlContext *httputils.ControlContext) map[string]string {
	sites := map[string]string{}
	rows, err := controlContext.ControlDB.Queries.GetSiteEmails(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "cannot read the sites' emails", "error", err)
	}
	for _, row := range rows {
		if address, err := controlContext.ControlDB.ServerConfig.Security.DecryptFor(security.PurposeSiteEmail, row.ID, *row.EmailEnc); err == nil {
			sites[strings.ToLower(address)] = row.ID
		}
	}
	return sites
}

// maskEmail is an address with only the first letter of its name and of its domain, and what follows its last dot:
// sam@example.com is s•••@e•••.com
func maskEmail(address string) string {
	name, domain, found := strings.Cut(address, "@")
	if !found || name == "" || domain == "" {
		return "•••"
	}
	end := ""
	if dot := strings.LastIndex(domain, "."); dot > 0 {
		end = domain[dot:]
	}
	return name[:1] + "•••@" + domain[:1] + "•••" + end
}
