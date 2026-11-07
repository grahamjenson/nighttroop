package sitedb

import "nighttroop.t1d.nz/data/types"

// ComponentEventOtherJSON is a component event's other attributes as JSON for an update, as the update merges them into
// those kept
func ComponentEventOtherJSON(attributes types.ComponentEventOther) *string {
	return types.OtherJSON(attributes)
}
