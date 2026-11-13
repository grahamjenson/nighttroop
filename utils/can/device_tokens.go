package can

import (
	"slices"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// DeviceTokenRoles can change the site's data, so are given Loop's and Trio's device tokens and may register them
var DeviceTokenRoles = []consts.IdentityRole{consts.IdentityRoleAdmin, consts.IdentityRoleUser,
	consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite}

// SeeDeviceTokens is whether the identity is given the device tokens Loop and Trio send in their settings, which,
// with the app's push credentials, push to the phone: only one that can change the site's data, as LoopFollow sending
// remote commands must, not a read only follower, a clinician, or no identity, as a broadcast to every socket has
func SeeDeviceTokens(identity *controldb.SiteIdentity) bool {
	return identity != nil && slices.Contains(DeviceTokenRoles, identity.Role)
}
