package sitecollections

import (
	"cmp"
	"math"
	"reflect"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/limits"
)

// tables are the tables of the collections not named for theirs
var tables = map[string]string{"food": "food_intakes", "glucose": "glucose_readings"}

// glucoseColumns are the columns in mg/dL of a model's glucose, which it has in the settings' units
var glucoseColumns = map[string]string{"glucose_readings.Reading": "glucose_readings.Mgdl",
	"temp_targets.LowTarget": "temp_targets.LowTargetMgdl", "temp_targets.HighTarget": "temp_targets.HighTargetMgdl"}

// checkLimits refuses a model with a value its column does not take, by limits.Of, before it is written: each field
// of the model of the same name as a column of the collection's table. A field not sent is not checked: a nil
// pointer or an empty text, and on an update any zero value.
func checkLimits(collection string, model any, siteContext *httputils.SiteContext, update bool) *apiutils.HTTPError {
	table, units := cmp.Or(tables[collection], collection), sitemodels.ContextUnits(siteContext)
	v := reflect.Indirect(reflect.ValueOf(model))
	for i := range v.NumField() {
		field, value := v.Type().Field(i), v.Field(i)
		column, glucose := glucoseColumns[table+"."+field.Name]
		limit, ok := limits.Of(cmp.Or(column, table+"."+field.Name))
		if value.Kind() == reflect.Pointer && !value.IsNil() {
			value = value.Elem()
		}
		if !ok || value.Kind() == reflect.Pointer || value.IsZero() && (update || value.Kind() == reflect.String) {
			continue
		}
		sent := value.Interface()
		if glucose {
			sent = units.ToMGDL(value.Float())
		}
		if !limit.Takes(sent) {
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			return apiutils.ErrConstraintCheckFailedWithHint(outOfLimit(name, limit, glucose, units))
		}
	}
	return nil
}

// outOfLimit says what a field takes: its range, in the settings' units for glucose, or its length
func outOfLimit(field string, limit sitedb.ColumnLimit, glucose bool, units *consts.Units) *i18n.Error {
	data := i18n.Data{"Field": field, "Max": limit.MaxLength}
	switch {
	case limit.MaxLength > 0 && limit.MinLength == 0:
		return i18n.NewError("errors.data.too_long", data)
	case limit.Min == nil || limit.Max == nil || limit.MaxExclusive || limit.MaxLength > 0:
		return i18n.NewError("errors.as_written", i18n.Data{"Text": field + " " + limit.Rule()})
	}
	low, high := *limit.Min, *limit.Max
	if glucose {
		// Rounded in to one decimal place, so each bound said is one taken
		low, high = math.Ceil(consts.Round(units.FromMGDL(low)*10, 6))/10, math.Floor(consts.Round(units.FromMGDL(high)*10, 6))/10
		data["Units"] = string(*units)
	}
	data["Min"], data["Max"] = strconv.FormatFloat(low, 'f', -1, 64), strconv.FormatFloat(high, 'f', -1, 64)
	if limit.MinExclusive {
		return i18n.NewError("errors.data.more_than", data)
	}
	return i18n.NewError("errors.data.out_of_range", data)
}
