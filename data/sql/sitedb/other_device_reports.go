package sitedb

import "nighttroop.t1d.nz/data/types"

// DeviceReportOtherJSON is a device report's other attributes as JSON for an update, as the update merges them into
// those kept
func DeviceReportOtherJSON(attributes types.DeviceReportOther) *string {
	return types.OtherJSON(attributes)
}
