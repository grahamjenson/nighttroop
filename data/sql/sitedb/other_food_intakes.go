package sitedb

import "nighttroop.t1d.nz/data/types"

// FoodIntakeOtherJSON is a food intake's other attributes as JSON for an update, as the update merges them into those
// kept
func FoodIntakeOtherJSON(attributes types.FoodIntakeOther) *string {
	return types.OtherJSON(attributes)
}
