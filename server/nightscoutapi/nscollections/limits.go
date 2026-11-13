package nscollections

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

// A document is checked against the limits of the columns it is stored in, from the site DB's schema, before it is
// stored. A value its column does not take, a number out of its range, a text too long, or a value not of its field's
// type, is of the field sent that it is from, found by leaving out each field sent in turn:
//   - one the document cannot be stored without, as its column is NOT NULL, refuses the document, saying the field,
//     what it takes and what was sent
//   - any other is left out and the rest stored, and the request records one client error saying each left out and why
//
// The few columns of a collection that are not as their NULL says are in columnRequired.

// columnRequired are the columns, by collection, table and column, whose value out of its limits refuses the document,
// true, or is left out, false, though the column says otherwise:
//   - a treatment's glucose is left out, as AAPS sends a Bolus Wizard that used no BG with glucose 0, and does not send
//     a refused one again
//   - a treatment's carbs refuse it, as xDrip+ is told so rather than its treatment kept without them
//   - a temp basal's percent refuses it, as one left out would keep a temp basal of 0 U/hr
//   - a profile switch's shift of more than a day refuses it, as one left out would keep it unshifted
//   - an activity's heart rate refuses it, as xDrip+ is told so rather than its heart rate kept without one
var columnRequired = map[string]bool{
	"treatments.glucose_readings.mgdl": false, "treatments.food_intakes.carbs": true,
	"treatments.temp_basals.multiplier": true, "treatments.profile_switches.timeshift": true,
	"treatments.profile_switches.original_timeshift_ms": true, "activity.activities.heart_rate": true,
}

const (
	// maxWrongTypes is the most fields of a document left out for their type before it is refused
	maxWrongTypes = 16
	// maxFieldsTried is the most fields sent that are left out in turn to find which a value out of its limits is from
	maxFieldsTried = 256
)

// leftOut is what was left out of a request's document, and why
type leftOut []string

// record has the request record what was left out as a client error, when the document was taken
func (l leftOut) record(ctx context.Context, err *nsutils.HTTPError) {
	if err == nil && len(l) > 0 {
		httputils.SetLeftOut(ctx, "Left out "+strings.Join(l, "; "))
	}
}

// rows are the rows the model data is would be stored as, with their values by row and column, false when data is not
// one
func (c Collection[NS, DM]) rows(data map[string]any, patched *DM) (map[string]sitedb.ColumnValue, bool) {
	var nsModel NS
	if utils.ConvertViaJSON(data, &nsModel) != nil {
		return nil, false
	}
	return columnValues(c.db.Rows(nsModel, patched)), true
}

// columnValues are the rows' values, by their table, which of its rows, and their column, e.g. glucose_readings#0.Mgdl
func columnValues(rows []sitedb.TableRow) map[string]sitedb.ColumnValue {
	values, seen := map[string]sitedb.ColumnValue{}, map[string]int{}
	for _, row := range rows {
		for _, value := range sitedb.ColumnValues(row) {
			values[fmt.Sprintf("%s#%d.%s", row.Table, seen[row.Table], value.Field)] = value
		}
		seen[row.Table]++
	}
	return values
}

// required is whether a value of a row, by its key in columnValues, refuses the document when out of its limits
func (c Collection[NS, DM]) required(key string, limit sitedb.ColumnLimit) bool {
	table, _, _ := strings.Cut(key, "#")
	if required, ok := columnRequired[c.db.Name()+"."+table+"."+limit.Column]; ok {
		return required
	}
	return limit.Required
}

// decode is the model data is, each field of a type its model cannot read, as a battery sent as text, left out of
// data and said in left, unless the model cannot be stored without it, which refuses it
func (c Collection[NS, DM]) decode(data map[string]any, left *leftOut) (NS, *nsutils.HTTPError) {
	var nsModel NS
	for range maxWrongTypes {
		nsModel = *new(NS)
		err := utils.ConvertViaJSON(data, &nsModel)
		if err == nil {
			return nsModel, nil
		}
		var typeErr *json.UnmarshalTypeError
		path := c.wrongTypeField(data, err)
		if path == nil || !errors.As(err, &typeErr) {
			return nsModel, nsutils.BadDataBodyErrorWithHint(err.Error())
		}
		reason := fmt.Sprintf("%s must be %s (was %s)", strings.Join(path, "."), typeName(typeErr.Type), shown(get(data, path)))
		if c.requiredOfType(data, path, typeErr.Type) {
			return nsModel, nsutils.BadDataBodyErrorWithHint(reason)
		}
		remove(data, path)
		*left = append(*left, reason)
	}
	return nsModel, nsutils.BadDataBodyErrorWithHint("too many fields of the wrong type")
}

// wrongTypeField is the field sent in data that err says is not of its type, nil when none is. A type's own decoding,
// as a number's that takes text too, says only the value, so it is the field sent with that value whose leaving out
// leaves the error.
func (c Collection[NS, DM]) wrongTypeField(data map[string]any, err error) []string {
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		return nil
	}
	if typeErr.Field != "" {
		if path := strings.Split(typeErr.Field, "."); has(data, path) {
			return path
		}
		return nil
	}
	for _, field := range sentFields(data, nil) {
		sent := get(data, field)
		if raw, _ := json.Marshal(sent); string(raw[:min(len(raw), len(typeErr.Value))]) != typeErr.Value {
			continue
		}
		remove(data, field)
		var nsModel NS
		without := utils.ConvertViaJSON(data, &nsModel)
		set(data, field, sent)
		if without == nil || without.Error() != err.Error() {
			return field
		}
	}
	return nil
}

// requiredOfType is whether a field of the wrong type is one its model cannot be stored without: whether a value of
// its type sent in its place is kept in a column the model cannot be stored without
func (c Collection[NS, DM]) requiredOfType(data map[string]any, path []string, of reflect.Type) bool {
	sent := get(data, path)
	defer set(data, path, sent)
	var instead any = 1
	switch of.Kind() {
	case reflect.Bool:
		instead = true
	case reflect.String:
		instead = "x"
	}
	set(data, path, instead)
	with, ok := c.rows(data, nil)
	unstable := c.unstable(data, with, nil)
	remove(data, path)
	without, ok2 := c.rows(data, nil)
	if !ok || !ok2 {
		return false
	}
	for key, value := range with {
		if other, kept := without[key]; !unstable[key] && (!kept || other.Value != value.Value) {
			if c.required(key, value.Limit) {
				return true
			}
		}
	}
	return false
}

// withinLimits is the model data is, with each field sent out of its column's limits left out of data and said in left,
// or the refusal of one it cannot be stored without. A patch, of patched, is refused only for what it sends.
func (c Collection[NS, DM]) withinLimits(data map[string]any, nsModel NS, patched *DM, left *leftOut) (NS, *nsutils.HTTPError) {
	// The rows are of a model of their own, as making them can change the model, as a profile's targets to mg/dL
	values, _ := c.rows(data, patched)
	var keys []string
	for key, value := range values {
		if !value.Taken() {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nsModel, nil
	}
	slices.Sort(keys)
	fields := c.fieldsOf(data, values, keys, patched)
	for _, key := range keys {
		value, field := values[key], fields[key]
		if field == nil && patched != nil {
			continue
		}
		name, sent := value.Limit.Column, value.Value
		if field != nil {
			name, sent = strings.Join(field, "."), get(data, field)
		}
		if field != nil && !has(data, field) {
			continue
		}
		reason := fmt.Sprintf("%s %s (was %s)", name, value.Limit.Rule(), shown(sent))
		// One kept otherwise than sent, as glucose in mmol/L kept in mg/dL, says as what
		if stored := shown(value.Value); isNumber(sent) && shown(sent) != stored {
			reason = fmt.Sprintf("%s %s (was %s, stored as %s)", name, value.Limit.Rule(), shown(sent), stored)
		}
		if field == nil || c.required(key, value.Limit) {
			return nsModel, &nsutils.HTTPError{Status: http.StatusBadRequest, Message: nsutils.BadDataBodyErrorWithHint(reason).Message,
				Column: value.Limit.Column}
		}
		remove(data, field)
		*left = append(*left, reason)
	}
	var kept NS
	if err := utils.ConvertViaJSON(data, &kept); err != nil {
		return nsModel, nsutils.BadDataBodyErrorWithHint(err.Error())
	}
	return kept, nil
}

// fieldsOf are the fields sent each value out of its limits is from, nil for one that is from none, found by leaving
// out each field sent in turn. Of those that change the value, the field is the one that best tells it: see better.
func (c Collection[NS, DM]) fieldsOf(data map[string]any, values map[string]sitedb.ColumnValue, keys []string, patched *DM) map[string][]string {
	fields, best := map[string][]string{}, map[string]fieldScore{}
	unstable := c.unstable(data, values, patched)
	for _, field := range sentFields(data, nil) {
		sent := get(data, field)
		remove(data, field)
		without, ok := c.rows(data, patched)
		set(data, field, sent)
		if !ok {
			continue
		}
		changes := 0
		for key, value := range values {
			if other, kept := without[key]; !unstable[key] && (!kept || other.Value != value.Value) {
				changes++
			}
		}
		for _, key := range keys {
			value, kept := without[key]
			score := fieldScore{depth: len(field), changes: changes}
			score.gone = !kept || value.Value == nil || (reflect.ValueOf(value.Value).IsZero() && !reflect.ValueOf(values[key].Value).IsZero())
			switch {
			case score.gone || value.Taken():
				score.class = 2
			case value.Value != values[key].Value:
				score.class = 1
			default:
				continue
			}
			if score.better(best[key]) {
				fields[key], best[key] = field, score
			}
		}
	}
	return fields
}

// fieldScore is how well leaving out a field sent tells it is the one a value out of its limits is from
type fieldScore struct {
	// class is 2 when it leaves the value out or makes it one its column takes, and 1 when it only changes it
	class int
	// depth is how deep in the document the field is, and gone whether it leaves the value out
	depth int
	gone  bool
	// changes is how many values of the rows it changes
	changes int
}

// better is whether leaving out a field tells it better than another: it leaves out or puts right the value rather than
// only change it; it is deeper, so a field rather than the object it is in; it changes fewer values, so a treatment's
// insulin rather than its eventType, and an entry's sgv rather than its units; and it leaves the value out
func (s fieldScore) better(other fieldScore) bool {
	return cmp.Or(cmp.Compare(s.class, other.class), cmp.Compare(s.depth, other.depth), cmp.Compare(other.changes, s.changes),
		cmp.Compare(btoi(s.gone), btoi(other.gone))) > 0
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// unstable are the values of the rows that are not the same each time data is converted, as a new row's identifier,
// which no field sent is the one they are from
func (c Collection[NS, DM]) unstable(data map[string]any, values map[string]sitedb.ColumnValue, patched *DM) map[string]bool {
	unstable := map[string]bool{}
	again, _ := c.rows(data, patched)
	for key, value := range values {
		if other, kept := again[key]; !kept || other.Value != value.Value {
			unstable[key] = true
		}
	}
	return unstable
}

// sentFields are the paths of the fields in data, each object's before the object, by name, at most maxFieldsTried
func sentFields(data map[string]any, prefix []string) [][]string {
	var fields [][]string
	for _, name := range slices.Sorted(func(yield func(string) bool) {
		for name := range data {
			if !yield(name) {
				return
			}
		}
	}) {
		field := append(slices.Clone(prefix), name)
		if object, ok := data[name].(map[string]any); ok {
			fields = append(fields, sentFields(object, field)...)
		}
		fields = append(fields, field)
	}
	return fields[:min(len(fields), maxFieldsTried)]
}

// object is the object in data that holds the field at path, nil when there is none
func object(data map[string]any, path []string) map[string]any {
	for _, name := range path[:len(path)-1] {
		data, _ = data[name].(map[string]any)
	}
	return data
}

func has(data map[string]any, path []string) bool {
	_, ok := object(data, path)[path[len(path)-1]]
	return ok
}

func get(data map[string]any, path []string) any {
	return object(data, path)[path[len(path)-1]]
}

func set(data map[string]any, path []string, value any) {
	if o := object(data, path); o != nil {
		o[path[len(path)-1]] = value
	}
}

func remove(data map[string]any, path []string) {
	delete(object(data, path), path[len(path)-1])
}

// typeName is what a value of the type is to a client
func typeName(of reflect.Type) string {
	switch of.Kind() {
	case reflect.Bool:
		return "true or false"
	case reflect.String:
		return "text"
	}
	return "a number"
}

// isNumber is whether a value sent is a number
func isNumber(value any) bool {
	_, ok := value.(float64)
	return ok
}

// shown is a value sent as a client error says it: its JSON, or for a long text how long it is
func shown(value any) string {
	if text, ok := value.(string); ok && utf8.RuneCountInString(text) > 32 {
		return fmt.Sprintf("%d characters", utf8.RuneCountInString(text))
	}
	raw, _ := json.Marshal(value)
	if len(raw) > 64 {
		return fmt.Sprintf("%d characters of JSON", len(raw))
	}
	return string(raw)
}
