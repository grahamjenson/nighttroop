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

// IdentitySeeAllPushRegistrations is whether the role sees, tests and deletes the push registrations of all identities
// of the site: only an admin. Other roles see only their own push registrations.
func IdentitySeeAllPushRegistrations(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// IdentityResetOthersCredentials is whether the role resets the password or TOTP of a different identity: only an
// admin. IdentityResetOwnCredentials (login.go) is who resets their own.
func IdentityResetOthersCredentials(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin
}

// ClientRoles are the roles an admin must give a device for a client, and cannot change: a loop (AAPS, Trio or Loop)
// uploads the treatments it gives and reads the profile it loops with, so it is the PWD phone
var ClientRoles = map[string]consts.IdentityRole{
	"aaps": consts.IdentityRoleDevicePWDPhone, "trio": consts.IdentityRoleDevicePWDPhone, "loop": consts.IdentityRoleDevicePWDPhone,
}
