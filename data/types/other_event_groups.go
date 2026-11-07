package types

import "database/sql/driver"

// EventGroupOther are a treatment's other attributes: the app's own id for it, given back as sent (xDrip+'s uuid), and
// AAPS's running mode (an OpenAPS Offline), whether the loop forced it and why, and how long it was to last in ms
type EventGroupOther struct {
	Uuid               *string `json:"uuid,omitempty"`
	Mode               *string `json:"mode,omitempty"`
	AutoForced         *bool   `json:"auto_forced,omitempty"`
	Reasons            *string `json:"reasons,omitempty"`
	OriginalDurationMs *int64  `json:"original_duration_ms,omitempty"`
}

func (o EventGroupOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *EventGroupOther) Scan(src any) error {
	*o = EventGroupOther{}
	return otherScan(src, o)
}
