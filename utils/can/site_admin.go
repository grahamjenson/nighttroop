package can

// site_admin.go: what only a site admin may do outside the stores of the site API (site_stores.go has the stores).

import (
	"nighttroop.t1d.nz/utils/consts"
)

// IdentityAdministerSite is whether the role administers the site: only an admin sets up where the data comes from and
// who can use the site, and sees its errors page. The Nightscout subjects show this role to AAPS as "admin".
func IdentityAdministerSite(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// IdentityReadSiteSettings is whether the role reads the settings of the site: only an admin, who changes them. Other
// roles see only their own settings, which inherit the site's settings.
func IdentityReadSiteSettings(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// IdentityRegenerateSummaries is whether the role tells the server to build the summaries of the site again: only an
// admin, as a build reads a year of the site's data.
func IdentityRegenerateSummaries(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// IdentitySeeAllPushRegistrations is whether the role sees, tests and deletes the push registrations of all identities
// of the site: only an admin. Other roles see only their own push registrations.
func IdentitySeeAllPushRegistrations(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// IdentityResetOthersCredentials is whether the role resets the password or TOTP of a different identity: only an
// admin. All identities can reset their own.
func IdentityResetOthersCredentials(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}
