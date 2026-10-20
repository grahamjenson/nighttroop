package nightscoutmodels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"nighttroop.t1d.nz/utils/types"
)

// //////
// DateString
// ///////
type DateString string

func (d *DateString) UnmarshalJSON(data []byte) error {
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("expected date string, got %s", data)
	}

	var t time.Time
	var err error
	switch typed := v.(type) {
	case nil:
		return nil
	case float64:
		t = time.UnixMilli(int64(typed))
	case json.Number:
		if asInt, err := typed.Int64(); err == nil {
			t = time.UnixMilli(asInt)
		}
		asFloat, err := typed.Float64()
		if err != nil {
			t = time.UnixMilli(int64(asFloat))
		}

	case string:
		t, err = types.StringToTime(typed)
		if err != nil {
			return fmt.Errorf("invalid date string %q: %w", typed, err)
		}
	default:
		return fmt.Errorf("expected string for date, got %T", v)
	}

	*d = DateString(t.Format(time.RFC3339Nano))
	return nil
}

func (d *DateString) Time() time.Time {
	if d == nil {
		return time.Time{}
	}
	// This should have errored on unmarshal or assignment
	t, _ := time.Parse(time.RFC3339Nano, string(*d))
	return t
}

func (d *DateString) Timep() *time.Time {
	if d == nil {
		return nil
	}
	// This should have errored on unmarshal or assignment
	t, _ := time.Parse(time.RFC3339Nano, string(*d))
	return &t
}

func (d *DateString) UnixMillip() *int64 {
	t := d.Time()
	dateInt := t.UnixMilli()
	return &dateInt
}

func (d *DateString) UnixMilli() int64 {
	t := d.Time()
	dateInt := t.UnixMilli()
	return dateInt
}

func (d *DateString) MillisUTCOffset() (*types.DateInt, *types.FlexibleInt64) {
	t := d.Time()
	_, offsetSeconds := t.Zone()
	dateInt := types.DateInt(t.UnixMilli())
	offset := types.FlexibleInt64(int64(offsetSeconds) / 60)
	return &dateInt, &offset
}

func (d *DateString) String() string {
	return string(*d)
}

func DateStringNow() (*DateString, time.Time) {
	now := time.Now().UTC()
	return ptr(DateStringFromTime(now)), now
}

func DateStringFromTime(t time.Time) DateString {
	return DateString(t.Format(time.RFC3339Nano))
}

func DateStringFromInt64AndUTCOffset(dateInt int64, offsetMinutes *int64) DateString {
	if offsetMinutes == nil {
		t := time.UnixMilli(dateInt).UTC()
		return DateStringFromTime(t)
	}
	t := time.UnixMilli(dateInt).In(LocationFromUTCOffsetMinutes(int(*offsetMinutes)))
	return DateStringFromTime(t)
}

func DateStringFromDateIntAndUTCOffset(dateInt types.DateInt, offsetMinutes types.FlexibleInt64) DateString {
	t := time.UnixMilli(dateInt.Int64()).In(LocationFromUTCOffsetMinutes(offsetMinutes.Int()))
	return DateStringFromTime(t)
}

func LocationFromUTCOffsetMinutes(offsetMin int) *time.Location {
	return time.FixedZone(
		fmt.Sprintf("UTC%+d", offsetMin),
		offsetMin*60,
	)
}

// //////
// Timezone
// ///////

type Timezone string

func (t *Timezone) Stringp() *string {
	if t == nil {
		return nil
	}
	str := string(*t)
	return &str
}

func (t *Timezone) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("timezone must be a string: %w", err)
	}
	*t = Timezone(s)
	return nil
}

func (t Timezone) MarshalJSON() ([]byte, error) {
	if t == "" {
		return []byte(`"UTC"`), nil
	}
	return json.Marshal(string(t))
}

// ///
// Notes
// ///

type JSONString string

func (s *JSONString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected JSON string, got %s", data)
	}

	if len(str) > 2048 {
		str = str[:2048]
	}

	*s = JSONString(str)
	return nil
}

func (s *JSONString) String() string {
	return string(*s)
}

type LongJSONString string

func (s *LongJSONString) String() string {
	return string(*s)
}

func (s *LongJSONString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

func (n *LongJSONString) UnmarshalJSON(data []byte) error {
	// We are going to sanitize notes here, no html allowed, trim white space, convert to string
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}

	// Test that this is valid json string
	var testJSON any
	err := json.Unmarshal([]byte(v), &testJSON)
	if err != nil {
		return fmt.Errorf("invalid JSON string, got %s", data)
	}

	// Concat to 1024 characters
	if len(v) > 4024 {
		v = v[:4024]
	}

	*n = LongJSONString(v)
	return nil
}

type IdentifierString string

func (i *IdentifierString) String() string {
	if i == nil {
		return ""
	}
	return string(*i)
}

func (i *IdentifierString) Stringp() *string {
	if i == nil {
		return nil
	}
	str := string(*i)
	return &str
}

func (i *IdentifierString) UnmarshalJSON(data []byte) error {
	// limit in possible characters and length, convert everything to a string
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected identifier string, got %s", data)
	}

	str := ""

	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		str = v.String()
	default:
		str = fmt.Sprintf("%v", v)
	}

	if len(str) > 32 {
		str = str[:32]
	}

	if len(str) == 0 {
		return fmt.Errorf("identifier string cannot be empty")
	}

	*i = IdentifierString(str)
	return nil
}

////
// Utils
////

func ptr[T any](v T) *T {
	return &v
}
