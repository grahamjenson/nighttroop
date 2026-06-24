package nightscout

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nighttroop.t1d.nz/utils/consts"
)

// //
// Numbers
// //
// FlexibleFloat64
type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	// number
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexibleFloat64(num)
		return nil
	}

	// string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", str, err)
		}
		*f = FlexibleFloat64(v)
		return nil
	}

	return fmt.Errorf("expected number or string, got %s", data)
}

func (f FlexibleFloat64) Float64() float64 {
	return float64(f)
}

type FlexibleInt64 int64

func (i *FlexibleInt64) UnmarshalJSON(data []byte) error {
	// number
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*i = FlexibleInt64(num)
		return nil
	}

	// string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", str, err)
		}
		*i = FlexibleInt64(v)
		return nil
	}

	return fmt.Errorf("expected number or string, got %s", data)
}

func (i *FlexibleInt64) Int64() int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

func (i *FlexibleInt64) Int() int {
	if i == nil {
		return 0
	}
	return int(*i)
}

// //
// Dates
// //
type DateString string

func (d *DateString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected date string, got %s", data)
	}
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return fmt.Errorf("invalid date string %q: %w", str, err)
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
func (d *DateString) MillisUTCOffset() (*DateInt, *FlexibleInt64) {
	t := d.Time()
	_, offsetSeconds := t.Zone()
	dateInt := DateInt(t.UnixMilli())
	offset := FlexibleInt64(int64(offsetSeconds) / 60)
	return &dateInt, &offset
}

func DateStringFromTime(t time.Time) DateString {
	return DateString(t.Format(time.RFC3339Nano))
}

func DateStringFromInt64(millis int64) DateString {
	return DateStringFromTime(time.UnixMilli(millis))
}

func DateStringFromDateIntAndUTCOffset(dateInt DateInt, offsetMinutes FlexibleInt64) DateString {
	t := time.UnixMilli(dateInt.Int64()).In(LocationFromUTCOffsetMinutes(offsetMinutes.Int()))
	return DateStringFromTime(t)
}

func LocationFromUTCOffsetMinutes(offsetMin int) *time.Location {
	return time.FixedZone(
		fmt.Sprintf("UTC%+d", offsetMin),
		offsetMin*60,
	)
}

type DateInt int64

func (d *DateInt) UnmarshalJSON(data []byte) error {
	var millis int64
	if err := json.Unmarshal(data, &millis); err != nil {
		return fmt.Errorf("expected date int, got %s", data)
	}
	t := time.UnixMilli(millis)
	*d = DateInt(t.UnixMilli())
	return nil
}

func DateIntFromTime(t time.Time) DateInt {
	return DateInt(t.UnixMilli())
}
func (d DateInt) Time() time.Time {
	return time.UnixMilli(int64(d))
}

func (d DateInt) Int64() int64 {
	return int64(d)
}

func (d DateInt) Int64p() *int64 {
	v := int64(d)
	return &v
}

type Timezone struct {
	Location *time.Location
	Name     string
}

func (t *Timezone) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("timezone must be a string: %w", err)
	}

	loc, err := time.LoadLocation(s)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", s, err)
	}

	t.Name = s
	t.Location = loc
	return nil
}

func (t Timezone) MarshalJSON() ([]byte, error) {
	if t.Name == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(t.Name)
}

/////
// Units
/////

type Units string

const (
	UnitsMGDL Units = "mg/dL"
	UnitsMMOL Units = "mmol/L"
)

func (u *Units) ToMGDLMultiplier() float64 {
	if u == nil {
		return 1.0
	}
	switch *u {
	case UnitsMGDL:
		return 1.0
	case UnitsMMOL:
		return consts.MMOL_L_TO_MG_DL
	default:
		return 0.0
	}
}

func (u *Units) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("expected units string, got %s", data)
	}
	switch strings.ToLower(str) {
	case "mg/dl", "mgdl", "mg":
		*u = UnitsMGDL
	case "mmol/l", "mmol", "mmoll":
		*u = UnitsMMOL
	default:
		// We Default to MGDL
		// TODO maybe default to user setting, but that adds a ton of complexity
		*u = UnitsMGDL
	}
	return nil
}

////
// Utils
////

func ptr[T any](v T) *T {
	return &v
}
