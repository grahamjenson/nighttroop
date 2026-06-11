package nsutils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrGone     = errors.New("record deleted")
)

// Store abstracts persistence for Nightscout collections.
//
// Applications are expected to provide their own implementation.
type Store interface {
	Create(ctx context.Context, collection Collection, data map[string]any, subject string) (Record, bool, error)
	Get(ctx context.Context, collection Collection, identifier string) (Record, error)
	Search(ctx context.Context, collection Collection, query Query) ([]Record, error)
	Update(ctx context.Context, collection Collection, identifier string, patch map[string]any, subject string) (Record, error)
	Delete(ctx context.Context, collection Collection, identifier string, permanent bool, subject string) error
	DeleteMatching(ctx context.Context, collection Collection, query Query, permanent bool, subject string) (int, error)
}

// DefaultQuery returns the default query shape used by list-style endpoints.
func DefaultQuery() Query {
	return Query{
		Limit:          5,
		SortField:      "date",
		SortDesc:       true,
		IncludeDeleted: false,
	}
}

// ApplyQuery filters, sorts, skips, and limits a record set in memory.
func ApplyQuery(records []Record, query Query) []Record {
	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if !query.IncludeDeleted && !record.IsValid {
			continue
		}
		if matchesAll(record, query.Filters) {
			filtered = append(filtered, record.Clone())
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left := filtered[i].ToMap(true)
		right := filtered[j].ToMap(true)
		return CompareValues(left[query.SortField], right[query.SortField], query.SortDesc)
	})
	start := min(query.Skip, len(filtered))
	filtered = filtered[start:]
	if query.Limit > 0 && query.Limit < len(filtered) {
		filtered = filtered[:query.Limit]
	}
	return filtered
}

// SelectFields projects a record to the requested fields.
func SelectFields(record Record, fields []string) map[string]any {
	full := record.ToMap(true)
	if len(fields) == 0 {
		return full
	}
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := full[field]; ok {
			out[field] = value
		}
	}
	if len(fields) > 0 {
		delete(out, "_id")
	}
	return out
}

func matchesAll(record Record, filters []Filter) bool {
	full := record.ToMap(true)
	for _, filter := range filters {
		if !matchFilter(PathValue(full, filter.Field), filter) {
			return false
		}
	}
	return true
}

func matchFilter(value any, filter Filter) bool {
	switch filter.Field {
	case "dateString":
		if millis, ok := toInt64(value); ok == nil {
			value = time.UnixMilli(millis).UTC().Format("2006-01-02T15:04:05")
		}
	}
	if filter.Op == "exists" {
		want := strings.EqualFold(filter.Value, "true")
		return (value != nil) == want
	}
	if filter.Op == "re" || isRegexLiteral(filter.Value) {
		pattern, regexErr := compileRegexFilter(filter.Value)
		if regexErr == nil {
			matched := pattern.MatchString(strings.TrimSpace(fmt.Sprint(value)))
			if filter.Op == "ne" || filter.Op == "nin" {
				return !matched
			}
			return matched
		}
	}
	leftNum, leftNumErr := toFloat64(value)
	rightNum, rightNumErr := strconv.ParseFloat(filter.Value, 64)
	if leftNumErr == nil && rightNumErr == nil {
		switch filter.Op {
		case "eq":
			return leftNum == rightNum
		case "ne":
			return leftNum != rightNum
		case "gte":
			return leftNum >= rightNum
		case "lte":
			return leftNum <= rightNum
		case "gt":
			return leftNum > rightNum
		case "lt":
			return leftNum < rightNum
		case "in":
			return inNumericSet(leftNum, filter.Value)
		case "nin":
			return !inNumericSet(leftNum, filter.Value)
		}
	}
	leftTime, leftTimeErr := toTime(value)
	rightTime, rightTimeErr := toTime(filter.Value)
	if leftTimeErr == nil && rightTimeErr == nil {
		switch filter.Op {
		case "eq":
			return leftTime.Equal(rightTime)
		case "ne":
			return !leftTime.Equal(rightTime)
		case "gte":
			return leftTime.After(rightTime) || leftTime.Equal(rightTime)
		case "lte":
			return leftTime.Before(rightTime) || leftTime.Equal(rightTime)
		case "gt":
			return leftTime.After(rightTime)
		case "lt":
			return leftTime.Before(rightTime)
		}
	}
	left := strings.TrimSpace(fmt.Sprint(value))
	right := strings.TrimSpace(filter.Value)
	switch filter.Op {
	case "eq":
		return left == right
	case "ne":
		return left != right
	case "gte":
		return left >= right
	case "lte":
		return left <= right
	case "gt":
		return left > right
	case "lt":
		return left < right
	case "in":
		return inStringSet(left, right)
	case "nin":
		return !inStringSet(left, right)
	default:
		return false
	}
}

// CompareValues compares values using numeric, time, then string semantics.
func CompareValues(left, right any, desc bool) bool {
	if leftNum, err := toFloat64(left); err == nil {
		if rightNum, err := toFloat64(right); err == nil {
			if desc {
				return leftNum > rightNum
			}
			return leftNum < rightNum
		}
	}
	if leftTime, err := toTime(left); err == nil {
		if rightTime, err := toTime(right); err == nil {
			if desc {
				return leftTime.After(rightTime)
			}
			return leftTime.Before(rightTime)
		}
	}
	leftString := fmt.Sprint(left)
	rightString := fmt.Sprint(right)
	if desc {
		return leftString > rightString
	}
	return leftString < rightString
}

func toFloat64(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case int:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case int32:
		return float64(typed), nil
	case json.Number:
		return typed.Float64()
	case string:
		return strconv.ParseFloat(typed, 64)
	default:
		return 0, fmt.Errorf("not numeric")
	}
}

func toInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case float32:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("not int")
	}
}

func toTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed, nil
	case int64:
		return time.UnixMilli(typed), nil
	case int:
		return time.UnixMilli(int64(typed)), nil
	case float64:
		return time.UnixMilli(int64(typed)), nil
	case string:
		layouts := []string{
			time.RFC3339Nano,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02",
			"2006-01",
		}
		for _, layout := range layouts {
			parsed, err := time.Parse(layout, typed)
			if err == nil {
				return parsed, nil
			}
		}
		if millis, err := strconv.ParseInt(typed, 10, 64); err == nil {
			return time.UnixMilli(millis), nil
		}
		return time.Time{}, fmt.Errorf("not time")
	default:
		return time.Time{}, fmt.Errorf("not time")
	}
}

func isRegexLiteral(value string) bool {
	return strings.HasPrefix(value, "/") && strings.Count(value, "/") >= 2
}

func compileRegexFilter(value string) (*regexp.Regexp, error) {
	if isRegexLiteral(value) {
		last := strings.LastIndex(value, "/")
		body := value[1:last]
		flags := value[last+1:]
		if strings.Contains(flags, "i") {
			body = "(?i)" + body
		}
		return regexp.Compile(body)
	}
	return regexp.Compile(value)
}

func inStringSet(value, raw string) bool {
	for _, part := range strings.Split(raw, "|") {
		if value == part {
			return true
		}
	}
	return false
}

func inNumericSet(value float64, raw string) bool {
	for _, part := range strings.Split(raw, "|") {
		if parsed, err := strconv.ParseFloat(part, 64); err == nil && parsed == value {
			return true
		}
	}
	return false
}
