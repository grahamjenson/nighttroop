package controlmodels

// ServerSettings are the server's settings a control admin changes
type ServerSettings struct {
	// PublicSiteCreation is whether the public can create sites from the home page; a control admin always can
	PublicSiteCreation bool `json:"public_site_creation"`
}
