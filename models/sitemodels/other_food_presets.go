package sitemodels

import "nighttroop.t1d.nz/data/types"

// FoodPresetOther is a food preset's fields beyond its carbs, kept in its other attributes. It is written out, as a
// generated converter cannot be called from this package's own code.
func FoodPresetOther(source *SiteFoodPreset) types.FoodPresetOther {
	return types.FoodPresetOther{Category: source.Category, Subcategory: source.Subcategory, Fat: source.Fat,
		Protein: source.Protein, Fibre: source.Fibre, Energy: source.Energy, Portions: source.Portions,
		PreBolus: source.PreBolus}
}

// FoodPresetOtherJSON is a food preset's fields beyond its carbs, as JSON for an update to merge into its other
// attributes
func FoodPresetOtherJSON(source *SiteFoodPreset) *string {
	return types.OtherJSON(FoodPresetOther(source))
}
