package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"nighttroop.t1d.nz/utils/consts"
)

// Units are glucose units as they were sent, e.g. mg/dl or mmol, and what they mean. They are kept as sent, so the
// Nightscout API gives them back as a client wrote them, as xDrip4iOS matches a BG check by them; other APIs use
// what they mean
type Units struct {
	consts.Units
	Sent string
}

// UnitsFromEnum is units sent as Nighttroop writes them
func UnitsFromEnum(units consts.Units) Units {
	return Units{Units: units, Sent: string(units)}
}

// UnitsToEnum is what units mean
func UnitsToEnum(units Units) consts.Units {
	return units.Units
}

// CopyUnits is units as they are, as a converter copies them
func CopyUnits(units Units) Units {
	return units
}

// Value stores the units as they were sent
func (u Units) Value() (driver.Value, error) {
	if u.Sent == "" {
		return string(u.Units), nil
	}
	return u.Sent, nil
}

// MarshalJSON writes the units as they were sent, as they are stored
func (u Units) MarshalJSON() ([]byte, error) {
	sent, err := u.Value()
	if err != nil {
		return nil, err
	}
	return json.Marshal(sent)
}

// UnmarshalJSON reads units written as they were sent
func (u *Units) UnmarshalJSON(data []byte) error {
	var sent string
	if err := json.Unmarshal(data, &sent); err != nil {
		return err
	}
	return u.Scan(sent)
}

// Scan reads units stored as they were sent, and what they mean
func (u *Units) Scan(src any) error {
	var sent string
	switch typed := src.(type) {
	case string:
		sent = typed
	case []byte:
		sent = string(typed)
	default:
		return fmt.Errorf("cannot scan %T as units", src)
	}
	units, err := consts.ParseUnits(sent)
	if err != nil {
		return err
	}
	*u = Units{Units: units, Sent: sent}
	return nil
}
