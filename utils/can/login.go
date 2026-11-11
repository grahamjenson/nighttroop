package can

import (
	"nighttroop.t1d.nz/utils/consts"
)

func LoginToWebsite(role consts.IdentityRole) bool {
	switch role {
	case consts.IdentityRoleAdmin:
		return true
	case consts.IdentityRoleUser, consts.IdentityRoleClinician:
		return true
	default:
		return false
	}
}

func MustLoginWithTotp(role consts.IdentityRole, hasTotpSecret bool) bool {
	if hasTotpSecret {
		return true
	}
	switch role {
	case consts.IdentityRoleAdmin:
		return true
	case consts.IdentityRoleUser:
		return false
	default:
		return false
	}
}

// HaveNightscoutToken is whether the role is given a Nightscout token: not an admin, nor a clinician, who sees only the
// printable report
func HaveNightscoutToken(role consts.IdentityRole) bool {
	switch role {
	case consts.IdentityRoleAdmin, consts.IdentityRoleClinician:
		return false
	default:
		return true
	}
}
