package controlmodels

// SiteEmails are the pre-built emails a control admin sends to a site's verified email, by name. An admin picks one
// and writes none, so a site is only sent text that is in the locale files.
var SiteEmails = []string{"archive", "storage", "contact"}

// SiteEmailSent is that one of the pre-built emails was sent to a site
type SiteEmailSent struct {
	SiteID string `json:"site_id"`
	Email  string `json:"email"`
}

// SentEmail is an email the server sent, as Resend has it: when, its subject, the last that happened to it, the site
// it was to when that is known, and its address with most of it hidden
type SentEmail struct {
	ID      string  `json:"id"`
	SentAt  int64   `json:"sent_at"`
	Subject string  `json:"subject"`
	Status  string  `json:"status"`
	SiteID  *string `json:"site_id"`
	To      string  `json:"to"`
}

// SentEmailAddress is the whole address a sent email was to, which an admin asks for
type SentEmailAddress struct {
	To string `json:"to"`
}
