package can

// login.go: who may sign in to the website, who must give a TOTP code to sign in, and who may give themselves new
// credentials.

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

// IdentityResetOwnCredentials is whether the role gives itself a new password or TOTP secret: an admin and a user, who
// sign in to the website and have its settings page. A device or a read only user has only its Nightscout token, which
// its followers' apps share, so one holder of it would end the token for the others and keep the new one: only an admin
// gives it a new one. A clinician sees only the printable report.
func IdentityResetOwnCredentials(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleAdmin || role == consts.IdentityRoleUser
}
