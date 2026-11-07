package sitemodels

import "nighttroop.t1d.nz/data/types"

// ComponentEventOther is a component event's sensor and pump fields the site sends, kept in its other attributes. It is
// written out, as a generated converter cannot be called from this package's own code.
func ComponentEventOther(source *SiteComponentEvent) types.ComponentEventOther {
	return types.ComponentEventOther{SensorCode: source.SensorCode, TransmitterID: source.TransmitterID,
		PumpID: source.PumpID, PumpSerial: source.PumpSerial, PumpType: source.PumpType}
}

// ComponentEventOtherJSON is a component event's sensor and pump fields the site sends, as JSON for an update to merge
// into its other attributes
func ComponentEventOtherJSON(source *SiteComponentEvent) *string {
	return types.OtherJSON(ComponentEventOther(source))
}
