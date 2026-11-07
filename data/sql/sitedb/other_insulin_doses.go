package sitedb

import "nighttroop.t1d.nz/data/types"

// InsulinDoseOtherJSON is a dose's other attributes as JSON for an update, as the update merges them into those kept
func InsulinDoseOtherJSON(attributes types.InsulinDoseOther) *string {
	return types.OtherJSON(attributes)
}
