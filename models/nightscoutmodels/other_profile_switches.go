package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// TreatmentProfileSwitchOther is what only AAPS sends of a profile switch, kept in its other attributes: the AAPS
// switch it was made from, which its followers read an effective profile switch (a Note) by. It is written out, not
// generated, as a generated converter cannot be called from this package's own code, which goverter compiles first.
func TreatmentProfileSwitchOther(source Treatment) datatypes.ProfileSwitchOther {
	return datatypes.ProfileSwitchOther{OriginalProfileName: source.OriginalProfileName.Stringp(),
		OriginalCustomizedName: source.OriginalCustomizedName.Stringp(), OriginalTimeshiftMs: source.OriginalTimeshiftMs.Int64p(),
		OriginalPercentage: source.OriginalPercentage.Int64p(), OriginalDurationMs: source.OriginalDurationMs.Int64p(),
		OriginalEnd: source.OriginalEnd.Int64p()}
}
