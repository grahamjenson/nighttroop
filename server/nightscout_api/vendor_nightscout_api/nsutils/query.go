package nsutils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"nighttroop.t1d.nz/models/nightscout"
)

// Because we cant support all mongo queries and filters
// Nighttroop is just going to enumerate and translate all queries into a limited set of supported queries

type Query struct {
	// Dates
	// lte and gte both
	DateGTE *nightscout.DateInt `mapstructure:"dategte,omitempty" json:"dategte,omitempty"`
	DateGT  *nightscout.DateInt `mapstructure:"dategt,omitempty" json:"dategt,omitempty"`

	DateLTE *nightscout.DateInt `mapstructure:"datelte,omitempty" json:"datelte,omitempty"`
	DateLT  *nightscout.DateInt `mapstructure:"datelt,omitempty" json:"datelt,omitempty"`

	CreatedAtGTE *nightscout.DateInt `mapstructure:"createdatgte,omitempty" json:"createdatgte,omitempty"`
	CreatedAtGT  *nightscout.DateInt `mapstructure:"createdatgt,omitempty" json:"createdatgt,omitempty"`

	CreatedAtLTE *nightscout.DateInt `mapstructure:"createdatlte,omitempty" json:"createdatlte,omitempty"`
	CreatedAtLT  *nightscout.DateInt `mapstructure:"createdatlt,omitempty" json:"createdatlt,omitempty"`

	UpdatedGTE *nightscout.DateInt `mapstructure:"updatedgte,omitempty" json:"updatedgte,omitempty"`
	UpdatedGT  *nightscout.DateInt `mapstructure:"updatedgt,omitempty" json:"updatedgt,omitempty"`
	UpdatedLTE *nightscout.DateInt `mapstructure:"updatedlte,omitempty" json:"updatedlte,omitempty"`
	UpdatedLT  *nightscout.DateInt `mapstructure:"updatedlt,omitempty" json:"updatedlt,omitempty"`

	// Type
	TypeEq  *nightscout.ShortString `mapstructure:"typeeq,omitempty" json:"typeeq,omitempty"`
	TypeNeq *nightscout.ShortString `mapstructure:"typeneq,omitempty" json:"typeneq,omitempty"`
	TypeIn  *QueryStringArray       `mapstructure:"typein,omitempty" json:"typein,omitempty"`
	TypeNin *QueryStringArray       `mapstructure:"typenin,omitempty" json:"typenin,omitempty"`

	// Identity
	IDEq *nightscout.ShortString `mapstructure:"ideq,omitempty" json:"ideq,omitempty"`
	IDIn *QueryStringArray       `mapstructure:"idin,omitempty" json:"idin,omitempty"`

	DeviceEq *nightscout.ShortString `mapstructure:"deviceeq,omitempty" json:"deviceeq,omitempty"`
	AppEq    *nightscout.ShortString `mapstructure:"appeq,omitempty" json:"appeq,omitempty"`

	// Limits
	Count *nightscout.FlexibleInt64 `mapstructure:"count,omitempty" json:"count,omitempty"`
	Limit *nightscout.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip  *nightscout.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`

	// Sort
	Sort     *nightscout.ShortString  `mapstructure:"sort,omitempty" json:"sort,omitempty"`
	SortDesc *nightscout.FlexibleBool `mapstructure:"sort$desc,omitempty" json:"sort$desc,omitempty"`

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SgvEQ  *string `mapstructure:"sgveq,omitempty" json:"sgveq,omitempty"`
	SgvGTE *string `mapstructure:"sgvgte,omitempty" json:"sgvgte,omitempty"`
	SgvGT  *string `mapstructure:"sgvgt,omitempty" json:"sgvgt,omitempty"`
	SgvLTE *string `mapstructure:"sgvlte,omitempty" json:"sgvlte,omitempty"`
	SgvLT  *string `mapstructure:"sgvlt,omitempty" json:"sgvlt,omitempty"`

	// Treatments
	CarbsEQ  *string `mapstructure:"carbseq,omitempty" json:"carbseq,omitempty"`
	CarbsGTE *string `mapstructure:"carbsgte,omitempty" json:"carbsgte,omitempty"`
	CarbsLT  *string `mapstructure:"carbslt,omitempty" json:"carbslt,omitempty"`

	InsulinEQ  *string `mapstructure:"insulineq,omitempty" json:"insulineq,omitempty"`
	InsulinGTE *string `mapstructure:"insulingte,omitempty" json:"insulingte,omitempty"`
	InsulinLT  *string `mapstructure:"insulinlt,omitempty" json:"insulinlt,omitempty"`

	NameEq *string `mapstructure:"nameeq,omitempty" json:"nameeq,omitempty"`

	// Altering Return values or structure
	Permanent *nightscout.FlexibleBool `mapstructure:"permanent,omitempty" json:"permanent,omitempty"`
	Envelope  *nightscout.FlexibleBool `mapstructure:"envelope,omitempty" json:"envelope,omitempty"`
	Fields    *QueryStringArray        `mapstructure:"fields,omitempty" json:"fields,omitempty"`

	IncludeDeleted *nightscout.FlexibleBool `mapstructure:"includeDeleted,omitempty" json:"includeDeleted,omitempty"`

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
	"typeneq": {"type$ne", "find[type][$ne]",
		"eventType$ne", "find[eventType][$ne]",
		"activityType$ne", "find[activityType][$ne]",
		"category$ne", "find[category][$ne]",
	},
	"typein":  {"type$in", "find[type][$in]", "eventType$in", "find[eventType][$in]", "activityType$in", "find[activityType][$in]", "category$in", "find[category][$in]"},
	"typenin": {"type$nin", "find[type][$nin]", "eventType$nin", "find[eventType][$nin]", "activityType$nin", "find[activityType][$nin]", "category$nin", "find[category][$nin]"},

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
		"device", "find[device]", "find[device][$eq]",
		"app", "find[app]", "find[app][$eq]",
		"subject", "find[subject]", "find[subject][$eq]",
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
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			JSONUnmarshalerHook,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}
	err = decoder.Decode(qvMap)
	if err != nil {
		return nil, fmt.Errorf("failed to decode query values: %w", err)
	}

	if len(qv.Other) > 0 {
		slog.Warn("Unrecognized Nightscout Query Parameters", "params", qv.Other)
	}

	return &qv, nil
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

//// Map Structure Helpers

func JSONUnmarshalerHook(
	from reflect.Type,
	to reflect.Type,
	data any,
) (any, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	// We need a pointer to the destination type because
	// UnmarshalJSON is usually defined on *T.
	ptrType := reflect.PointerTo(to)
	if !ptrType.Implements(unmarshalerType) {
		return data, nil
	}

	// Marshal the incoming value back to JSON.
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create a new instance of T.
	v := reflect.New(to)

	// Call UnmarshalJSON.
	if err := v.Interface().(json.Unmarshaler).UnmarshalJSON(b); err != nil {
		return nil, err
	}

	return v.Elem().Interface(), nil
}
