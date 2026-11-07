package sitemodels

import "nighttroop.t1d.nz/data/types"

// ActivityOther is what a fitness tracker or watch measured of an activity, kept in its other attributes. It is written
// out, as a generated converter cannot be called from this package's own code.
func ActivityOther(source *SiteActivity) types.ActivityOther {
	return types.ActivityOther{Calories: source.Calories, Steps: source.Steps, HeartRate: source.HeartRate}
}

// ActivityOtherJSON is what a fitness tracker or watch measured of an activity, as JSON for an update to merge into its
// other attributes
func ActivityOtherJSON(source *SiteActivity) *string {
	return types.OtherJSON(ActivityOther(source))
}
