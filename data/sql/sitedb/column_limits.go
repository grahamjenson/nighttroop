package sitedb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ColumnLimit is what a column takes, as its NOT NULL and CHECK say, generated from the schema in z_limits.go
type ColumnLimit struct {
	// Column is the column's name
	Column string
	// Required is a column NOT NULL, or a primary key, which a row cannot be kept without
	Required bool
	// Min and Max are the least and most a number may be, and MinExclusive and MaxExclusive whether that itself is not
	Min, Max                   *float64
	MinExclusive, MaxExclusive bool
	// MinLength and MaxLength are the fewest and most characters a text may have, 0 for no limit
	MinLength, MaxLength int
}

// ColumnLimitOf is the limit of a table's column, by the name of its field in the table's rows, e.g. Mgdl
func ColumnLimitOf(table, field string) (ColumnLimit, bool) {
	limit, ok := columnLimits[table][field]
	return limit, ok
}

// Takes is whether the column takes the value, a number or a text: one of another kind, as a bool, has no limit here
func (l ColumnLimit) Takes(value any) bool {
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String:
		n := utf8.RuneCountInString(v.String())
		return n >= l.MinLength && (l.MaxLength == 0 || n <= l.MaxLength)
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := v.Convert(reflect.TypeFor[float64]()).Float()
		// Written so NaN is out of every range
		low := l.Min == nil || (!l.MinExclusive && n >= *l.Min) || (l.MinExclusive && n > *l.Min)
		high := l.Max == nil || (!l.MaxExclusive && n <= *l.Max) || (l.MaxExclusive && n < *l.Max)
		return low && high
	}
	return true
}

// Rule says what the column takes, e.g. "must be between 0 and 100" or "must be at most 64 characters", with mg/dL
// for a column of glucose kept in it
func (l ColumnLimit) Rule() string {
	if l.MaxLength > 0 || l.MinLength > 0 {
		switch {
		case l.MinLength == l.MaxLength:
			return fmt.Sprintf("must be %d characters", l.MaxLength)
		case l.MaxLength == 0:
			return fmt.Sprintf("must be at least %d characters", l.MinLength)
		case l.MinLength > 0:
			return fmt.Sprintf("must be %d to %d characters", l.MinLength, l.MaxLength)
		}
		return fmt.Sprintf("must be at most %d characters", l.MaxLength)
	}
	unit := ""
	if strings.HasSuffix(l.Column, "mgdl") {
		unit = " mg/dL"
	}
	low, high := "at least", "at most"
	if l.MinExclusive {
		low = "more than"
	}
	if l.MaxExclusive {
		high = "less than"
	}
	switch {
	case l.Min != nil && l.Max != nil && !l.MinExclusive && !l.MaxExclusive:
		return fmt.Sprintf("must be between %s and %s%s", number(*l.Min), number(*l.Max), unit)
	case l.Min != nil && l.Max != nil:
		return fmt.Sprintf("must be %s %s and %s %s%s", low, number(*l.Min), high, number(*l.Max), unit)
	case l.Min != nil:
		return fmt.Sprintf("must be %s %s%s", low, number(*l.Min), unit)
	case l.Max != nil:
		return fmt.Sprintf("must be %s %s%s", high, number(*l.Max), unit)
	}
	return "must be kept"
}

// number is a limit as it is written, never in an exponent
func number(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// TableRow is a row of a table as it is written, such as a CreateGlucoseReadingParams for glucose_readings
type TableRow struct {
	Table string
	Row   any
}

// ColumnValue is a row's value of a column with a limit, nil when it is a nil pointer, as a value not set, or not
// changed by an update, which is not checked
type ColumnValue struct {
	// Field is the value's field in the row, as ColumnLimitOf names it
	Field string
	Value any
	Limit ColumnLimit
}

// Taken is whether the column takes the value
func (c ColumnValue) Taken() bool {
	return c.Value == nil || c.Limit.Takes(c.Value)
}

// ColumnValues are the values of a row, a struct of a table's columns such as its Create params, of each column with a
// limit
func ColumnValues(row TableRow) []ColumnValue {
	var values []ColumnValue
	v := reflect.Indirect(reflect.ValueOf(row.Row))
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := range v.NumField() {
		field, value := v.Type().Field(i).Name, v.Field(i)
		limit, ok := ColumnLimitOf(row.Table, field)
		if !ok {
			continue
		}
		var kept any
		if value.Kind() != reflect.Pointer {
			kept = value.Interface()
		} else if !value.IsNil() {
			kept = value.Elem().Interface()
		}
		values = append(values, ColumnValue{Field: field, Value: kept, Limit: limit})
	}
	return values
}
