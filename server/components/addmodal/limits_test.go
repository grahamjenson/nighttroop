package addmodal

import (
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// TestLimitAttributes checks each input takes its column's limits, narrowed by its override, as it is typed in
func TestLimitAttributes(t *testing.T) {
	mmol, mgdl := new(consts.UnitsMMOLL), new(consts.UnitsMGDL)
	tests := []struct {
		name  string
		field Field
		want  templ.Attributes
	}{
		{"a number takes its column's limits", Field{Column: "food_intakes.Carbs", Step: "any"},
			templ.Attributes{"step": "any", "min": "-1000", "max": "1000"}},
		{"its column's override narrows them", Field{Column: "insulin_doses.Dose", Step: "any"},
			templ.Attributes{"step": "any", "min": "-100", "max": "100"}},
		{"a number with no column has none", Field{Step: "1"}, templ.Attributes{"step": "1"}},
		{"a percent is its column's fraction as a percent", Field{Column: "temp_basals.Multiplier", Kind: Percent},
			templ.Attributes{"step": "any", "min": "1", "max": "1100", "data-percent": true}},
		{"a percent's override is a fraction", Field{Column: "profile_switches.Multiplier", Kind: Percent},
			templ.Attributes{"step": "any", "min": "10", "max": "500", "data-percent": true}},
		{"glucose in mmol/L, more than 0 as one decimal place, and 1000 mg/dL as 55.5 which is 999.999 mg/dL",
			Field{Column: "glucose_readings.Mgdl", Units: mmol, Step: "any"}, templ.Attributes{"step": "any", "min": "0.1", "max": "55.5"}},
		{"glucose in mg/dL", Field{Column: "glucose_readings.Mgdl", Units: mgdl, Step: "any"},
			templ.Attributes{"step": "any", "min": "0.1", "max": "1000"}},
		{"glucose's override is in mg/dL, rounded in to the units' place", Field{Column: "temp_targets.LowTargetMgdl", Units: mmol, Step: "any"},
			templ.Attributes{"step": "any", "min": "0.1", "max": "55.5"}},
		{"an excluded bound moves in by the step", Field{Column: "glucose_readings.Mgdl", Step: "1"},
			templ.Attributes{"step": "1", "min": "1", "max": "1000"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, numberAttributes(tt.field))
		})
	}
}

// TestTextLimits checks a text takes its column's length, narrowed by its override
func TestTextLimits(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		want  templ.Attributes
	}{
		{"its column's length", Field{Column: "insulin_doses.Notes", Kind: Text}, templ.Attributes{"maxlength": "1024"}},
		{"its override narrows it", Field{Column: "push_credentials.KeyID", Kind: Text}, templ.Attributes{"maxlength": "10"}},
		{"a column of one length", Field{Column: "device_reports.ConfigurationID", Kind: Text}, templ.Attributes{"minlength": "64", "maxlength": "64"}},
		{"with suggestions", Field{Column: "sites.ID", Kind: Text, Options: []string{"a"}}, templ.Attributes{"maxlength": "32", "list": "id--options"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, textAttributes(tt.field, "id"))
		})
	}
}

// TestCheck checks each number and text field names its column
func TestCheck(t *testing.T) {
	require.NoError(t, Check([]Field{
		{Name: "dose", Column: "insulin_doses.Dose"},
		{Name: "key_id", Column: "push_credentials.KeyID", Kind: Text},
		{Name: "occurred_at", Kind: When}, {Name: "reading_type", Kind: Select},
	}))
	require.Error(t, Check([]Field{{Name: "dose", Column: "insulin_doses.Doses"}}), "a column there is not")
	require.Error(t, Check([]Field{{Name: "notes", Kind: Text}}), "no column")
}
