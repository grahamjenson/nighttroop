package can

// nightscout.go: the Nightscout API. Who has a Nightscout token and may sign in with one, the Nightscout permissions of
// each role, and who pairs AAPS clients as Nightscout subjects.

import (
	"fmt"
	"slices"

	"nighttroop.t1d.nz/utils/consts"
)

// CaregiverNightscoutPermissions are a read and write device's, the token of a carer's or follower's app such as
// LoopFollow or Nightguard: it reads everything and writes treatments, as Nightscout's careportal does, a device status
// and settings. It cannot add, change or delete the readings, profiles or device statuses that the PWD phone uploads
var CaregiverNightscoutPermissions = []string{"api:*:read", "api:treatments:create", "api:treatments:update",
	"api:treatments:delete", "api:devicestatus:create", "api:settings:*"}

// aapsClientNightscoutPermissions are a device AAPS pairs as a client, by its QR code: it reads everything and writes
// only the settings that client control is made of, as its role does in Nightscout. Whoever sees the code can read the
// site
var aapsClientNightscoutPermissions = []string{"api:*:read", "api:settings:create", "api:settings:update"}

// NightscoutPermissions are the Nightscout permissions of the role and the client it was added for, e.g. "api:*:read".
// A role with none (an admin or a clinician) cannot use the Nightscout API, so it is an error.
//
// The Nightscout collections are entries, treatments, devicestatus, profile, settings, food and properties. A
// permission is "api:<collection>:<read|create|update|delete|admin>", and "*" matches any part.
func NightscoutPermissions(role consts.IdentityRole, client *string) ([]string, error) {
	var permissions []string
	switch role {
	case consts.IdentityRoleUserReadOnly, consts.IdentityRoleDeviceRead:
		permissions = []string{"api:*:read"}
	case consts.IdentityRoleUser, consts.IdentityRoleDevicePWDPhone:
		permissions = []string{"api:*:*"}
	case consts.IdentityRoleDeviceReadWrite:
		permissions = slices.Clone(CaregiverNightscoutPermissions)
	default:
		return nil, fmt.Errorf("unknown identity role: %s", role)
	}
	if client != nil && *client == consts.AAPSClient {
		permissions = slices.Clone(aapsClientNightscoutPermissions)
	}
	return permissions, nil
}

// IdentityHaveNightscoutToken is whether the role is given a Nightscout token: a user, a read only user and each device,
// not an admin, nor a clinician, who sees only the printable report. A role it does not know has none.
func IdentityHaveNightscoutToken(role consts.IdentityRole) bool {
	switch role {
	case consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly, consts.IdentityRoleDevicePWDPhone,
		consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite:
		return true
	default:
		return false
	}
}

// IdentityUseNightscoutToken is whether the identity may sign in with a Nightscout token (or its SHA1): not an admin,
// and not an identity with a TOTP secret, as a token has no TOTP code
func IdentityUseNightscoutToken(role consts.IdentityRole, hasTotpSecret bool) bool {
	return !hasTotpSecret && role != consts.IdentityRoleAdmin
}

// IdentityAuthorizeDevices is whether the role adds, lists and deletes the Nightscout subjects that AAPS pairs its
// clients with: only the PWD phone, the AAPS that pairs them
func IdentityAuthorizeDevices(role consts.IdentityRole) bool {
	return role == consts.IdentityRoleDevicePWDPhone
}
