package types

import "database/sql/driver"

// ActivityOther are an activity's other attributes: what a fitness tracker or watch measured of it, such as xDrip+'s
// heart rate and steps
type ActivityOther struct {
	Calories  *int64 `json:"calories,omitempty"`
	Steps     *int64 `json:"steps,omitempty"`
	HeartRate *int64 `json:"heart_rate,omitempty"`
}

func (o ActivityOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *ActivityOther) Scan(src any) error {
	*o = ActivityOther{}
	return otherScan(src, o)
}
