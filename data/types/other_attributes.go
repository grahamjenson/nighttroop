package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// A row's other attributes are the fields only some apps send, which are empty for most rows. They are kept together
// as a JSON object in its other_attributes column, NULL when it has none, rather than a column each.

// otherValue is a row's other attributes as their column's JSON object, or NULL when they are all unset
func otherValue(attributes any) (driver.Value, error) {
	b, err := json.Marshal(attributes)
	if err != nil || string(b) == "{}" {
		return nil, err
	}
	return string(b), nil
}

// OtherJSON is the other attributes set, as JSON for an update to merge into a row's, or nil when none are set
func OtherJSON(attributes any) *string {
	value, err := otherValue(attributes)
	if err != nil || value == nil {
		return nil
	}
	text := value.(string)
	return &text
}

// otherScan reads a row's other attributes from their column, leaving them all unset for NULL
func otherScan(src any, attributes any) error {
	switch src := src.(type) {
	case nil:
		return nil
	case string:
		return json.Unmarshal([]byte(src), attributes)
	case []byte:
		return json.Unmarshal(src, attributes)
	default:
		return fmt.Errorf("unsupported type for other attributes: %T", src)
	}
}

// GlucoseReadingOther are a glucose reading's other attributes: a CGM's raw signal, from xDrip+ and its followers, and
// AAPS's calibrations
type GlucoseReadingOther struct {
	Noise               *float64 `json:"noise,omitempty"`
	Filtered            *float64 `json:"filtered,omitempty"`
	Unfiltered          *float64 `json:"unfiltered,omitempty"`
	Rssi                *float64 `json:"rssi,omitempty"`
	IsCalibration       *bool    `json:"is_calibration,omitempty"`
	SensorMgdlAtPairing *float64 `json:"sensor_mgdl_at_pairing,omitempty"`
}

func (o GlucoseReadingOther) Value() (driver.Value, error) { return otherValue(o) }

func (o *GlucoseReadingOther) Scan(src any) error {
	*o = GlucoseReadingOther{}
	return otherScan(src, o)
}
