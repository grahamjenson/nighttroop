package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// TreatmentInsulinDoseOther is what only some apps send of a bolus, kept in its dose's other attributes: a combo
// bolus's split, sent as percents and kept as fractions, and whether a loop gave it. It is written out, not generated,
// as a generated converter cannot be called from this package's own code, which goverter compiles first.
func TreatmentInsulinDoseOther(source Treatment) datatypes.InsulinDoseOther {
	return datatypes.InsulinDoseOther{SplitNow: PercentageToMultiplier(source.SplitNow),
		SplitExt: PercentageToMultiplier(source.SplitExt), Automatic: source.Automatic}
}
