package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// TreatmentFoodIntakeOther is a treatment's food fields beyond its carbs, kept in its food intake's other attributes.
// It is written out, not generated, as a generated converter cannot be called from this package's own code, which
// goverter compiles first.
func TreatmentFoodIntakeOther(source Treatment) datatypes.FoodIntakeOther {
	return datatypes.FoodIntakeOther{Category: source.Category.Stringp(), Subcategory: source.Subcategory.Stringp(),
		Fat: source.Fat.Float64p(), Protein: source.Protein.Float64p(), Fibre: source.Fibre.Float64p(),
		Energy: source.Energy.Float64p(), Portions: source.Portions.Float64p(), PreBolus: source.PreBolus.Float64p()}
}
