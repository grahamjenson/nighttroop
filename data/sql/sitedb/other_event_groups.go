package sitedb

import "nighttroop.t1d.nz/data/types"

// EventGroupOtherJSON is a treatment's other attributes as JSON for an update, as the update merges them into those kept
func EventGroupOtherJSON(attributes types.EventGroupOther) *string {
	return types.OtherJSON(attributes)
}
