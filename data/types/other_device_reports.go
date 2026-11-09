package types

import (
	"database/sql/driver"
	"encoding/json"
)

// DeviceReportOther are a device status's other attributes: the pump's clock, manufacturer and model, Loop's override,
// and the loop's name and its status, with its predictions. The JSON ones are kept as JSON, as sent.
type DeviceReportOther struct {
	PumpClock                *int64          `json:"pump_clock,omitempty"`
	PumpManufacturer         *string         `json:"pump_manufacturer,omitempty"`
	PumpModel                *string         `json:"pump_model,omitempty"`
	Override                 json.RawMessage `json:"override,omitempty"`
	LoopName                 *string         `json:"loop_name,omitempty"`
	LoopSettings             json.RawMessage `json:"loop_settings,omitempty"`
}

func (o DeviceReportOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *DeviceReportOther) Scan(src any) error {
	*o = DeviceReportOther{}
	return otherScan(src, o)
}
