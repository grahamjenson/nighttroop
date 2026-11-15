package testutils

import (
	"cmp"
	"context"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// LimitRecipe writes a valid row of a table, its i-th, with one field changed, so each of its columns' limits can be
// checked through the database
type LimitRecipe struct {
	row   reflect.Type
	write func(ctx context.Context, i int, set func(row reflect.Value)) error
	// as is the row's field a limit's field is written by, when it is named another way
	as map[string]string
	// notWritten are the fields with a limit the write does not take, such as one only an update sets
	notWritten []string
}

// Recipe writes the i-th row of a table: row is a valid one, its unique columns made from i, written by write, a
// prepared query. notWritten are the fields with a limit it does not take, so are not checked.
func Recipe[P any](row func(i int) P, write func(context.Context, P) error, notWritten ...string) LimitRecipe {
	return LimitRecipe{row: reflect.TypeFor[P](), notWritten: notWritten,
		write: func(ctx context.Context, i int, set func(reflect.Value)) error {
			p := row(i)
			set(reflect.ValueOf(&p).Elem())
			return write(ctx, p)
		}}
}

// As writes a limit's field by the row's field param, such as loop settings' first and last seen by when it was seen
func (r LimitRecipe) As(field, param string) LimitRecipe {
	r.as = maps.Clone(r.as)
	if r.as == nil {
		r.as = map[string]string{}
	}
	r.as[field] = param
	return r
}

// NoResult is a query that writes a row and gives it back, as one that only writes
func NoResult[P, R any](query func(context.Context, P) (R, error)) func(context.Context, P) error {
	return func(ctx context.Context, p P) error {
		_, err := query(ctx, p)
		return err
	}
}

// limitCase is a value at a limit's bound, or just past it, and whether the database takes it: a number, or a text of
// a length when length is not -1
type limitCase struct {
	name   string
	number float64
	length int
	taken  bool
}

// limitCases are the values at each of a limit's bounds and just past them, by 1 for an integer field and 0.001 for
// any other number, and just inside a bound it may not be
func limitCases(l sitedb.ColumnLimit, field reflect.Type) []limitCase {
	step := 0.001
	if field.Kind() == reflect.Pointer {
		field = field.Elem()
	}
	if field.Kind() == reflect.Int64 || field.Kind() == reflect.Int {
		step = 1
	}
	var cases []limitCase
	if l.Min != nil {
		in, out := *l.Min, *l.Min-step
		if l.MinExclusive {
			in, out = *l.Min+step, *l.Min
		}
		cases = append(cases, limitCase{"least", in, -1, true}, limitCase{"below the least", out, -1, false})
	}
	if l.Max != nil {
		in, out := *l.Max, *l.Max+step
		if l.MaxExclusive {
			in, out = *l.Max-step, *l.Max
		}
		cases = append(cases, limitCase{"most", in, -1, true}, limitCase{"above the most", out, -1, false})
	}
	if l.MinLength > 0 {
		cases = append(cases, limitCase{"shortest", 0, l.MinLength, true}, limitCase{"too short", 0, l.MinLength - 1, false})
	}
	if l.MaxLength > 0 {
		cases = append(cases, limitCase{"longest", 0, l.MaxLength, true}, limitCase{"too long", 0, l.MaxLength + 1, false})
	}
	return cases
}

// text is a text of n characters unique to i, a JSON string for a column of JSON, which is checked for its length only
func text(n, i int, json bool) string {
	if json {
		return `"` + text(max(n-2, 0), i, false) + `"`
	}
	s := strconv.Itoa(i) + "-" + strings.Repeat("x", n)
	return s[len(s)-n:]
}

// setLimitCase sets a row's field to the case's value
func setLimitCase(field reflect.Value, c limitCase, i int, json bool) {
	if field.Kind() == reflect.Pointer {
		field.Set(reflect.New(field.Type().Elem()))
		field = field.Elem()
	}
	switch {
	case c.length >= 0:
		field.SetString(text(c.length, i, json))
	case field.CanInt():
		field.SetInt(int64(c.number))
	default:
		field.SetFloat(c.number)
	}
}

// CheckColumnLimits writes, through each table's recipe, each column's bounds and just past them: its least and most,
// or just inside a bound it may not be, and its fewest and most characters. The database must take the first and
// refuse the second, so the limits values are checked by before they are written are the CHECKs the database keeps. A
// table with limits and no recipe, or a column with a limit its recipe does not write and does not say so, fails, so a
// new one is checked too.
func CheckColumnLimits(t *testing.T, limits map[string]map[string]sitedb.ColumnLimit, recipes map[string]LimitRecipe) {
	for table := range recipes {
		_, ok := limits[table]
		require.True(t, ok, "%s has a recipe but no limits", table)
	}
	i, columns, bounds := 0, 0, 0
	for _, table := range slices.Sorted(maps.Keys(limits)) {
		t.Run(table, func(t *testing.T) {
			recipe, hasRecipe := recipes[table]
			for _, field := range slices.Sorted(maps.Keys(limits[table])) {
				limit := limits[table][field]
				if limit.Min == nil && limit.Max == nil && limit.MinLength == 0 && limit.MaxLength == 0 {
					continue
				}
				require.True(t, hasRecipe, "%s has limits but no recipe to write it", table)
				param, written := recipe.row.FieldByName(cmp.Or(recipe.as[field], field))
				if slices.Contains(recipe.notWritten, field) {
					require.False(t, written, "%s.%s is written, so is checked", table, field)
					continue
				}
				require.True(t, written, "%s.%s has a limit its recipe does not write", table, field)
				json := strings.Contains(limit.Column, "json") || strings.HasSuffix(limit.Column, "errors") ||
					limit.Column == "attributes"
				cases := limitCases(limit, param.Type)
				columns, bounds = columns+1, bounds+len(cases)/2
				for _, c := range cases {
					i++
					err := recipe.write(t.Context(), i, func(row reflect.Value) {
						setLimitCase(row.FieldByIndex(param.Index), c, i, json)
					})
					value := any(c.number)
					if c.length >= 0 {
						value = strconv.Itoa(c.length) + " characters"
					}
					msg := []any{"%s.%s at its %s, %v (%s)", table, field, c.name, value, limit.Rule()}
					if c.taken {
						require.NoError(t, err, msg...)
					} else {
						require.ErrorContains(t, err, "CHECK constraint failed", msg...)
					}
				}
			}
		})
	}
	t.Logf("checked %d bounds of %d columns", bounds, columns)
}
