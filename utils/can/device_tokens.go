package can

// device_tokens.go: who gets the device tokens that Loop and Trio send in their settings. With the push credentials of
// the app, a device token sends remote commands to the PWD phone.

import (
	"slices"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// DeviceTokenRoles can change the site's data, so are given Loop's and Trio's device tokens
var DeviceTokenRoles = []consts.IdentityRole{consts.IdentityRoleAdmin, consts.IdentityRoleUser,
	consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceReadWrite}

// IdentitySeeDeviceTokens is whether the identity is given the device tokens Loop and Trio send in their settings, which,
// with the app's push credentials, push to the phone: only one that can change the site's data, as LoopFollow sending
// remote commands must, not a read only follower, a clinician, or no identity, as a broadcast to every socket has
func IdentitySeeDeviceTokens(identity *controldb.SiteIdentity) bool {
	return identity != nil && slices.Contains(DeviceTokenRoles, identity.Role)
}

// LoopRegistrationRoles can write the site's profiles, as Loop and Trio do, so may register either for its remote
// commands: not a Caregiver device, which could take the commands sent to the PWD phone or stop them
var LoopRegistrationRoles = []consts.IdentityRole{consts.IdentityRoleAdmin, consts.IdentityRoleUser,
	consts.IdentityRoleDevicePWDPhone}

// IdentityRegisterLoop is whether the identity may register Loop or Trio for its remote commands
func IdentityRegisterLoop(identity *controldb.SiteIdentity) bool {
	return identity != nil && slices.Contains(LoopRegistrationRoles, identity.Role)
}
