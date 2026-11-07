package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// ActivityOther is what a fitness tracker or watch measured of an activity, kept in its other attributes. It is written
// out, not generated, as a generated converter cannot be called from this package's own code, which goverter compiles
// first.
func ActivityOther(source Activity) datatypes.ActivityOther {
	return datatypes.ActivityOther{Calories: source.Calories.Int64p(), Steps: source.Steps.Int64p(),
		HeartRate: source.HeartRate.Int64p()}
}

// ActivityOtherJSON is what a fitness tracker or watch measured of an activity, as JSON for an update to merge into its
// other attributes
func ActivityOtherJSON(source Activity) *string {
	return datatypes.OtherJSON(ActivityOther(source))
}

// TreatmentActivityOther is what a fitness tracker or watch measured of a treatment's activity, kept in its other
// attributes
func TreatmentActivityOther(source Treatment) datatypes.ActivityOther {
	return datatypes.ActivityOther{Calories: source.Calories.Int64p(), Steps: source.Steps.Int64p(),
		HeartRate: source.HeartRate.Int64p()}
}
