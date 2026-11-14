package addmodal

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

// columnLimit is the limit of a column, by its table and its field in the table's rows, in the site's DB or the
// control's
func columnLimit(column string) (sitedb.ColumnLimit, bool) {
	table, field, _ := strings.Cut(column, ".")
	if limit, ok := sitedb.ColumnLimitOf(table, field); ok {
		return limit, true
	}
	return controldb.ColumnLimitOf(table, field)
}

// limit is what the field takes: its column's limit, with its own Min and Max in place of the column's
func (f Field) limit() sitedb.ColumnLimit {
	limit, _ := columnLimit(f.Column)
	if n, err := strconv.Atoi(f.Max); err == nil && f.Kind == Text {
		limit.MaxLength = n
	}
	if v, err := strconv.ParseFloat(f.Min, 64); err == nil && f.Kind != Text {
		limit.Min, limit.MinExclusive = &v, false
	}
	if v, err := strconv.ParseFloat(f.Max, 64); err == nil && f.Kind != Text {
		limit.Max, limit.MaxExclusive = &v, false
	}
	return limit
}

// Check is an error for each number or text field that names no column, or whose Min or Max takes what its column
// does not, as a field's own limits only narrow its column's
func Check(fields []Field) error {
	var errs []error
	for _, f := range fields {
		column, ok := columnLimit(f.Column)
		n, err := strconv.Atoi(f.Max)
		switch {
		case f.Kind == Select || f.Kind == When:
		case !ok:
			errs = append(errs, fmt.Errorf("%s names no column: %q", f.Name, f.Column))
		case f.Kind == Text && f.Max != "" && (err != nil || n < 1 || column.MaxLength > 0 && n > column.MaxLength):
			errs = append(errs, fmt.Errorf("%s's length %q is more than its column %s takes", f.Name, f.Max, f.Column))
		case f.Kind != Text && (!takes(column, f.Min) || !takes(column, f.Max)):
			errs = append(errs, fmt.Errorf("%s's min %q or max %q is outside its column %s", f.Name, f.Min, f.Max, f.Column))
		}
	}
	return errors.Join(errs...)
}

// takes is whether a field's own limit, a number, is one its column takes, or it has none
func takes(column sitedb.ColumnLimit, limit string) bool {
	v, err := strconv.ParseFloat(limit, 64)
	return limit == "" || err == nil && column.Takes(v)
}

// numberAttributes are a number's step, and its limit as it is typed in: a Percent's fraction as a percent, and
// glucose in its Units. Each bound is rounded to the step's decimal places, one for any, and moved in by one place
// when its limit does not take it, so the input never takes what the column does not.
func numberAttributes(f Field) templ.Attributes {
	attributes, step, limit := templ.Attributes{}, f.Step, f.limit()
	from, to := func(v float64) float64 { return v }, func(v float64) float64 { return v }
	if f.Kind == Percent {
		attributes["data-percent"], step = true, "any"
		from, to = func(v float64) float64 { return v * 100 }, func(v float64) float64 { return v / 100 }
	} else if f.Units != nil {
		from, to = f.Units.FromMGDL, f.Units.ToMGDL
	}
	attributes["step"] = step
	places := 0
	if _, decimals, ok := strings.Cut(step, "."); ok {
		places = len(decimals)
	} else if step == "any" {
		places = 1
	}
	if limit.Min != nil {
		attributes["min"] = bound(*limit.Min, limit.MinExclusive, 1, from, to, places)
	}
	if limit.Max != nil {
		attributes["max"] = bound(*limit.Max, limit.MaxExclusive, -1, from, to, places)
	}
	return attributes
}

// bound is a limit as it is typed in, from, rounded to the places, then moved by one place inward, 1 for a min and -1
// for a max, when the limit does not take it back, to
func bound(limit float64, exclusive bool, inward float64, from, to func(float64) float64, places int) string {
	v := consts.Round(from(limit), places)
	if out := inward * (to(v) - limit); out < 0 || (out == 0 && exclusive) {
		v = consts.Round(v+inward*math.Pow10(-places), places)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
