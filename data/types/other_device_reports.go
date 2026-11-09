package types

import (
	"database/sql/driver"
)

// DeviceReportOther are a device status's other attributes, those without a column yet, kept as JSON
type DeviceReportOther struct{}

func (o DeviceReportOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *DeviceReportOther) Scan(src any) error {
	*o = DeviceReportOther{}
	return otherScan(src, o)
}
