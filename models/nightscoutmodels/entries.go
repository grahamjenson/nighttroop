package nightscoutmodels

import (
	"encoding/json"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/types"
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
	Identifier *types.IdentifierString `json:"identifier,omitempty"`
	Type       *EntryType              `json:"type,omitempty"`

	SGV       *types.FlexibleFloat64 `json:"sgv,omitempty"`
	MBG       *types.FlexibleFloat64 `json:"mbg,omitempty"`
	Direction *string                `json:"direction,omitempty"`
	Trend     *types.FlexibleFloat64 `json:"trend,omitempty"`
	Units     *consts.Units          `json:"units,omitempty"`

	Filtered   *types.FlexibleFloat64 `json:"filtered,omitempty"`
	Unfiltered *types.FlexibleFloat64 `json:"unfiltered,omitempty"`
	RSSI       *types.FlexibleFloat64 `json:"rssi,omitempty"`
	Noise      *types.FlexibleFloat64 `json:"noise,omitempty"`
	SysTime    *types.DateString      `json:"sysTime,omitempty"`

	// Date fields
	Date             *types.DateInt       `json:"date,omitempty"`
	DateString       *types.DateString    `json:"dateString,omitempty"`
	Mills            *types.DateInt       `json:"mills,omitempty"`
	UtcOffsetMinutes *types.FlexibleInt64 `json:"utcOffset,omitempty"`

	// Default String Fields
	Device  *types.ShortString `json:"device,omitempty"`  // Device used for the activity
	App     *types.ShortString `json:"app,omitempty"`     // App used for the activity
	Subject *types.ShortString `json:"subject,omitempty"` // Subject of the activity

	// Non-JSON values to make mapping easier
	Mgdl        *types.FlexibleFloat64 `json:"-"`
	RawJsonData string                 `json:"-"`
}

func (t *Entry) UnmarshalJSON(data []byte) error {
	type Alias Entry
	aux := struct {
		ID *types.IdentifierString `json:"_id"`
		// Trio's own id, which it deletes by
		TrioID json.RawMessage `json:"id"`
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
	if t.Identifier == nil {
		t.Identifier = types.IdentifierFromJSON(aux.TrioID)
	}

	// Setting the value, only when sgv or mbg is given so a patch without them leaves it as it is
	if t.Mgdl == nil {
		value := t.SGV
		if value == nil {
			value = t.MBG
		}

		if value != nil {
			res := types.FlexibleFloat64(t.Units.ToMGDL(value.Float64()))
			t.Mgdl = &res
		}
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
		zero := types.FlexibleInt64(0)
		t.UtcOffsetMinutes = &zero
	}

	ds := types.DateStringFromDateIntAndUTCOffset(*t.Date, *t.UtcOffsetMinutes)
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

func (t *EntryType) String() string {
	if t == nil {
		return ""
	}
	return string(*t)
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
