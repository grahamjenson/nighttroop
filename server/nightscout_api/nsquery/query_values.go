package nsquery

import (
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"nighttroop.t1d.nz/models/nightscout"
)

type QueryValues struct {
	// Dates
	// lte and gte both
	DateGTE *nightscout.DateInt `mapstructure:"dategte,omitempty" json:"dategte,omitempty"`
	DateGT  *nightscout.DateInt `mapstructure:"dategt,omitempty" json:"dategt,omitempty"`

	DateEQ *nightscout.DateInt `mapstructure:"dateeq,omitempty" json:"dateeq,omitempty"`

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

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SgvEQ  *nightscout.FlexibleInt64 `mapstructure:"sgveq,omitempty" json:"sgveq,omitempty"`
	SgvGTE *nightscout.FlexibleInt64 `mapstructure:"sgvgte,omitempty" json:"sgvgte,omitempty"`
	SgvGT  *nightscout.FlexibleInt64 `mapstructure:"sgvgt,omitempty" json:"sgvgt,omitempty"`
	SgvLTE *nightscout.FlexibleInt64 `mapstructure:"sgvlte,omitempty" json:"sgvlte,omitempty"`
	SgvLT  *nightscout.FlexibleInt64 `mapstructure:"sgvlt,omitempty" json:"sgvlt,omitempty"`

	// Treatments
	CarbsEQ  *nightscout.FlexibleInt64 `mapstructure:"carbseq,omitempty" json:"carbseq,omitempty"`
	CarbsGTE *nightscout.FlexibleInt64 `mapstructure:"carbsgte,omitempty" json:"carbsgte,omitempty"`
	CarbsGT  *nightscout.FlexibleInt64 `mapstructure:"carbsgt,omitempty" json:"carbsgt,omitempty"`
	CarbsLTE *nightscout.FlexibleInt64 `mapstructure:"carbslte,omitempty" json:"carbslte,omitempty"`
	CarbsLT  *nightscout.FlexibleInt64 `mapstructure:"carbslt,omitempty" json:"carbslt,omitempty"`

	InsulinEQ  *nightscout.FlexibleInt64 `mapstructure:"insulineq,omitempty" json:"insulineq,omitempty"`
	InsulinGTE *nightscout.FlexibleInt64 `mapstructure:"insulingte,omitempty" json:"insulingte,omitempty"`
	InsulinGT  *nightscout.FlexibleInt64 `mapstructure:"insulingt,omitempty" json:"insulingt,omitempty"`
	InsulinLTE *nightscout.FlexibleInt64 `mapstructure:"insulinlte,omitempty" json:"insulinlte,omitempty"`
	InsulinLT  *nightscout.FlexibleInt64 `mapstructure:"insulinlt,omitempty" json:"insulinlt,omitempty"`

	DurationEQ  *nightscout.FlexibleInt64 `mapstructure:"durationeq,omitempty" json:"durationeq,omitempty"`
	DurationGTE *nightscout.FlexibleInt64 `mapstructure:"durationgte,omitempty" json:"durationgte,omitempty"`
	DurationGT  *nightscout.FlexibleInt64 `mapstructure:"durationgt,omitempty" json:"durationgt,omitempty"`
	DurationLTE *nightscout.FlexibleInt64 `mapstructure:"durationlte,omitempty" json:"durationlte,omitempty"`
	DurationLT  *nightscout.FlexibleInt64 `mapstructure:"durationlt,omitempty" json:"durationlt,omitempty"`

	NameEq *nightscout.ShortString `mapstructure:"nameeq,omitempty" json:"nameeq,omitempty"`

	// Altering Return values or structure
	// Limits
	Limit *nightscout.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip  *nightscout.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`

	// Sort
	Sort     *nightscout.ShortString `mapstructure:"sort,omitempty" json:"sort,omitempty"`
	SortDesc *nightscout.ShortString `mapstructure:"sort$desc,omitempty" json:"sort$desc,omitempty"`

	Envelope *nightscout.FlexibleBool `mapstructure:"envelope,omitempty" json:"envelope,omitempty"`
	Fields   *QueryStringArray        `mapstructure:"fields,omitempty" json:"fields,omitempty"`

	IncludeDeleted *nightscout.FlexibleBool `mapstructure:"includeDeleted,omitempty" json:"includeDeleted,omitempty"`

	// This is Nighttroops value to show RawJsonData
	Raw *nightscout.FlexibleBool `mapstructure:"raw,omitempty" json:"raw,omitempty"`

	Token *string `mapstructure:"token,omitempty" json:"-"`

	Other map[string]interface{} `mapstructure:",remain"`
}

// This is just much nicer to work with but the reverse is faster, so we
var forwardLookupName = map[string][]string{
	"dategte": {"date$gte", "find[date][$gte]", "find[dateString][$gte]", "mills$gte", "find[mills][$gte]", "startDate$gte", "find[startDate][$gte]", "start$gte", "find[start][$gte]"},
	"dategt":  {"date$gt", "find[date][$gt]", "find[dateString][$gt]", "mills$gt", "find[mills][$gt]", "startDate$gt", "find[startDate][$gt]", "start$gt", "find[start][$gt]"},

	"dateeq": {"date$eq", "find[date][$eq]", "find[dateString][$eq]", "mills$eq", "find[mills][$eq]", "startDate$eq", "find[startDate][$eq]", "start$eq", "find[start][$eq]"},

	"datelte": {"date$lte", "find[date][$lte]", "find[dateString][$lte]", "mills$lte", "find[mills][$lte]", "startDate$lte", "find[startDate][$lte]", "start$lte", "find[start][$lte]"},
	"datelt":  {"date$lt", "find[date][$lt]", "find[dateString][$lt]", "mills$lt", "find[mills][$lt]", "startDate$lt", "find[startDate][$lt]", "start$lt", "find[start][$lt]"},

	"createdatgte": {"created_at$gte", "find[created_at][$gte]", "srvCreated$gte", "find[srvCreated][$gte]"},
	"createdatgt":  {"created_at$gt", "find[created_at][$gt]", "srvCreated$gt", "find[srvCreated][$gt]"},

	"createdatlte": {"created_at$lte", "find[created_at][$lte]", "srvCreated$lte", "find[srvCreated][$lte]"},
	"createdatlt":  {"created_at$lt", "find[created_at][$lt]", "srvCreated$lt", "find[srvCreated][$lt]"},

	"updatedgte": {"updated_at$gte", "find[updated_at][$gte]", "srvModified$gte", "find[srvModified][$gte]"},
	"updatedgt":  {"updated_at$gt", "find[updated_at][$gt]", "srvModified$gt", "find[srvModified][$gt]"},

	"updatedlte": {"updated_at$lte", "find[updated_at][$lte]", "srvModified$lte", "find[srvModified][$lte]"},
	"updatedlt":  {"updated_at$lt", "find[updated_at][$lt]", "srvModified$lt", "find[srvModified][$lt]"},

	"durationeq":  {"duration", "find[duration]", "find[duration][$eq]", "durationMinutes", "find[durationMinutes]", "find[durationMinutes][$eq]"},
	"durationgte": {"duration$gte", "find[duration][$gte]", "durationMinutes$gte", "find[durationMinutes][$gte]"},
	"durationgt":  {"duration$gt", "find[duration][$gt]", "durationMinutes$gt", "find[durationMinutes][$gt]"},
	"durationlte": {"duration$lte", "find[duration][$lte]", "durationMinutes$lte", "find[durationMinutes][$lte]"},
	"durationlt":  {"duration$lt", "find[duration][$lt]", "durationMinutes$lt", "find[durationMinutes][$lt]"},

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

	"ideq": {"_id", "find[_id]", "find[_id][$eq]", "identifier", "find[identifier]", "find[identifier][$eq]", "id", "find[id]", "find[id][$eq]"},
	"idin": {"_id$in", "find[_id][$in]", "identifier$in", "find[identifier][$in]", "id$in", "find[id][$in]"},

	// Model Specific
	"sgveq": {"sgv", "find[sgv]", "find[sgv][$eq]", "mbg", "find[mbg]", "find[mbg][$eq]"},

	"sgvgte": {"sgv$gte", "find[sgv][$gte]", "mbg$gte", "find[mbg][$gte]"},
	"sgvgt":  {"sgv$gt", "find[sgv][$gt]", "mbg$gt", "find[mbg][$gt]"},

	"sgvlte": {"sgv$lte", "find[sgv][$lte]", "mbg$lte", "find[mbg][$lte]"},
	"sgvlt":  {"sgv$lt", "find[sgv][$lt]", "mbg$lt", "find[mbg][$lt]"},

	"carbseq":  {"carbs", "find[carbs]", "find[carbs][$eq]"},
	"carbsgte": {"carbs$gte", "find[carbs][$gte]"},
	"carbsgt":  {"carbs$gt", "find[carbs][$gt]"},
	"carbslte": {"carbs$lte", "find[carbs][$lte]"},
	"carbslt":  {"carbs$lt", "find[carbs][$lt]"},

	"insulineq":  {"insulin", "find[insulin]", "find[insulin][$eq]"},
	"insulingte": {"insulin$gte", "find[insulin][$gte]"},
	"insulingt":  {"insulin$gt", "find[insulin][$gt]"},
	"insulinlte": {"insulin$lte", "find[insulin][$lte]"},
	"insulinlt":  {"insulin$lt", "find[insulin][$lt]"},

	"nameeq": {"name", "find[name]", "find[name][$eq]",
		"defaultProfile", "find[defaultProfile]", "find[defaultProfile][$eq]",
		"device", "find[device]", "find[device][$eq]",
		"app", "find[app]", "find[app][$eq]",
		"subject", "find[subject]", "find[subject][$eq]",
	},

	"limit": {"limit", "count"},
}

var reverseLookupName = map[string]string{}

func ParseQueryValues(values url.Values) (QueryValues, error) {
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

	var qv QueryValues
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
		return QueryValues{}, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}
	err = decoder.Decode(qvMap)
	if err != nil {
		return QueryValues{}, fmt.Errorf("failed to decode query values: %w", err)
	}

	if qv.Sort != nil && qv.SortDesc != nil {
		return QueryValues{}, fmt.Errorf("cannot specify both sort and sort$desc parameters")
	}

	if len(qv.Other) > 0 {
		slog.Warn("Unrecognized Nightscout Query Parameters", "params", qv.Other)
	}

	return qv, nil
}

func (q *QueryValues) SortOrder() *QuerySort {
	if q == nil {
		return nil
	}
	if q.Sort == nil && q.SortDesc == nil {
		return nil
	}

	sortField := "occurred_at"
	desc := false
	if q.Sort != nil {
		sortField = q.Sort.String()
	}
	if q.SortDesc != nil {
		sortField = q.SortDesc.String()
		desc = true
	}

	switch sortField {
	case "date", "dateString", "mills", "startDate", "occurred_at", "start":
		sortField = "occurred_at"
	case "created_at", "srvCreated":
		sortField = "created_at"
	case "updated_at", "srvModified":
		sortField = "updated_at"
	case "name", "defaultProfile", "device":
		sortField = "name"
	default:
		slog.Warn("Unknown sort or sort$desc field, defaulting to occurred_at desc", "sort", sortField)
		return nil
	}

	fullString := sortField
	if desc {
		fullString = "-" + fullString
	}
	for _, allowed := range allowedSortFields {
		if string(allowed) == fullString {
			return &allowed
		}
	}

	slog.Warn("Unknown sort field, defaulting to occurred_at desc", "sort", fullString)
	return nil
}

var nonFilters = map[string]bool{
	"Limit":    true,
	"Skip":     true,
	"Sort":     true,
	"SortDesc": true,
	"Envelope": true,
	"Fields":   true,
	"Token":    true,
	"Raw":      true,
	"Other":    true,
}

func (q *QueryValues) HasAnyFilters() bool {
	if q == nil {
		return false
	}

	rv := reflect.ValueOf(q).Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// skip ignored fields
		if nonFilters[fieldType.Name] {
			continue
		}

		if isNonEmpty(field) {
			return true
		}
	}

	return false
}

func isNonEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		return !v.IsNil()
	default:
		// zero value check for everything else
		return !v.IsZero()
	}
}
