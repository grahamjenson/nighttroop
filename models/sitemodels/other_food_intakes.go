package sitemodels

import "nighttroop.t1d.nz/data/types"

// FoodIntakeOther is a food intake's fields beyond its carbs, kept in its other attributes. It is written out, as a
// generated converter cannot be called from this package's own code.
func FoodIntakeOther(source *SiteFoodIntake) types.FoodIntakeOther {
	return types.FoodIntakeOther{Category: source.Category, Subcategory: source.Subcategory, Fat: source.Fat,
		Protein: source.Protein, Fibre: source.Fibre, Energy: source.Energy, Portions: source.Portions,
		PreBolus: source.PreBolus}
}

// FoodIntakeOtherJSON is a food intake's fields beyond its carbs, as JSON for an update to merge into its other
// attributes
func FoodIntakeOtherJSON(source *SiteFoodIntake) *string {
	return types.OtherJSON(FoodIntakeOther(source))
}
