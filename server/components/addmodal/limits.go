package addmodal

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/limits"
)

// limit is what the field takes: its column's limit, narrowed by the column's override
func (f Field) limit() sitedb.ColumnLimit {
	limit, _ := limits.Of(f.Column)
	return limit
}

// Check is an error for each number or text field that names no column
func Check(fields []Field) error {
	var errs []error
	for _, f := range fields {
		if _, ok := limits.Column(f.Column); !ok && f.Kind != Select && f.Kind != When {
			errs = append(errs, fmt.Errorf("%s names no column: %q", f.Name, f.Column))
		}
	}
	return errors.Join(errs...)
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
