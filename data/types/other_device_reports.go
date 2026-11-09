package types

import (
	"database/sql/driver"
	"encoding/json"
)

// DeviceReportOther are a device status's other attributes: the pump's clock, manufacturer and model, Loop's override,
// the loop's name, whose status is the report's loop_status_json, and the pump and uploader as sent, for their fields
// without a column, like AAPS's extended. The JSON ones are kept as JSON.
type DeviceReportOther struct {
	PumpClock        *int64          `json:"pump_clock,omitempty"`
	PumpManufacturer *string         `json:"pump_manufacturer,omitempty"`
	PumpModel        *string         `json:"pump_model,omitempty"`
	Override         json.RawMessage `json:"override,omitempty"`
	LoopName         *string         `json:"loop_name,omitempty"`
	Pump             json.RawMessage `json:"pump,omitempty"`
	Uploader         json.RawMessage `json:"uploader,omitempty"`
}

func (o DeviceReportOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *DeviceReportOther) Scan(src any) error {
	*o = DeviceReportOther{}
	return otherScan(src, o)
}
