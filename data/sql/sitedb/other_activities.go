package sitedb

import "nighttroop.t1d.nz/data/types"

// ActivityOtherJSON is an activity's other attributes as JSON for an update, as the update merges them into those kept
func ActivityOtherJSON(attributes types.ActivityOther) *string {
	return types.OtherJSON(attributes)
}
