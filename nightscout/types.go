package nightscout

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
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
func DateStringFromTime(t time.Time) DateString {
	return DateString(t.Format(time.RFC3339Nano))
}

func DateStringFromInt64(millis int64) DateString {
	return DateStringFromTime(time.UnixMilli(millis))
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

/////
// Units
/////

type Units string

const (
	UnitsMGDL Units = "mg/dL"
	UnitsMMOL Units = "mmol/L"
)

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
		return fmt.Errorf("invalid units value %q", str)
	}
	return nil
}

////
// Utils
////

func ptr[T any](v T) *T {
	return &v
}
