package nsutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// StripImplicitDateFilters removes the default rolling-window filters injected
// by ParseV1 so slice/times style endpoints can search across larger ranges.
func StripImplicitDateFilters(filters []Filter) []Filter {
	out := make([]Filter, 0, len(filters))
	for _, filter := range filters {
		if filter.Field == "date" && filter.Op == "gte" && strings.Contains(filter.Value, "T") {
			continue
		}
		if filter.Field == "created_at" && filter.Op == "gte" && strings.Contains(filter.Value, "T") {
			continue
		}
		out = append(out, filter)
	}
	return out
}

// WriteRecords serializes records using the public Nightscout response shape.
func WriteRecords(w http.ResponseWriter, records []Record, fields []string) {
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if len(fields) == 0 {
			response = append(response, record.ToMap(false))
		} else {
			response = append(response, SelectFields(record, fields))
		}
	}
	WriteJSON(w, http.StatusOK, response)
}

// ParseTimesWildcard extracts the `prefix/expr` wildcard payload used by the
// v1 /times endpoints.
func ParseTimesWildcard(r *http.Request) (prefix, expr string, ok bool) {
	wild := chi.URLParam(r, "*")
	parts := strings.SplitN(wild, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	prefix = parts[0]
	expr = strings.TrimSuffix(parts[1], ".json")
	if decoded, err := url.PathUnescape(prefix); err == nil {
		prefix = decoded
	}
	if decoded, err := url.PathUnescape(expr); err == nil {
		expr = decoded
	}
	if prefix == "" || expr == "" {
		return "", "", false
	}
	return prefix, expr, true
}

// NormalizeData normalizes incoming Nightscout payloads into a consistent shape
// before persistence.
func NormalizeData(collection Collection, data map[string]any) (map[string]any, error) {
	clean := CloneMap(data)

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

	if rawDate, exists := clean["date"]; exists && rawDate != nil {
		millis, offset, err := parseDateValue(rawDate)
		if err != nil {
			return nil, fmt.Errorf("bad or missing date field")
		}
		clean["date"] = millis
		if _, exists := clean["utcOffset"]; !exists && offset != nil {
			clean["utcOffset"] = *offset
		}
	}
	if value, ok := StringField(clean, "created_at"); ok && value != "" {
		normalized, offset, err := ToUTCString(value)
		if err != nil {
			return nil, fmt.Errorf("bad created_at field")
		}
		clean["created_at"] = normalized
		if _, exists := clean["utcOffset"]; !exists {
			clean["utcOffset"] = offset
		}
	}
	if value, ok := Int64Field(clean, "date"); ok {
		if value <= 100000000000 {
			return nil, fmt.Errorf("bad or missing date field")
		}
	} else if value, ok := StringField(clean, "created_at"); ok && value != "" {
		parsed, _ := time.Parse("2006-01-02T15:04:05.000Z", value)
		if !parsed.IsZero() {
			clean["date"] = parsed.UnixMilli()
		}
	}
	if clean["date"] == nil {
		if collection == Entries || collection == Treatments || collection == Devicestatus || collection == Food {
			return nil, fmt.Errorf("bad or missing date field")
		}
	} else if _, ok := StringField(clean, "created_at"); !ok && collection != Entries && collection != Settings {
		if date, ok := Int64Field(clean, "date"); ok {
			clean["created_at"] = time.UnixMilli(date).UTC().Format("2006-01-02T15:04:05.000Z")
		}
	}
	if offset, exists := clean["utcOffset"]; exists && offset != nil {
		num, err := toInt64(offset)
		if err != nil || num < -1440 || num > 1440 {
			return nil, fmt.Errorf("bad or missing utcOffset field")
		}
		clean["utcOffset"] = int(num)
	}
	if notes, ok := StringField(clean, "notes"); ok {
		clean["notes"] = sanitizeNotes(notes)
	}
	if _, ok := StringField(clean, "type"); !ok && collection == Entries {
		clean["type"] = "sgv"
	}
	if _, ok := StringField(clean, "identifier"); !ok {
		if existingID, ok := StringField(clean, "_id"); ok && existingID != "" {
			clean["identifier"] = existingID
		}
	}
	return clean, nil
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

func normalizeNumericFields(data map[string]any, fields ...string) error {
	for _, field := range fields {
		value, ok := data[field]
		if !ok || value == nil {
			continue
		}
		number, err := toFloat64(value)
		if err != nil {
			return fmt.Errorf("bad numeric field %s", field)
		}
		if number == float64(int64(number)) {
			data[field] = int64(number)
			continue
		}
		data[field] = number
	}
	return nil
}
