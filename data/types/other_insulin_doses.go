package types

import "database/sql/driver"

// InsulinDoseOther are an insulin dose's other attributes: a combo bolus's split, the fractions of it given at once
// and over time, and whether a loop gave it, Loop's automatic
type InsulinDoseOther struct {
	SplitNow  *float64 `json:"split_now,omitempty"`
	SplitExt  *float64 `json:"split_ext,omitempty"`
	Automatic *bool    `json:"automatic,omitempty"`
}

func (o InsulinDoseOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *InsulinDoseOther) Scan(src any) error {
	*o = InsulinDoseOther{}
	return otherScan(src, o)
}
