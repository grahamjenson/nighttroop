package addmodal

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// TestLimitAttributes checks each input takes its column's limits, narrowed by its own, as it is typed in
func TestLimitAttributes(t *testing.T) {
	mmol, mgdl := new(consts.UnitsMMOLL), new(consts.UnitsMGDL)
	tests := []struct {
		name  string
		field Field
		want  templ.Attributes
	}{
		{"a number takes its column's limits", Field{Column: "insulin_doses.Dose", Step: "any"},
			templ.Attributes{"step": "any", "min": "-500", "max": "500"}},
		{"its own Min and Max narrow them", Field{Column: "insulin_doses.Dose", Min: "-100", Max: "100", Step: "any"},
			templ.Attributes{"step": "any", "min": "-100", "max": "100"}},
		{"a number with no column has only its own", Field{Max: "5", Step: "1"}, templ.Attributes{"step": "1", "max": "5"}},
		{"a percent is its column's fraction as a percent", Field{Column: "insulin_doses.SplitNow", Kind: Percent},
			templ.Attributes{"step": "any", "min": "0", "max": "100", "data-percent": true}},
		{"a percent's own limits are fractions", Field{Column: "profile_switches.Multiplier", Kind: Percent, Min: "0.1", Max: "5"},
			templ.Attributes{"step": "any", "min": "10", "max": "500", "data-percent": true}},
		{"glucose in mmol/L, more than 0 as one decimal place, and 1000 mg/dL as 55.5 which is 999.999 mg/dL",
			Field{Column: "glucose_readings.Mgdl", Units: mmol, Step: "any"}, templ.Attributes{"step": "any", "min": "0.1", "max": "55.5"}},
		{"glucose in mg/dL", Field{Column: "glucose_readings.Mgdl", Units: mgdl, Step: "any"},
			templ.Attributes{"step": "any", "min": "0.1", "max": "1000"}},
		{"glucose's own limits are in mg/dL, rounded in to the units' place", Field{Column: "temp_targets.LowTargetMgdl", Units: mmol, Min: "0.1", Max: "100", Step: "any"},
			templ.Attributes{"step": "any", "min": "0.1", "max": "5.5"}},
		{"an excluded bound moves in by the step", Field{Column: "glucose_readings.Mgdl", Step: "1"},
			templ.Attributes{"step": "1", "min": "1", "max": "1000"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, numberAttributes(tt.field))
		})
	}
}

// TestTextLimits checks a text takes its column's length, narrowed by its own Max
func TestTextLimits(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		want  templ.Attributes
	}{
		{"its column's length", Field{Column: "insulin_doses.Notes", Kind: Text}, templ.Attributes{"maxlength": "1024"}},
		{"its own Max narrows it", Field{Column: "push_credentials.KeyID", Kind: Text, Max: "10"}, templ.Attributes{"maxlength": "10"}},
		{"a column of one length", Field{Column: "device_reports.ConfigurationID", Kind: Text}, templ.Attributes{"minlength": "64", "maxlength": "64"}},
		{"with suggestions", Field{Column: "sites.ID", Kind: Text, Options: []string{"a"}}, templ.Attributes{"maxlength": "32", "list": "id--options"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, textAttributes(tt.field, "id"))
		})
	}
}

// TestCheck checks a field's own limits may only narrow its column's, which must be there
func TestCheck(t *testing.T) {
	require.NoError(t, Check([]Field{
		{Name: "dose", Column: "insulin_doses.Dose", Min: "-500", Max: "100"},
		{Name: "multiplier", Column: "profile_switches.Multiplier", Kind: Percent, Min: "0.1", Max: "5"},
		{Name: "key_id", Column: "push_credentials.KeyID", Kind: Text, Max: "10"},
		{Name: "occurred_at", Kind: When}, {Name: "reading_type", Kind: Select},
	}))
	tests := []struct {
		name  string
		field Field
	}{
		{"a min below its column's", Field{Name: "dose", Column: "insulin_doses.Dose", Min: "-600"}},
		{"a max above its column's", Field{Name: "dose", Column: "insulin_doses.Dose", Max: "600"}},
		{"a min at an excluded bound", Field{Name: "reading", Column: "glucose_readings.Mgdl", Min: "0"}},
		{"a percent above its column's fraction", Field{Name: "multiplier", Column: "temp_basals.Multiplier", Kind: Percent, Max: "1100"}},
		{"a min that is not a number", Field{Name: "dose", Column: "insulin_doses.Dose", Min: "low"}},
		{"a text longer than its column", Field{Name: "key_id", Column: "push_credentials.KeyID", Kind: Text, Max: "33"}},
		{"a column there is not", Field{Name: "dose", Column: "insulin_doses.Doses"}},
		{"no column", Field{Name: "notes", Kind: Text}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, Check([]Field{tt.field}))
		})
	}
}
