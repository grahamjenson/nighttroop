package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// TreatmentEventGroupOther is what only some apps send of a treatment, kept in its group's other attributes: the app's
// own id for it (xDrip+'s uuid), and AAPS's running mode, whether the loop forced it and why, and how long it was to
// last. It is written out, not generated, as a generated converter cannot be called from this package's own code,
// which goverter compiles first.
func TreatmentEventGroupOther(source Treatment) datatypes.EventGroupOther {
	return datatypes.EventGroupOther{Uuid: source.UUID.Stringp(), Mode: source.Mode.Stringp(),
		AutoForced: source.AutoForced, Reasons: source.Reasons.Stringp(), OriginalDurationMs: source.OriginalDurationMs.Int64p()}
}
