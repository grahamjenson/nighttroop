package nightscoutmodels

import datatypes "nighttroop.t1d.nz/data/types"

// TreatmentComponentEventOther is what only some apps send of a treatment's component, kept in its component event's
// other attributes: the sensor's code and transmitter, and the pump's ID, serial and type. It is written out, not
// generated, as a generated converter cannot be called from this package's own code, which goverter compiles first.
func TreatmentComponentEventOther(source Treatment) datatypes.ComponentEventOther {
	return datatypes.ComponentEventOther{SensorCode: source.SensorCode.Stringp(),
		TransmitterID: source.TransmitterID.Stringp(), PumpID: source.PumpID.Stringp(),
		PumpSerial: source.PumpSerial.Stringp(), PumpType: source.PumpType.Stringp()}
}
