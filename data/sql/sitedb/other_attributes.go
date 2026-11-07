package sitedb

import "nighttroop.t1d.nz/data/types"

// GlucoseReadingOtherJSON is a reading's other attributes as JSON for an update, as the update merges them into those
// kept
func GlucoseReadingOtherJSON(attributes types.GlucoseReadingOther) *string {
	return types.OtherJSON(attributes)
}
