package types

import "database/sql/driver"

// ComponentEventOther are a component event's other attributes: the sensor's code and transmitter, and the pump's ID,
// serial and type, which only some apps send
type ComponentEventOther struct {
	SensorCode    *string `json:"sensor_code,omitempty"`
	TransmitterID *string `json:"transmitter_id,omitempty"`
	PumpID        *string `json:"pump_id,omitempty"`
	PumpSerial    *string `json:"pump_serial,omitempty"`
	PumpType      *string `json:"pump_type,omitempty"`
}

func (o ComponentEventOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *ComponentEventOther) Scan(src any) error {
	*o = ComponentEventOther{}
	return otherScan(src, o)
}
