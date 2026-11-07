package sitemodels

import "nighttroop.t1d.nz/data/types"

// InsulinDoseOther is a dose's combo bolus split the site sends, kept in its other attributes. Only a loop's upload
// says whether a dose was automatic, not the site. It is written out, as a generated converter cannot be called from
// this package's own code.
func InsulinDoseOther(source *SiteInsulinDose) types.InsulinDoseOther {
	return types.InsulinDoseOther{SplitNow: source.SplitNow, SplitExt: source.SplitExt}
}

// InsulinDoseOtherJSON is a dose's combo bolus split the site sends, as JSON for an update to merge into its other
// attributes
func InsulinDoseOtherJSON(source *SiteInsulinDose) *string {
	return types.OtherJSON(InsulinDoseOther(source))
}
