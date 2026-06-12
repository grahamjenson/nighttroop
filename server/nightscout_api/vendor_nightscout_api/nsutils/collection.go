package nsutils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/nserrors"
)

type Collection string

const (
	Entries      Collection = "entries"
	Treatments   Collection = "treatments"
	Devicestatus Collection = "devicestatus"
	Profile      Collection = "profile"
	Properties   Collection = "properties"
	Food         Collection = "food"
)

const DATE_FORMAT_STRING = "2006-01-02T15:04:05.000Z"

func ParseCollection(s string) (Collection, error) {

	// Validate and normalize
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "device_status":
		s = "devicestatus"
	case "profiles":
		s = "profile"
	case "foods":
		s = "food"
	}

	c := Collection(s)
	switch c {
	case Entries,
		Treatments,
		Devicestatus,
		Profile,
		Properties,
		Food:
		return c, nil
	default:
		return "", fmt.Errorf("unknown collection %q", s)
	}
}

func (c Collection) DefaultDateField() string {
	switch c {
	case Entries:
		return "date"
	default:
		return "created_at"
	}
}

// NormalizeData normalizes incoming Nightscout payloads into a consistent shape
// before persistence.
func (c Collection) DefaultNormalizeValidateCreateDoc(dontUseData map[string]any) (map[string]any, error) {
	if dontUseData == nil {
		return nil, nserrors.BadDataBodyError
	}
	clean := CloneMap(dontUseData)

	// default [date]
	now := time.Now().UTC()
	if v, ok := clean["date"]; !ok || (v == nil || v == 0) {
		clean["date"] = now.UnixMilli()
	}
	// default [created_at]
	if value, ok := StringField(clean, "created_at"); !ok || value == "" {
		// TODO default offset to users set local timezone
		clean["created_at"] = now.Format(DATE_FORMAT_STRING)
	}

	// default [utcOffset] if created_at or dateString exist
	if v, ok := clean["utcOffset"]; !ok || v == nil {
		if _, off, _ := parseDateValue(clean["created_at"]); off != nil {
			clean["utcOffset"] = *off
		} else if _, off, _ := parseDateValue(clean["dateString"]); off != nil {
			clean["utcOffset"] = *off
		} else {
			// TODO DEFAULT AS USER SET TIMEZONE
			clean["utcOffset"] = 0
		}
	}

	// default [app]
	if app, ok := StringField(clean, "app"); !ok || strings.TrimSpace(app) == "" {
		clean["app"] = "unknown"
	}

	// default [type] for entries
	if c == Entries {
		if _, ok := StringField(clean, "type"); !ok {
			clean["type"] = "sgv"
		}
	}

	// Ensure identifier exists
	if _, ok := StringField(clean, "identifier"); !ok {
		if existingID, ok := StringField(clean, "_id"); ok && existingID != "" {
			clean["identifier"] = existingID
		}
	}

	// Normalize fields
	if err := normalizeNumericFields(clean,
		"sgv",
		"mbg",
		"carbs",
		"insulin",
		"preBolus",
		"glucose",
		"uploaderBattery",
	); err != nil {
		return nil, err
	}

	// normalize [created_at]
	if value, ok := StringField(clean, "created_at"); ok && value != "" {
		normalized, _, err := toUTCString(value)
		if err != nil {
			return nil, nserrors.BadCreatedAtError
		}
		clean["created_at"] = normalized
	}

	// normalize [date]
	if rawDate, exists := clean["date"]; exists && rawDate != nil {
		millis, _, err := parseDateValue(rawDate)
		if err != nil {
			return nil, nserrors.BadDateFieldError
		}
		clean["date"] = millis
	}

	// normalize [dateString]
	if dateString, ok := StringField(clean, "dateString"); ok && dateString != "" {
		normalized, _, err := toUTCString(dateString)
		if err != nil {
			return nil, nserrors.BadDateStringError
		}
		clean["dateString"] = normalized
	}

	// normalize [notes]
	if notes, ok := StringField(clean, "notes"); ok {
		clean["notes"] = sanitizeNotes(notes)
	}

	// Validate Fields

	// validate [utcOffset]
	offset, ok, err := Int64Field(clean, "utcOffset")
	if err != nil || !ok {
		return nil, nserrors.BadUTCTOffsetError
	}

	if ok && (offset < -1440 || offset > 1440) {
		return nil, nserrors.BadUTCTOffsetError
	}

	// validate [date]
	date, ok, err := Int64Field(clean, "date")
	if err != nil || !ok {
		return nil, nserrors.BadDateFieldError
	}
	if ok && date <= 100000000000 {
		return nil, nserrors.BadDateFieldError
	}

	// validate [created_at]
	if value, ok := StringField(clean, "created_at"); ok {
		// Make sure the time is the correct time format
		if _, err := time.Parse(DATE_FORMAT_STRING, value); err != nil {
			return nil, nserrors.BadCreatedAtError
		}
	}

	// validate [dateString]
	if value, ok := StringField(clean, "dateString"); ok {
		// Make sure the time is the correct time format
		if _, err := time.Parse(DATE_FORMAT_STRING, value); err != nil {
			return nil, nserrors.BadDateStringError
		}
	}

	if app, ok := StringField(clean, "app"); !ok || app == "" {
		return nil, fmt.Errorf("Bad or missing app field")
	}

	return clean, nil
}

func (c Collection) DefaultNormalizeValidateUpdateDoc(existing Record, dontUseData map[string]any) (map[string]any, error) {
	if dontUseData == nil {
		return nil, nserrors.BadDataBodyError
	}
	clean := CloneMap(dontUseData)

	// Ignore the identifier if it is different from the existing record
	if identifier, ok := StringField(clean, "identifier"); ok && identifier != "" && identifier != existing.Identifier() {
		delete(clean, "identifier")
	}

	immutable := []string{"identifier", "date", "utcOffset", "eventType", "device", "app", "srvCreated", "subject", "srvModified", "modifiedBy", "isValid"}
	for _, field := range immutable {
		if proposed, exists := clean[field]; exists {
			current := PathValue(existing.ToMap(true), field)
			if field == "identifier" {
				continue
			}
			if fmt.Sprint(proposed) != fmt.Sprint(current) {
				return nil, fmt.Errorf("Field %s cannot be modified by the client", field)
			}
		}
	}

	return clean, nil
}

func normalizeNumericFields(data map[string]any, fields ...string) error {
	for _, field := range fields {
		value, ok := data[field]
		if !ok || value == nil {
			continue
		}
		number, err := toFloat64(value)
		if err != nil {
			return fmt.Errorf("Bad numeric field %s", field)
		}
		if number == float64(int64(number)) {
			data[field] = int64(number)
			continue
		}
		data[field] = number
	}
	return nil
}

// /
// Field Processing
// //

// StringField reads a field as a string when possible.
func StringField(data map[string]any, key string) (string, bool) {
	val, ok := data[key]
	if !ok || val == nil {
		return "", false
	}
	switch typed := val.(type) {
	case string:
		return typed, true
	case json.Number:
		return typed.String(), true
	case fmt.Stringer:
		return typed.String(), true
	default:
		return fmt.Sprint(typed), true
	}
}

// Int64Field reads a field as an int64 when possible.
func Int64Field(data map[string]any, key string) (int64, bool, error) {
	val, ok := data[key]
	if !ok || val == nil {
		return 0, false, nil
	}
	switch typed := val.(type) {
	case int:
		return int64(typed), true, nil
	case int32:
		return int64(typed), true, nil
	case int64:
		return typed, true, nil
	case float32:
		return int64(typed), true, nil
	case float64:
		return int64(typed), true, nil
	case json.Number:
		num, err := typed.Int64()
		return num, err == nil, nil
	case string:
		if typed == "" {
			return 0, false, nserrors.BadOffsetError
		}
		num, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0, false, nserrors.BadOffsetError
		}
		return num, true, nil
	default:
		return 0, false, nserrors.BadOffsetError
	}
}

// toUTCString normalizes a timestamp string into Nightscout's UTC format and
// returns the original offset in minutes.
func toUTCString(value string) (string, int, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		layouts := []string{
			"2006-01-02T15:04:05.000-0700",
			"2006-01-02T15:04:05-0700",
			"2006-01-02T15:04:05.000Z07:00",
			"2006-01-02T15:04:05Z07:00",
		}
		for _, layout := range layouts {
			parsed, err = time.Parse(layout, value)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return "", 0, err
	}
	_, offsetSeconds := parsed.Zone()
	return parsed.UTC().Format(DATE_FORMAT_STRING), offsetSeconds / 60, nil
}

func sanitizeNotes(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if strings.Contains(trimmed, "<img") {
		return "<img>"
	}
	return value
}

func parseDateValue(value any) (int64, *int, error) {
	switch typed := value.(type) {
	case nil:
		return 0, nil, nil
	case int64:
		return typed, nil, nil
	case int:
		return int64(typed), nil, nil
	case int32:
		return int64(typed), nil, nil
	case float64:
		return int64(typed), nil, nil
	case float32:
		return int64(typed), nil, nil
	case json.Number:
		if asInt, err := typed.Int64(); err == nil {
			return asInt, nil, nil
		}
		asFloat, err := typed.Float64()
		if err != nil {
			return 0, nil, err
		}
		return int64(asFloat), nil, nil
	case string:
		raw := strings.TrimSpace(typed)
		if raw == "" {
			return 0, nil, fmt.Errorf("empty date")
		}
		if millis, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return millis, nil, nil
		}
		layouts := []string{
			time.RFC3339Nano,
			"2006-01-02T15:04:05,000Z07:00",
			"2006-01-02T15:04:05,000-0700",
			"2006-01-02T15:04:05.000-0700",
			"2006-01-02T15:04:05-0700",
			"2006-01-02T15:04:05.000Z07:00",
			"2006-01-02T15:04:05Z07:00",
		}
		for _, layout := range layouts {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				_, offsetSeconds := parsed.Zone()
				offset := offsetSeconds / 60
				return parsed.UnixMilli(), &offset, nil
			}
		}
		return 0, nil, fmt.Errorf("bad date")
	default:
		return 0, nil, fmt.Errorf("bad date")
	}
}
