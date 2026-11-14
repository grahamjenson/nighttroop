// Package limits is what the site's forms and its JSON API take of each column: the column's limit from its schema,
// narrowed by an override for what people enter, so each limit is written in one place
package limits

import (
	"strings"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// Overrides narrow a column's limit for what people enter on the site, by table and field in the table's rows, in the
// column's units: a percent's fraction, glucose in mg/dL, and MaxLength a text's length. Nightscout clients and syncs
// are only held to the column's.
var Overrides = map[string]sitedb.ColumnLimit{
	// A dose far above any a pump or pen gives; as much again below 0 takes back one entered by mistake
	"insulin_doses.Dose": {Min: new(-100.0), Max: new(100.0)},
	// A day's activity
	"activities.DurationMinutes": {Min: new(1.0), Max: new(1440.0)},
	"activities.Calories":        {Max: new(20000.0)},
	"activities.Steps":           {Max: new(200000.0)},
	"activities.HeartRate":       {Min: new(20.0), Max: new(250.0)},
	// A profile switch to 10% to 500% of its insulin, for at most a week
	"profile_switches.Multiplier":      {Min: new(0.1), Max: new(5.0)},
	"profile_switches.DurationMinutes": {Max: new(10080.0)},
	// A portion of food has none of a nutrient taken back
	"food_presets.Carbs":    {Min: new(0.0)},
	"food_presets.Fat":      {Min: new(0.0)},
	"food_presets.Protein":  {Min: new(0.0)},
	"food_presets.Fibre":    {Min: new(0.0)},
	"food_presets.Energy":   {Min: new(0.0)},
	"food_presets.Portions": {Min: new(0.0)},
	"food_intakes.Energy":   {Min: new(0.0)},
	"food_intakes.Portions": {Min: new(0.0)},
	// A rate far above any pump's, for at most a day, the longest pumps set one for, of at least 1%
	"temp_basals.Rate":             {Max: new(50.0)},
	"temp_basals.DurationMinutes":  {Min: new(1.0), Max: new(1440.0)},
	"temp_basals.Multiplier":       {Min: new(0.01)},
	"temp_targets.LowTargetMgdl":   {Min: new(0.1)},
	"temp_targets.HighTargetMgdl":  {Min: new(0.1)},
	"temp_targets.DurationMinutes": {Min: new(1.0)},
	// An APNs key ID is 10 characters
	"push_credentials.KeyID": {MaxLength: 10},
}

// Derived are the limits of fields people enter in other units than their column keeps, by table and field: a dose's
// extended duration is typed in minutes and kept in milliseconds, at most a week
var Derived = map[string]sitedb.ColumnLimit{
	"insulin_doses.ExtendedDurationMinutes": {Min: new(0.0), Max: new(10080.0)},
}

// Column is the limit of a column, by its table and its field in the table's rows, e.g. insulin_doses.Dose, in the
// site's DB or the control's, or of a Derived field, false when there is none
func Column(column string) (sitedb.ColumnLimit, bool) {
	if limit, ok := Derived[column]; ok {
		return limit, true
	}
	table, field, _ := strings.Cut(column, ".")
	if limit, ok := sitedb.ColumnLimitOf(table, field); ok {
		return limit, true
	}
	return controldb.ColumnLimitOf(table, field)
}

// Of is what people may enter of a column: its limit, with its override's bounds and length in place of its own
func Of(column string) (sitedb.ColumnLimit, bool) {
	limit, ok := Column(column)
	override := Overrides[column]
	if override.Min != nil {
		limit.Min, limit.MinExclusive = override.Min, false
	}
	if override.Max != nil {
		limit.Max, limit.MaxExclusive = override.Max, false
	}
	if override.MaxLength > 0 {
		limit.MaxLength = override.MaxLength
	}
	return limit, ok
}
