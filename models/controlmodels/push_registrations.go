package controlmodels

import "nighttroop.t1d.nz/models/sitemodels"

// PushRegistration is an app on a device that a site's identity has pushes sent to, never with its token, for the
// control site, which sees every site's
type PushRegistration struct {
	*sitemodels.PushRegistration
	SiteID       string `json:"site_id"`
	IdentityName string `json:"identity_name"`
}
