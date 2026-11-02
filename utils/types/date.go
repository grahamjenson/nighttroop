package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// //////
// DateInt
// ///////

type DateInt int64

func (d *DateInt) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected date int, got %s", data)
	}

	var mills int64
	switch val := v.(type) {
	case float64:
		mills = int64(val)

	case json.Number:
		if asInt, err := val.Int64(); err == nil {
			mills = asInt
		} else {
			asFloat, err := val.Float64()
			if err != nil {
				return fmt.Errorf("invalid date int value %q: %w", val, err)
			}
			mills = int64(asFloat)
		}
	case string:
		var err error
		mills, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			// Try Parsing as a DateString and then converting to DateInt

			t, err := StringToTime(val)
			if err != nil {
				return fmt.Errorf("invalid date int %q: %w", val, err)
			}
			mills = t.UnixMilli()
		}
	default:
		return fmt.Errorf("expected number or string for date int, got %T", v)
	}

	// We are going to assume anything over a trillion is a second so convert to mills
	if mills < 1_000_000_000_000 {
		mills = mills * 1000
	}

	*d = DateInt(mills)
	return nil
}

func DateIntFromTime(t time.Time) DateInt {
	return DateInt(t.UnixMilli())
}
func (d DateInt) Time() time.Time {
	return time.UnixMilli(int64(d))
}

func (d DateInt) TimeString() string {
	return time.UnixMilli(int64(d)).UTC().Format(time.RFC3339Nano)
}

func (d *DateInt) Int64() int64 {
	if d == nil {
		return 0
	}
	return int64(*d)
}

func (d *DateInt) Int64p() *int64 {
	if d == nil {
		return nil
	}
	v := int64(*d)
	return &v
}

// StringToTime tries the date layouts clients send until one parses
func StringToTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		time.Layout,
		time.DateTime,
		time.DateOnly,
		"2006-01-02T15:04:05,000Z07:00",
		"2006-01-02T15:04:05,000-0700",
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.000Z",
	}
	// Juggluco writes a zone west of UTC with half hours, e.g. Newfoundland's, with a second minus: -03:-30
	value = strings.Replace(value, ":-", ":", 1)
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("bad date string: %s", value)
}

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
		// A fractional or exponent number is not an int64, so fall back to reading it as a float
		if asInt, err := typed.Int64(); err == nil {
			t = time.UnixMilli(asInt)
		} else if asFloat, err := typed.Float64(); err == nil {
			t = time.UnixMilli(int64(asFloat))
		} else {
			return fmt.Errorf("invalid date number %q: %w", typed, err)
		}

	case string:
		t, err = StringToTime(typed)
		if err != nil {
			return fmt.Errorf("invalid date string %q: %w", typed, err)
		}
	default:
		return fmt.Errorf("expected string for date, got %T", v)
	}

	// A number has no time zone, so it is UTC rather than whatever zone the server runs in
	if _, isString := v.(string); !isString {
		t = t.UTC()
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

func (d *DateString) MillisUTCOffset() (*DateInt, *FlexibleInt64) {
	t := d.Time()
	_, offsetSeconds := t.Zone()
	dateInt := DateInt(t.UnixMilli())
	offset := FlexibleInt64(int64(offsetSeconds) / 60)
	return &dateInt, &offset
}

func (d *DateString) String() string {
	return string(*d)
}

func DateStringNow() (*DateString, time.Time) {
	now := time.Now().UTC()
	dateString := DateStringFromTime(now)
	return &dateString, now
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
