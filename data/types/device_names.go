package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// DeviceNames are the names devices give their statuses, kept as a JSON array
type DeviceNames []string

func (d DeviceNames) Value() (driver.Value, error) {
	if d == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]string(d))
	return string(b), err
}

func (d *DeviceNames) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*d = nil
		return nil
	case string:
		return json.Unmarshal([]byte(src), d)
	case []byte:
		return json.Unmarshal(src, d)
	default:
		return fmt.Errorf("unsupported type for DeviceNames: %T", src)
	}
}
