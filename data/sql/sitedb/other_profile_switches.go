package sitedb

import "nighttroop.t1d.nz/data/types"

// ProfileSwitchOtherJSON is a profile switch's other attributes as JSON for an update, as the update merges them into
// those kept
func ProfileSwitchOtherJSON(attributes types.ProfileSwitchOther) *string {
	return types.OtherJSON(attributes)
}
