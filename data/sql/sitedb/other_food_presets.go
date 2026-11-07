package sitedb

import "nighttroop.t1d.nz/data/types"

// FoodPresetOtherJSON is a food preset's other attributes as JSON for an update, as the update merges them into those
// kept
func FoodPresetOtherJSON(attributes types.FoodPresetOther) *string {
	return types.OtherJSON(attributes)
}
