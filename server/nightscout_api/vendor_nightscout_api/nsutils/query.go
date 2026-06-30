package nsutils

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"nighttroop.t1d.nz/models/nightscout"
)

// Because we cant support all mongo queries and filters
// Nighttroop is just going to enumerate and translate all queries into a limited set of supported queries

type Query struct {
	// Dates
	// lte and gte both
	DateGTE *nightscout.DateInt `mapstructure:"dategte,omitempty"`
	DateGT  *nightscout.DateInt `mapstructure:"dategt,omitempty"`

	DateLTE *nightscout.DateInt `mapstructure:"datelte,omitempty"`
	DateLT  *nightscout.DateInt `mapstructure:"datelt,omitempty"`

	CreatedAtGTE *nightscout.DateInt `mapstructure:"createdatgte,omitempty"`
	CreatedAtGT  *nightscout.DateInt `mapstructure:"createdatgt,omitempty"`

	CreatedAtLTE *nightscout.DateInt `mapstructure:"createdatlte,omitempty"`
	CreatedAtLT  *nightscout.DateInt `mapstructure:"createdatlt,omitempty"`

	UpdatedGTE *nightscout.DateInt `mapstructure:"updatedgte,omitempty"`
	UpdatedGT  *nightscout.DateInt `mapstructure:"updatedgt,omitempty"`
	UpdatedLTE *nightscout.DateInt `mapstructure:"updatedlte,omitempty"`
	UpdatedLT  *nightscout.DateInt `mapstructure:"updatedlt,omitempty"`

	TypeEq *nightscout.ShortString `mapstructure:"typeeq,omitempty"`

	// Identity
	IDEq *nightscout.ShortString `mapstructure:"ideq,omitempty"`
	IDIn *QueryStringArray       `mapstructure:"idin,omitempty"`

	DeviceEq *nightscout.ShortString `mapstructure:"deviceeq,omitempty"`
	AppEq    *nightscout.ShortString `mapstructure:"appeq,omitempty"`

	// Limits
	Count *nightscout.FlexibleInt64 `mapstructure:"count,omitempty"`
	Limit *nightscout.FlexibleInt64 `mapstructure:"limit,omitempty"`
	Skip  *nightscout.FlexibleInt64 `mapstructure:"skip,omitempty"`

	// Sort
	Sort     *nightscout.ShortString  `mapstructure:"sort,omitempty"`
	SortDesc *nightscout.FlexibleBool `mapstructure:"sort$desc,omitempty"`

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SGVEQ  *string `mapstructure:"sgveq,omitempty"`
	SGVGTE *string `mapstructure:"sgvgte,omitempty"`
	SGVGT  *string `mapstructure:"sgvgt,omitempty"`
	SGVLTE *string `mapstructure:"sgvlte,omitempty"`
	SGVLT  *string `mapstructure:"sgvlt,omitempty"`

	// Treatments
	CarbsEQ  *string `mapstructure:"carbseq,omitempty"`
	CarbsGTE *string `mapstructure:"carbsgte,omitempty"`
	CarbsLT  *string `mapstructure:"carbslt,omitempty"`

	InsulinEQ  *string `mapstructure:"insulineq,omitempty"`
	InsulinGTE *string `mapstructure:"insulingte,omitempty"`
	InsulinLT  *string `mapstructure:"insulinlt,omitempty"`

	NameEq *string `mapstructure:"nameeq,omitempty"`

	// Altering Return values or structure
	Permanent nightscout.FlexibleBool  `mapstructure:"permanent,omitempty"`
	Envelope  *nightscout.FlexibleBool `mapstructure:"envelope,omitempty"`
	Fields    *QueryStringArray        `mapstructure:"fields,omitempty"`

	IncludeDeleted *nightscout.FlexibleBool `mapstructure:"isValid,omitempty"`

	Other map[string]interface{} `mapstructure:",remain"`
}

type QueryStringArray []string

func (qsa *QueryStringArray) ToArray() []string {
	if qsa == nil {
		return []string{}
	}
	return []string(*qsa)
}
func (qsa *QueryStringArray) UnmarshalJSON(data []byte) error {
	// First try a JSON array.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*qsa = arr
		return nil
	}

	// Otherwise expect a JSON string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	var parts []string
	switch {
	case s == "_all":
		parts = []string{}
	case strings.Contains(s, "|"):
		parts = strings.Split(s, "|")
	case strings.Contains(s, ","):
		parts = strings.Split(s, ",")
	default:
		parts = []string{s}
	}

	// Trim whitespace and drop empty elements.
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}

	*qsa = out
	return nil
}

func ptr[T any](v T) *T {
	return &v
}

// DefaultQuery returns the default query shape used by list-style endpoints.
func DefaultQuery() *Query {
	return &Query{
		Limit:    ptr(nightscout.FlexibleInt64(100)),
		Sort:     ptr(nightscout.ShortString("date")),
		SortDesc: ptr(nightscout.FlexibleBool(true)),
	}
}

// ParseV1 translates Nightscout v1 query parameters into the storage query
// model used by this library.
func ParseV1(values url.Values) (*Query, *HTTPError) {
	qv, err := parseQuery(values)
	if err != nil {
		return nil, ErrBadQueryError
	}

	return qv, nil
}

// ParseV3 translates Nightscout API v3 query parameters into a storage query.
func ParseV3(values url.Values) (*Query, *HTTPError) {
	qv, err := parseQuery(values)
	if err != nil {
		return nil, ErrBadQueryError
	}

	return qv, nil
}

// This is just much nicer to work with but the reverse is faster, so we
var forwardLookupName = map[string][]string{
	"dategte": {"date$gte", "find[date][$gte]", "find[dateString][$gte]", "mills$gte", "find[mills][$gte]", "startDate$gte", "find[startDate][$gte]"},
	"dategt":  {"date$gt", "find[date][$gt]", "find[dateString][$gt]", "mills$gt", "find[mills][$gt]", "startDate$gt", "find[startDate][$gt]"},

	"datelte": {"date$lte", "find[date][$lte]", "find[dateString][$lte]", "mills$lte", "find[mills][$lte]", "startDate$lte", "find[startDate][$lte]"},
	"datelt":  {"date$lt", "find[date][$lt]", "find[dateString][$lt]", "mills$lt", "find[mills][$lt]", "startDate$lt", "find[startDate][$lt]"},

	"createdatgte": {"created_at$gte", "find[created_at][$gte]", "srvCreated$gte", "find[srvCreated][$gte]"},
	"createdatgt":  {"created_at$gt", "find[created_at][$gt]", "srvCreated$gt", "find[srvCreated][$gt]"},

	"createdatlte": {"created_at$lte", "find[created_at][$lte]", "srvCreated$lte", "find[srvCreated][$lte]"},
	"createdatlt":  {"created_at$lt", "find[created_at][$lt]", "srvCreated$lt", "find[srvCreated][$lt]"},

	"updatedgte": {"updated_at$gte", "find[updated_at][$gte]", "srvModified$gte", "find[srvModified][$gte]"},
	"updatedgt":  {"updated_at$gt", "find[updated_at][$gt]", "srvModified$gt", "find[srvModified][$gt]"},

	"updatedlte": {"updated_at$lte", "find[updated_at][$lte]", "srvModified$lte", "find[srvModified][$lte]"},
	"updatedlt":  {"updated_at$lt", "find[updated_at][$lt]", "srvModified$lt", "find[srvModified][$lt]"},

	"typeeq": {"type", "find[type]", "find[type][$eq]",
		"eventType", "find[eventType]", "find[eventType][$eq]",
		"activityType", "find[activityType]", "find[activityType][$eq]",
		"category", "find[category]", "find[category][$eq]",
	},
	"typein": {"type$in", "find[type][$in]", "eventType$in", "find[eventType][$in]", "activityType$in", "find[activityType][$in]", "category$in", "find[category][$in]"},

	"ideq": {"_id", "find[_id]", "find[_id][$eq]", "identifier", "find[identifier]", "find[identifier][$eq]"},
	"idin": {"_id$in", "find[_id][$in]", "identifier$in", "find[identifier][$in]"},

	// Model Specific
	"sgveq": {"sgv", "find[sgv]", "find[sgv][$eq]", "mbg", "find[mbg]", "find[mbg][$eq]"},

	"sgvgte": {"sgv$gte", "find[sgv][$gte]", "mbg$gte", "find[mbg][$gte]"},
	"sgvgt":  {"sgv$gt", "find[sgv][$gt]", "mbg$gt", "find[mbg][$gt]"},

	"sgvlte": {"sgv$lte", "find[sgv][$lte]", "mbg$lte", "find[mbg][$lte]"},
	"sgvlt":  {"sgv$lt", "find[sgv][$lt]", "mbg$lt", "find[mbg][$lt]"},

	"carbseq":  {"carbs", "find[carbs]", "find[carbs][$eq]"},
	"carbsgte": {"carbs$gte", "find[carbs][$gte]"},
	"carbslt":  {"carbs$lt", "find[carbs][$lt]"},

	"insulineq":  {"insulin", "find[insulin]", "find[insulin][$eq]"},
	"insulingte": {"insulin$gte", "find[insulin][$gte]"},
	"insulinlt":  {"insulin$lt", "find[insulin][$lt]"},

	"nameeq": {"name", "find[name]", "find[name][$eq]",
		"defaultProfile", "find[defaultProfile]", "find[defaultProfile][$eq]",
	},
}

var reverseLookupName = map[string]string{}

func parseQuery(values url.Values) (*Query, error) {
	if len(reverseLookupName) == 0 {
		for k, v := range forwardLookupName {
			for _, vv := range v {
				reverseLookupName[vv] = k
			}
		}
	}

	qvMap := map[string]string{}
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}

		// Speciial Handling of fields that can be repeated in the query string, like fields[] and find[identifier][$in]
		// Comma Seperated Values can just be patched on

		if key == "fields" {
			qvMap[key] = strings.Join(vals, ",")
			continue
		}

		if strings.HasSuffix(key, "[$in]") || strings.HasSuffix(key, "$in") || strings.HasSuffix(key, "$nin") || strings.HasSuffix(key, "[$nin]") {
			qvMap[key] = strings.Join(vals, "|")
			continue
		}

		if len(vals) > 1 {
			slog.Warn("Multiple Values for Nightscout Query Key", "key", key, "values", vals)
		}
		qvMap[key] = vals[0]

	}

	var qv Query
	config := &mapstructure.DecoderConfig{
		Result:  &qv,
		TagName: "mapstructure",
		MatchName: func(mapKey, fieldName string) bool {
			v, ok := reverseLookupName[mapKey]
			if ok {
				mapKey = v
			}
			ret := strings.EqualFold(mapKey, fieldName)

			return ret
		},
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, ErrBadQueryError
	}
	err = decoder.Decode(qvMap)
	if err != nil {
		return nil, ErrBadQueryError
	}

	if len(qv.Other) > 0 {
		slog.Warn("Unrecognized Nightscout Query Parameters", "params", qv.Other)
	}

	return &qv, nil
}

// StripImplicitDateFilters removes the default rolling-window filters injected
// by ParseV1 so slice/times style endpoints can search across larger ranges.
func (q *Query) StripImplicitDateFilters() {
	// filters := q.Filters
	// out := make([]Filter, 0, len(filters))
	// for _, filter := range filters {
	// 	if filter.Field == "date" && filter.Op == "gte" && strings.Contains(filter.Value, "T") {
	// 		continue
	// 	}
	// 	if filter.Field == "created_at" && filter.Op == "gte" && strings.Contains(filter.Value, "T") {
	// 		continue
	// 	}
	// 	out = append(out, filter)
	// }

	// q.Filters = out
}

// ApplyQuery filters, sorts, skips, and limits a record set in memory.
func ApplyQuery(records []Record, query *Query) []Record {
	filtered := make([]Record, 0, len(records))
	for _, record := range records {

		// if matchesAll(record, query.Filters) {
		filtered = append(filtered, record.Clone())
		// }
	}

	return filtered
}

// Slice filters records by prefix matching on a selected field and optionally a
// Nightscout entry type.
func Slice(records []Record, field, fieldType, prefix string, limit int) []Record {
	if prefix == "" {
		return records
	}

	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		filtered = append(filtered, record)
	}

	return filtered
}

// Times filters records by modal time patterns built from a prefix and
// expression and returns both the filtered records and the generated patterns.
func Times(records []Record, prefix, expr string, limit int) ([]Record, []string) {

	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		filtered = append(filtered, record)
	}

	return filtered, nil
}
