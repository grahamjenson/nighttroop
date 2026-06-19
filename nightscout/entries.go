package nightscout

import (
	"encoding/json"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/data/sql/user"
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
	Identifier *string          `json:"identifier,omitempty"`
	Type       *EntryType       `json:"type,omitempty"`
	SGV        *FlexibleFloat64 `json:"sgv,omitempty"`
	MBG        *FlexibleFloat64 `json:"mbg,omitempty"`
	Date       *DateInt         `json:"date,omitempty"`
	UTCOffset  *FlexibleInt64   `json:"utcOffset,omitempty"`
	DateString *DateString      `json:"dateString,omitempty"`
	Device     *string          `json:"device,omitempty"`
	Direction  *string          `json:"direction,omitempty"`
	Units      *Units           `json:"units,omitempty"`
	RawJSON    string           `json:"-"`
}

////
// Mappers
////

func (e *Entry) FromMap(m map[string]any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, e)
	e.RawJSON = string(b)
	return err
}

func (e *Entry) ToMap() (map[string]any, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}

func (e *Entry) ToDBModel() (*user.GlucoseReading, error) {
	if e.Identifier == nil || e.SGV == nil || e.Date == nil {
		return nil, fmt.Errorf("missing required fields for DB model")
	}

	return &user.GlucoseReading{
		Eid:                   *e.Identifier,
		Value:                 float64(*e.SGV),
		Type:                  string(SGVType),
		RawJsonData:           &e.RawJSON,
		UtcOffsetMinutes:      e.UTCOffset.Int64(),
		OccurredAt:            e.Date.Time().UnixMilli(),
		CreatedByIdentityID:   0,  // This will need to be set by the caller based on the user context
		CreatedByIdentityName: "", // This will need to be set by the caller based on the user context
	}, nil
}

func (e *Entry) FromDBModel(sgv *user.GlucoseReading) error {
	e.Identifier = ptr(sgv.Eid)
	entryType := EntryType(sgv.Type)
	e.Type = &entryType
	e.SGV = ptr(FlexibleFloat64(sgv.Value))
	dateInt := DateInt(sgv.OccurredAt)
	e.Date = &dateInt
	dateString := DateStringFromInt64(sgv.OccurredAt)
	e.DateString = &dateString
	e.UTCOffset = ptr(FlexibleInt64(sgv.UtcOffsetMinutes))
	device := fmt.Sprintf("%d", sgv.CreatedByIdentityID)
	e.Device = &device

	return nil
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
