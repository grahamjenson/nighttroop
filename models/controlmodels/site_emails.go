package controlmodels

// SiteEmails are the pre-built emails a control admin sends to a site's verified email, by name. An admin picks one
// and writes none, so a site is only sent text that is in the locale files.
var SiteEmails = []string{"archive", "storage", "contact"}

// VerifiedSite is a site whose admin verified an email. It has no address: control sends to it, and does not show it.
type VerifiedSite struct {
	SiteID          string `json:"site_id"`
	EmailVerifiedAt int64  `json:"email_verified_at"`
}

// SiteEmailSent is that one of the pre-built emails was sent to a site
type SiteEmailSent struct {
	SiteID string `json:"site_id"`
	Email  string `json:"email"`
}
