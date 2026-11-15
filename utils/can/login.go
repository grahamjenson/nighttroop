package can

// login.go: who may sign in to the website, and who must give a TOTP code to sign in.

import (
	"nighttroop.t1d.nz/utils/consts"
)

// IdentityLoginToWebsite is whether the role signs in to the website: the people, admins, users and clinicians. A
// device and a read only user use only their token.
func IdentityLoginToWebsite(role consts.IdentityRole) bool {
	switch role {
	case consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleClinician:
		return true
	default:
		return false
	}
}

// IdentityMustLoginWithTotp is whether the identity must give a TOTP code to sign in: an admin always, as it
// administers the site, and any identity with a TOTP secret. The server makes a TOTP secret for a new identity if this
// is true without one.
func IdentityMustLoginWithTotp(role consts.IdentityRole, hasTotpSecret bool) bool {
	return hasTotpSecret || role == consts.IdentityRoleAdmin
}
