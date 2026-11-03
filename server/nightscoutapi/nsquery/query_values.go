package nsquery

import (
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"nighttroop.t1d.nz/utils/mapstructure"
	"nighttroop.t1d.nz/utils/types"
)

type QueryValues struct {
	// Dates
	// lte and gte both
	DateGTE *types.DateInt `mapstructure:"dategte,omitempty" json:"dategte,omitempty"`
	DateGT  *types.DateInt `mapstructure:"dategt,omitempty" json:"dategt,omitempty"`

	DateEQ *types.DateInt `mapstructure:"dateeq,omitempty" json:"dateeq,omitempty"`

	DateLTE *types.DateInt `mapstructure:"datelte,omitempty" json:"datelte,omitempty"`
	DateLT  *types.DateInt `mapstructure:"datelt,omitempty" json:"datelt,omitempty"`

	DateIn *QueryStringArray `mapstructure:"datein,omitempty" json:"datein,omitempty"`

	CreatedAtGTE *types.DateInt `mapstructure:"createdatgte,omitempty" json:"createdatgte,omitempty"`
	CreatedAtGT  *types.DateInt `mapstructure:"createdatgt,omitempty" json:"createdatgt,omitempty"`

	CreatedAtLTE *types.DateInt `mapstructure:"createdatlte,omitempty" json:"createdatlte,omitempty"`
	CreatedAtLT  *types.DateInt `mapstructure:"createdatlt,omitempty" json:"createdatlt,omitempty"`

	UpdatedGTE *types.DateInt `mapstructure:"updatedgte,omitempty" json:"updatedgte,omitempty"`
	UpdatedGT  *types.DateInt `mapstructure:"updatedgt,omitempty" json:"updatedgt,omitempty"`
	UpdatedLTE *types.DateInt `mapstructure:"updatedlte,omitempty" json:"updatedlte,omitempty"`
	UpdatedLT  *types.DateInt `mapstructure:"updatedlt,omitempty" json:"updatedlt,omitempty"`

	// Type
	TypeEq  *types.ShortString `mapstructure:"typeeq,omitempty" json:"typeeq,omitempty"`
	TypeNeq *types.ShortString `mapstructure:"typeneq,omitempty" json:"typeneq,omitempty"`
	TypeIn  *QueryStringArray  `mapstructure:"typein,omitempty" json:"typein,omitempty"`
	TypeNin *QueryStringArray  `mapstructure:"typenin,omitempty" json:"typenin,omitempty"`

	// Identity
	IDEq *types.ShortString `mapstructure:"ideq,omitempty" json:"ideq,omitempty"`
	IDIn *QueryStringArray  `mapstructure:"idin,omitempty" json:"idin,omitempty"`

	DeviceEq *types.ShortString `mapstructure:"deviceeq,omitempty" json:"deviceeq,omitempty"`
	AppEq    *types.ShortString `mapstructure:"appeq,omitempty" json:"appeq,omitempty"`

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SgvEQ  *types.FlexibleInt64 `mapstructure:"sgveq,omitempty" json:"sgveq,omitempty"`
	SgvGTE *types.FlexibleInt64 `mapstructure:"sgvgte,omitempty" json:"sgvgte,omitempty"`
	SgvGT  *types.FlexibleInt64 `mapstructure:"sgvgt,omitempty" json:"sgvgt,omitempty"`
	SgvLTE *types.FlexibleInt64 `mapstructure:"sgvlte,omitempty" json:"sgvlte,omitempty"`
	SgvLT  *types.FlexibleInt64 `mapstructure:"sgvlt,omitempty" json:"sgvlt,omitempty"`

	// Treatments
	CarbsEQ  *types.FlexibleInt64 `mapstructure:"carbseq,omitempty" json:"carbseq,omitempty"`
	CarbsGTE *types.FlexibleInt64 `mapstructure:"carbsgte,omitempty" json:"carbsgte,omitempty"`
	CarbsGT  *types.FlexibleInt64 `mapstructure:"carbsgt,omitempty" json:"carbsgt,omitempty"`
	CarbsLTE *types.FlexibleInt64 `mapstructure:"carbslte,omitempty" json:"carbslte,omitempty"`
	CarbsLT  *types.FlexibleInt64 `mapstructure:"carbslt,omitempty" json:"carbslt,omitempty"`

	InsulinEQ  *types.FlexibleInt64 `mapstructure:"insulineq,omitempty" json:"insulineq,omitempty"`
	InsulinGTE *types.FlexibleInt64 `mapstructure:"insulingte,omitempty" json:"insulingte,omitempty"`
	InsulinGT  *types.FlexibleInt64 `mapstructure:"insulingt,omitempty" json:"insulingt,omitempty"`
	InsulinLTE *types.FlexibleInt64 `mapstructure:"insulinlte,omitempty" json:"insulinlte,omitempty"`
	InsulinLT  *types.FlexibleInt64 `mapstructure:"insulinlt,omitempty" json:"insulinlt,omitempty"`

	DurationEQ  *types.FlexibleInt64 `mapstructure:"durationeq,omitempty" json:"durationeq,omitempty"`
	DurationGTE *types.FlexibleInt64 `mapstructure:"durationgte,omitempty" json:"durationgte,omitempty"`
	DurationGT  *types.FlexibleInt64 `mapstructure:"durationgt,omitempty" json:"durationgt,omitempty"`
	DurationLTE *types.FlexibleInt64 `mapstructure:"durationlte,omitempty" json:"durationlte,omitempty"`
	DurationLT  *types.FlexibleInt64 `mapstructure:"durationlt,omitempty" json:"durationlt,omitempty"`

	NameEq *types.ShortString `mapstructure:"nameeq,omitempty" json:"nameeq,omitempty"`

	// Trio's carbs and temp targets fetches, and who entered a treatment, as Trio and iAPS leave out their own
	CarbsExists    *types.FlexibleBool `mapstructure:"carbsexists,omitempty" json:"carbsexists,omitempty"`
	DurationExists *types.FlexibleBool `mapstructure:"durationexists,omitempty" json:"durationexists,omitempty"`
	EnteredByEq    *types.ShortString  `mapstructure:"enteredbyeq,omitempty" json:"enteredbyeq,omitempty"`
	EnteredByNin   *QueryStringArray   `mapstructure:"enteredbynin,omitempty" json:"enteredbynin,omitempty"`

	// Altering Return values or structure
	// Limits
	Limit *types.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip  *types.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`

	// Sort
	Sort     *types.ShortString `mapstructure:"sort,omitempty" json:"sort,omitempty"`
	SortDesc *types.ShortString `mapstructure:"sort$desc,omitempty" json:"sort$desc,omitempty"`

	Envelope *types.FlexibleBool `mapstructure:"envelope,omitempty" json:"envelope,omitempty"`
	Fields   *QueryStringArray   `mapstructure:"fields,omitempty" json:"fields,omitempty"`

	IncludeDeleted *types.FlexibleBool `mapstructure:"includeDeleted,omitempty" json:"includeDeleted,omitempty"`

	// This is Nighttroops value to show RawJsonData
	Raw *types.FlexibleBool `mapstructure:"raw,omitempty" json:"raw,omitempty"`

	Token *string `mapstructure:"token,omitempty" json:"-"`

	// UnknownFilter is a find[...] filter on a field Nighttroop cannot filter on, so the search matches nothing
	UnknownFilter bool `mapstructure:"-" json:"-"`

	Other map[string]interface{} `mapstructure:",remain"`
}

// This is just much nicer to work with but the reverse is faster, so we
var forwardLookupName = map[string][]string{
	// A Nightscout created_at is when it happened, as it is written out, and srvCreated is when the server got it
	"dategte": {"date$gte", "find[date][$gte]", "find[dateString][$gte]", "mills$gte", "find[mills][$gte]", "startDate$gte", "find[startDate][$gte]", "start$gte", "find[start][$gte]", "created_at$gte", "find[created_at][$gte]"},
	"dategt":  {"date$gt", "find[date][$gt]", "find[dateString][$gt]", "mills$gt", "find[mills][$gt]", "startDate$gt", "find[startDate][$gt]", "start$gt", "find[start][$gt]", "created_at$gt", "find[created_at][$gt]"},

	"dateeq": {"date$eq", "find[date][$eq]", "find[dateString][$eq]", "mills$eq", "find[mills][$eq]", "startDate$eq", "find[startDate][$eq]", "start$eq", "find[start][$eq]"},

	"datelte": {"date$lte", "find[date][$lte]", "find[dateString][$lte]", "mills$lte", "find[mills][$lte]", "startDate$lte", "find[startDate][$lte]", "start$lte", "find[start][$lte]", "created_at$lte", "find[created_at][$lte]"},
	"datelt":  {"date$lt", "find[date][$lt]", "find[dateString][$lt]", "mills$lt", "find[mills][$lt]", "startDate$lt", "find[startDate][$lt]", "start$lt", "find[start][$lt]", "created_at$lt", "find[created_at][$lt]"},

	"datein": {"date$in", "find[date][$in]"},

	"createdatgte": {"srvCreated$gte", "find[srvCreated][$gte]"},
	"createdatgt":  {"srvCreated$gt", "find[srvCreated][$gt]"},

	"createdatlte": {"srvCreated$lte", "find[srvCreated][$lte]"},
	"createdatlt":  {"srvCreated$lt", "find[srvCreated][$lt]"},

	"updatedgte": {"updated_at$gte", "find[updated_at][$gte]", "srvModified$gte", "find[srvModified][$gte]"},
	"updatedgt":  {"updated_at$gt", "find[updated_at][$gt]", "srvModified$gt", "find[srvModified][$gt]"},

	"updatedlte": {"updated_at$lte", "find[updated_at][$lte]", "srvModified$lte", "find[srvModified][$lte]"},
	"updatedlt":  {"updated_at$lt", "find[updated_at][$lt]", "srvModified$lt", "find[srvModified][$lt]"},

	"durationeq":  {"duration", "find[duration]", "find[duration][$eq]", "durationMinutes", "find[durationMinutes]", "find[durationMinutes][$eq]"},
	"durationgte": {"duration$gte", "find[duration][$gte]", "durationMinutes$gte", "find[durationMinutes][$gte]"},
	"durationgt":  {"duration$gt", "find[duration][$gt]", "durationMinutes$gt", "find[durationMinutes][$gt]"},
	"durationlte": {"duration$lte", "find[duration][$lte]", "durationMinutes$lte", "find[durationMinutes][$lte]"},
	"durationlt":  {"duration$lt", "find[duration][$lt]", "durationMinutes$lt", "find[durationMinutes][$lt]"},

	// API v3 also writes it field$eq, as Nightguard asks eventType$eq
	"typeeq": {"type", "type$eq", "find[type]", "find[type][$eq]",
		"eventType", "eventType$eq", "find[eventType]", "find[eventType][$eq]",
		"activityType", "activityType$eq", "find[activityType]", "find[activityType][$eq]",
		"category", "category$eq", "find[category]", "find[category][$eq]",
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

	"carbsexists":    {"carbs$exists", "find[carbs][$exists]"},
	"durationexists": {"duration$exists", "find[duration][$exists]"},
	"enteredbyeq":    {"enteredBy", "enteredBy$eq", "find[enteredBy]", "find[enteredBy][$eq]"},
	"enteredbynin":   {"enteredBy$nin", "find[enteredBy][$nin]"},

	"limit": {"limit", "count"},
}

// enteredByNe is a $ne on enteredBy, which Trio sends as find[$and][n][enteredBy][$ne] and iAPS repeated
var enteredByNe = regexp.MustCompile(`^find(\[\$and\]\[\d+\])?\[enteredBy\]\[\$ne\]$`)

// Built once at startup; filling it lazily raced when the first requests arrived together
var reverseLookupName = func() map[string]string {
	reverse := map[string]string{}
	for k, v := range forwardLookupName {
		for _, vv := range v {
			reverse[vv] = k
		}
	}
	return reverse
}()

func ParseQueryValues(values url.Values) (QueryValues, error) {
	qvMap := map[string]string{}
	notEnteredBy := []string{}
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		// A key ending [] is a list, as xDrip4iOS sends find[date][$in][]
		key = strings.TrimSuffix(key, "[]")

		// Each $ne on enteredBy is one of a $nin; iAPS encodes them twice, so iAPS%20Watch is iAPS Watch
		if enteredByNe.MatchString(key) {
			for _, val := range vals {
				if unescaped, err := url.PathUnescape(val); err == nil {
					val = unescaped
				}
				notEnteredBy = append(notEnteredBy, val)
			}
			continue
		}

		// Speciial Handling of fields that can be repeated in the query string, like fields[] and find[identifier][$in]
		// Comma Seperated Values can just be patched on

		if key == "fields" {
			qvMap[key] = strings.Join(vals, ",")
			continue
		}

		// A uuid is a treatment's identifier when it was sent without an _id. xDrip+ sends its uuid without dashes,
		// cut to 24 characters, as the _id, and finds the treatment by its uuid to delete it.
		if key == "find[uuid]" {
			id := strings.ReplaceAll(vals[0], "-", "")
			qvMap["find[_id][$in]"] = vals[0] + "|" + id[:min(len(id), 24)]
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
	if len(notEnteredBy) > 0 {
		qvMap["find[enteredBy][$nin]"] = strings.Join(append(notEnteredBy, qvMap["find[enteredBy][$nin]"]), "|")
	}

	var qv QueryValues
	err := mapstructure.Decode(qvMap, &qv, func(mapKey, fieldName string) bool {
		if v, ok := reverseLookupName[mapKey]; ok {
			mapKey = v
		}
		return strings.EqualFold(mapKey, fieldName)
	})
	if err != nil {
		return QueryValues{}, fmt.Errorf("failed to decode query values: %w", err)
	}

	if qv.Sort != nil && qv.SortDesc != nil {
		return QueryValues{}, fmt.Errorf("cannot specify both sort and sort$desc parameters")
	}

	if len(qv.Other) > 0 {
		slog.Warn("Unrecognized Nightscout Query Parameters", "params", qv.Other)
	}
	for key := range qv.Other {
		qv.UnknownFilter = qv.UnknownFilter || strings.HasPrefix(key, "find[")
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
	case "date", "dateString", "mills", "startDate", "occurred_at", "start", "created_at":
		sortField = "occurred_at"
	case "srvCreated":
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
