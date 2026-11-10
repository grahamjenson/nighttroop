package controlmodels

import "nighttroop.t1d.nz/models/sitemodels"

// PushCredential is a site's push credential, never with its secret, or with no SiteID one for every site, for the
// control site, which sees them all
type PushCredential struct {
	*sitemodels.PushCredential
	SiteID *string `json:"site_id"`
}
