package can

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// SeeDeviceTokens is whether the identity is given the device tokens Loop and Trio send in their settings, which,
// with the app's push credentials, push to the phone: only one that can change the site's data, as LoopFollow sending
// remote commands must, not a read only follower, a clinician, or no identity, as a broadcast to every socket has
func SeeDeviceTokens(identity *controldb.SiteIdentity) bool {
	if identity == nil {
		return false
	}
	switch identity.Role {
	case consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite:
		return true
	default:
		return false
	}
}
