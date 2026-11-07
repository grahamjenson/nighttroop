package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// FoodOther is a food's fields beyond its carbs, kept in its food preset's other attributes. It is written out, not
// generated, as a generated converter cannot be called from this package's own code, which goverter compiles first.
func FoodOther(source Food) datatypes.FoodPresetOther {
	return datatypes.FoodPresetOther{Category: source.Category.Stringp(), Subcategory: source.Subcategory.Stringp(),
		Fat: source.Fat.Float64p(), Protein: source.Protein.Float64p(), Fibre: source.Fibre.Float64p(),
		Energy: source.Energy.Float64p(), Portions: source.Portions.Float64p(), PreBolus: source.PreBolus.Float64p()}
}

// FoodOtherJSON is a food's fields beyond its carbs, as JSON for an update to merge into its food preset's other
// attributes
func FoodOtherJSON(source Food) *string {
	return datatypes.OtherJSON(FoodOther(source))
}
