package types

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("bad date string: %s", value)
}
