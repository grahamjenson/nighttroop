package can

import (
	"nighttroop.t1d.nz/utils/consts"
)

func LoginToWebsite(role consts.IdentityRole) bool {
	switch role {
	case consts.IdentityRoleAdmin:
		return true
	case consts.IdentityRoleUser:
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
