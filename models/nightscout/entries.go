package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
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
	Identifier *string    `json:"identifier,omitempty"`
	Type       *EntryType `json:"type,omitempty"`

	SGV       *FlexibleFloat64 `json:"sgv,omitempty"`
	MBG       *FlexibleFloat64 `json:"mbg,omitempty"`
	Direction *string          `json:"direction,omitempty"`
	Trend     *FlexibleFloat64 `json:"trend,omitempty"`
	Units     *Units           `json:"units,omitempty"`

	Device *string `json:"device,omitempty"`

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
	SrvCreated       *DateInt       `json:"srvCreated,omitempty"`
	SrvModified      *DateInt       `json:"srvModified,omitempty"`

	// Non-JSON values to make mapping easier
	Value       *FlexibleFloat64 `json:"-"`
	RawJsonData string           `json:"-"`
}

func (t *Entry) UnmarshalJSON(data []byte) error {
	type Alias Entry
	aux := struct {
		ID string `json:"_id"`
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
		t.Identifier = &aux.ID
	}

	// Setting the value
	if t.Value == nil {
		var res FlexibleFloat64
		if t.SGV != nil {
			res = *t.SGV
		} else if t.MBG != nil {
			res = *t.MBG
		}
		res = FlexibleFloat64(t.Units.ToMGDLMultiplier() * res.Float64())
		t.Value = &res
	}

	// Cleaning Updates
	if t.DateString != nil {
		if t.Date == nil || t.UtcOffsetMinutes == nil {
			t.Date, t.UtcOffsetMinutes = t.DateString.MillisUTCOffset()
		}
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
	if t.Date == nil {
		now := DateInt(time.Now().UnixMilli())
		t.Date = &now
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
