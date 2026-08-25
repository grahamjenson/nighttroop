package nightscout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"nighttroop.t1d.nz/utils/consts"
)

// //////
// FlexibleFloat64
// ///////

type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("expected float value, got %s", data)
	}

	switch val := v.(type) {
	case float64:
		*f = FlexibleFloat64(val)
	case json.Number:
		if asFloat, err := val.Float64(); err == nil {
			*f = FlexibleFloat64(asFloat)
			return nil
		}
		return fmt.Errorf("invalid float value %q", val)
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q", val)
		}
		*f = FlexibleFloat64(newF)
	default:
		return fmt.Errorf("expected number or string for float, got %v:%T", v, v)
	}
	return nil
}

func (f FlexibleFloat64) Float64() float64 {
	return float64(f)
}

func (f *FlexibleFloat64) Float64p() *float64 {
	if f == nil {
		return nil
	}
	v := float64(*f)
	return &v
}

// //////
// FlexibleInt64
// ///////

type FlexibleInt64 int64

func (i *FlexibleInt64) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected int value, got %s", data)
	}

	switch val := v.(type) {
	case float64:
		*i = FlexibleInt64(int64(val))
	case json.Number:
		if asInt, err := val.Int64(); err == nil {
			*i = FlexibleInt64(asInt)
			return nil
		}
		asFloat, err := val.Float64()
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		*i = FlexibleInt64(int64(asFloat))
	case string:
		var err error
		newF, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		*i = FlexibleInt64(int64(newF))
	default:
		return fmt.Errorf("expected number or string for int, got %T", v)
	}

	return nil
}

func (i *FlexibleInt64) Int64p() *int64 {
	if i == nil {
		return nil
	}
	v := int64(*i)
	return &v
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

// //////
// FlexibleBool
// ///////

type FlexibleBool bool

func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected bool value, got %s", data)
	}

	switch val := v.(type) {
	case bool:
		*b = FlexibleBool(val)
	case string:
		lower := strings.ToLower(strings.TrimSpace(val))
		if lower == "true" || lower == "1" {
			*b = FlexibleBool(true)
		} else if lower == "false" || lower == "0" {
			*b = FlexibleBool(false)
		} else {
			return fmt.Errorf("invalid bool value %q", val)
		}
	default:
		return fmt.Errorf("expected bool or string for bool, got %T", v)
	}
	return nil
}

func (b *FlexibleBool) Boolp() *bool {
	if b == nil {
		return nil
	}
	v := bool(*b)
	return &v
}

func (b *FlexibleBool) Bool() bool {
	if b == nil {
		return false
	}
	return bool(*b)
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
		if asInt, err := typed.Int64(); err == nil {
			t = time.UnixMilli(asInt)
		}
		asFloat, err := typed.Float64()
		if err != nil {
			t = time.UnixMilli(int64(asFloat))
		}

	case string:
		t, err = stringToTime(typed)
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

			t, err := stringToTime(val)
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

/////
// Units
/////

type Units string

const (
	UnitsMGDL Units = "mg/dL"
	UnitsMMOL Units = "mmol/L"
)

func (u *Units) Stringp() *string {
	if u == nil {
		return nil
	}
	str := string(*u)
	return &str
}

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

type ShortString string

func (s *ShortString) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}
	var str string
	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	if len(str) > 64 {
		str = str[:64]
	}

	*s = ShortString(str)
	return nil
}

func (s *ShortString) String() string {
	return string(*s)
}

func (s *ShortString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
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

type SanitizedString string

func (s *SanitizedString) String() string {
	return string(*s)
}

func (s *SanitizedString) Stringp() *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func StripAllTags(input string) string {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return input
	}

	var buf bytes.Buffer
	var f func(*html.Node)

	f = func(n *html.Node) {
		// Only keep text nodes, skip script and style contents
		if n.Type == html.TextNode {
			if n.Parent != nil && (n.Parent.Data == "script" || n.Parent.Data == "style") {
				return
			}
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return buf.String()
}

func (n *SanitizedString) UnmarshalJSON(data []byte) error {
	// We are going to sanitize notes here, no html allowed, trim white space, convert to string
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("expected notes string, got %s", data)
	}
	var str string
	switch v := v.(type) {
	case string:
		str = v
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	// Trim white space
	str = strings.TrimSpace(str)
	// Concat to 1024 characters
	if len(str) > 1024 {
		str = str[:1024]
	}
	// Strip all tags
	str = StripAllTags(str)

	*n = SanitizedString(str)
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

// We are trying to find a time in the time
func stringToTime(value string) (time.Time, error) {
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
