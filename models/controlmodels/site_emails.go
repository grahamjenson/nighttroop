package controlmodels

// SiteEmails are the pre-built emails a control admin sends to a site's verified email, by name. An admin picks one
// and writes none, so a site is only sent text that is in the locale files.
var SiteEmails = []string{"archive", "storage", "contact"}

// SiteEmailSent is that one of the pre-built emails was sent to a site
type SiteEmailSent struct {
	SiteID string `json:"site_id"`
	Email  string `json:"email"`
}
