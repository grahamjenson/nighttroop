package nsutils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Filter represents a single field predicate in a storage query.
type Filter struct {
	Field string
	Op    string
	Value string
}

// Query describes the storage-level query contract used by the handlers.
type Query struct {
	Filters        []Filter
	Limit          int
	Skip           int
	SortField      string
	SortDesc       bool
	Fields         []string
	IncludeDeleted bool
}

// DefaultQuery returns the default query shape used by list-style endpoints.
func DefaultQuery() *Query {
	return &Query{
		Limit:          5,
		SortField:      "date",
		SortDesc:       true,
		IncludeDeleted: false,
	}
}

var bracePattern = regexp.MustCompile(`\{([^{}]+)\}`)

// ParseV1 translates Nightscout v1 query parameters into the storage query
// model used by this library.
func ParseV1(values url.Values, defaultDateField string) *Query {
	query := DefaultQuery()
	if defaultDateField == "" {
		defaultDateField = "date"
	}
	if count, err := strconv.Atoi(values.Get("count")); err == nil && count > 0 {
		query.Limit = count
	}
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		if !strings.HasPrefix(key, "find[") {
			continue
		}
		field, op := parseBracketFilter(key)
		if field == "" {
			continue
		}
		query.Filters = append(query.Filters, Filter{Field: field, Op: op, Value: vals[0]})
	}
	if rawFind := values.Get("find"); rawFind != "" {
		query.Filters = append(query.Filters, parseMongoFindJSON(rawFind)...)
	}
	if !hasIDFilter(query.Filters) && !hasDateFilter(query.Filters) {
		query.Filters = append(query.Filters, Filter{
			Field: defaultDateField,
			Op:    "gte",
			Value: time.Now().UTC().Add(-96 * time.Hour).Format(time.RFC3339),
		})
	}
	return query
}

// ParseV3 translates Nightscout API v3 query parameters into a storage query.
func ParseV3(values url.Values) (*Query, error) {
	query := DefaultQuery()
	if values.Has("sort") && values.Has("sort$desc") {
		return query, fmt.Errorf("Parameters sort and sort_desc cannot be combined")
	}
	if limitRaw := values.Get("limit"); limitRaw != "" {
		limit, err := strconv.Atoi(limitRaw)
		if err != nil || limit <= 0 {
			return query, fmt.Errorf("Parameter limit out of tolerance")
		}
		if limit > 1000 {
			return query, fmt.Errorf("Parameter limit out of tolerance")
		}
		query.Limit = limit
	}
	if skipRaw := values.Get("skip"); skipRaw != "" {
		skip, err := strconv.Atoi(skipRaw)
		if err != nil || skip < 0 {
			return query, fmt.Errorf("Parameter skip out of tolerance")
		}
		query.Skip = skip
	}
	if sortField := values.Get("sort"); sortField != "" {
		query.SortField = sortField
		query.SortDesc = false
	}
	if sortField := values.Get("sort$desc"); sortField != "" {
		query.SortField = sortField
		query.SortDesc = true
	}
	if fields := values.Get("fields"); fields != "" && fields != "_all" {
		for _, field := range strings.Split(fields, ",") {
			field = strings.TrimSpace(field)
			if field != "" {
				query.Fields = append(query.Fields, field)
			}
		}
	}
	for key, vals := range values {
		if len(vals) == 0 || isReservedV3Key(key) {
			continue
		}
		field, op := parseV3Filter(key)
		if field == "" {
			continue
		}
		query.Filters = append(query.Filters, Filter{Field: field, Op: op, Value: vals[0]})
	}
	return query, nil
}

// SelectFields projects a record to the requested fields.
func (q *Query) SelectFields(record Record) map[string]any {
	fields := q.Fields
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

// StripImplicitDateFilters removes the default rolling-window filters injected
// by ParseV1 so slice/times style endpoints can search across larger ranges.
func (q *Query) StripImplicitDateFilters() {
	filters := q.Filters
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

	q.Filters = out
}

// Slice filters records by prefix matching on a selected field and optionally a
// Nightscout entry type.
func Slice(records []Record, field, fieldType, prefix string, limit int) []Record {
	if prefix == "" {
		return records
	}
	patterns := ExpandBraces(prefix)
	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		value := fmt.Sprint(PathValue(record.ToMap(true), field))
		if fieldType != "" && fmt.Sprint(PathValue(record.ToMap(true), "type")) != fieldType {
			continue
		}
		if matchesAnyPrefix(value, patterns) {
			filtered = append(filtered, record)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left := PathValue(filtered[i].ToMap(true), "date")
		right := PathValue(filtered[j].ToMap(true), "date")
		return CompareValues(left, right, true)
	})
	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}
	return filtered
}

// Times filters records by modal time patterns built from a prefix and
// expression and returns both the filtered records and the generated patterns.
func Times(records []Record, prefix, expr string, limit int) ([]Record, []string) {
	prefixes := ExpandBraces(prefix)
	expressions := ExpandBraces(expr)
	patterns := make([]string, 0, len(prefixes)*max(1, len(expressions)))
	if len(expressions) == 0 {
		expressions = []string{""}
	}
	for _, p := range prefixes {
		for _, e := range expressions {
			switch {
			case p == "":
				patterns = append(patterns, e)
			case e == "":
				patterns = append(patterns, p)
			case strings.HasPrefix(e, ".*"):
				patterns = append(patterns, p+e)
			default:
				patterns = append(patterns, p+".*"+e)
			}
		}
	}
	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		dateString := fmt.Sprint(PathValue(record.ToMap(true), "dateString"))
		if dateString == "" {
			continue
		}
		if matchesAnyPrefix(dateString, patterns) {
			filtered = append(filtered, record)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left := PathValue(filtered[i].ToMap(true), "date")
		right := PathValue(filtered[j].ToMap(true), "date")
		return CompareValues(left, right, true)
	})
	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}
	return filtered, patterns
}

// ExpandBraces performs simple brace expansion used by the v1 slice/times
// query helpers.
func ExpandBraces(input string) []string {
	if input == "" {
		return []string{""}
	}
	if !strings.Contains(input, "{") {
		return []string{input}
	}
	match := bracePattern.FindStringSubmatchIndex(input)
	if match == nil {
		return []string{input}
	}
	body := input[match[2]:match[3]]
	parts := expandBody(body)
	prefix := input[:match[0]]
	suffix := input[match[1]:]
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, expanded := range ExpandBraces(suffix) {
			out = append(out, prefix+part+expanded)
		}
	}
	return out
}

func parseBracketFilter(key string) (field, op string) {
	key = strings.TrimPrefix(key, "find[")
	parts := strings.Split(key, "][")
	if len(parts) == 0 {
		return "", ""
	}
	field = strings.TrimSuffix(parts[0], "]")
	op = "eq"
	if len(parts) > 1 {
		op = strings.TrimPrefix(strings.TrimSuffix(parts[1], "]"), "$")
	}
	return field, op
}

func parseV3Filter(key string) (field, op string) {
	if strings.Contains(key, "$") {
		parts := strings.SplitN(key, "$", 2)
		return parts[0], parts[1]
	}
	return key, "eq"
}

func parseMongoFindJSON(raw string) []Filter {
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil
	}
	var filters []Filter
	for field, value := range doc {
		if nested, ok := value.(map[string]any); ok {
			for op, subValue := range nested {
				filters = append(filters, Filter{Field: field, Op: strings.TrimPrefix(op, "$"), Value: fmt.Sprint(subValue)})
			}
			continue
		}
		filters = append(filters, Filter{Field: field, Op: "eq", Value: fmt.Sprint(value)})
	}
	return filters
}

func isReservedV3Key(key string) bool {
	switch key {
	case "limit", "skip", "sort", "sort$desc", "fields", "token":
		return true
	default:
		return false
	}
}

func hasDateFilter(filters []Filter) bool {
	for _, filter := range filters {
		if filter.Field == "date" || filter.Field == "dateString" || filter.Field == "created_at" {
			return true
		}
	}
	return false
}

func hasIDFilter(filters []Filter) bool {
	for _, filter := range filters {
		if filter.Field == "_id" || filter.Field == "identifier" {
			return true
		}
	}
	return false
}

func expandBody(body string) []string {
	if strings.Contains(body, "..") {
		parts := strings.SplitN(body, "..", 2)
		start := parts[0]
		end := parts[1]
		startInt, err1 := strconv.Atoi(start)
		endInt, err2 := strconv.Atoi(end)
		if err1 == nil && err2 == nil {
			width := len(start)
			out := make([]string, 0, endInt-startInt+1)
			for i := startInt; i <= endInt; i++ {
				out = append(out, fmt.Sprintf("%0*d", width, i))
			}
			return out
		}
	}
	return strings.Split(body, ",")
}

func matchesAnyPrefix(value string, patterns []string) bool {
	for _, pattern := range patterns {
		regex := prefixPatternToRegexp(pattern)
		if regex.MatchString(value) {
			return true
		}
	}
	return false
}

func prefixPatternToRegexp(pattern string) *regexp.Regexp {
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, `\.\*`, ".*")
	regex := "^" + quoted
	return regexp.MustCompile(regex)
}

// ApplyQuery filters, sorts, skips, and limits a record set in memory.
func ApplyQuery(records []Record, query *Query) []Record {
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
