package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"
)

////
// Each model needs to be able to
// 1. Unmarshal from JSON
// 2. Unmarshal from a Map
// 3. From DB Model
// 4. To Map
// 5. To DB Model
////
// Models
////

type Entry struct {
	Identifier *IdentifierString `json:"identifier,omitempty"`
	Type       *EntryType        `json:"type,omitempty"`

	SGV       *FlexibleFloat64 `json:"sgv,omitempty"`
	MBG       *FlexibleFloat64 `json:"mbg,omitempty"`
	Direction *string          `json:"direction,omitempty"`
	Trend     *FlexibleFloat64 `json:"trend,omitempty"`
	Units     *Units           `json:"units,omitempty"`

	Filtered   *FlexibleFloat64 `json:"filtered,omitempty"`
	Unfiltered *FlexibleFloat64 `json:"unfiltered,omitempty"`
	RSSI       *FlexibleFloat64 `json:"rssi,omitempty"`
	Noise      *FlexibleFloat64 `json:"noise,omitempty"`
	SysTime    *DateString      `json:"sysTime,omitempty"`

	// Date fields
	Date             *DateInt       `json:"date,omitempty"`
	DateString       *DateString    `json:"dateString,omitempty"`
	Mills            *DateInt       `json:"mills,omitempty"`
	UtcOffsetMinutes *FlexibleInt64 `json:"utcOffset,omitempty"`

	// Default String Fields
	Device  *ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Reading     *FlexibleFloat64 `json:"-"`
	RawJsonData string           `json:"-"`
}

func (t *Entry) UnmarshalJSON(data []byte) error {
	type Alias Entry
	aux := struct {
		ID *IdentifierString `json:"_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Raw Value
	t.RawJsonData = string(data)

	// Identitfiers
	if t.Identifier == nil {
		t.Identifier = aux.ID
	}

	// Setting the value
	if t.Reading == nil {
		var res FlexibleFloat64
		if t.SGV != nil {
			res = *t.SGV
		} else if t.MBG != nil {
			res = *t.MBG
		}
		res = FlexibleFloat64(t.Units.ToMGDLMultiplier() * res.Float64())
		t.Reading = &res
	}

	// We Just need Date and UtcOffset to be populated
	if t.DateString != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			t.Date, t.UtcOffsetMinutes = t.DateString.MillisUTCOffset()
		}
	}

	if t.Mills != nil && t.Date == nil {
		t.Date = t.Mills
	}

	return nil
}

func (t *Entry) MarshalJSON() ([]byte, error) {
	type Alias Entry
	// Clean Up and Populate Date Fields
	if t.UtcOffsetMinutes == nil {
		zero := FlexibleInt64(0)
		t.UtcOffsetMinutes = &zero
	}

	ds := DateStringFromDateIntAndUTCOffset(*t.Date, *t.UtcOffsetMinutes)
	t.DateString = &ds

	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	})
}

//////
// Subtypes and Constants
//////

type EntryType string

const (
	SGVType EntryType = "sgv"
	MBGType EntryType = "mbg"
)

func (t *EntryType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sgv":
		*t = SGVType
	case "mbg":
		*t = MBGType
	default:
		return fmt.Errorf("unknown entry type: %s", s)
	}

	return nil
}

func (t Entry) GetDevice() *string {
	return t.Device.Stringp()
}

func (t Entry) GetApp() *string {
	return t.App.Stringp()
}

func (t Entry) GetSubject() *string {
	return t.Subject.Stringp()
}

func (t Entry) GetDate() *int64 {
	return t.Date.Int64p()
}

func (t Entry) GetUtcOffsetMinutes() *int64 {
	return t.UtcOffsetMinutes.Int64p()
}

func (t Entry) GetIdentifier() *string {
	return t.Identifier.Stringp()
}
