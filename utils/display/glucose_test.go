package display

import (
	"testing"

	"nighttroop.t1d.nz/utils/consts"
)

func TestPercent(t *testing.T) {
	number := func(v float64) *float64 { return &v }
	tests := []struct {
		value *float64
		want  string
	}{{nil, "–"}, {number(0), "0%"}, {number(0.4), "<1%"}, {number(0.5), "1%"}, {number(33.3), "33%"}, {number(79.6), "80%"}, {number(100), "100%"}}
	for _, tt := range tests {
		if got := Percent(tt.value); got != tt.want {
			t.Errorf("Percent(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestGlucose(t *testing.T) {
	mmol, mgdl := consts.UnitsMMOLL, consts.UnitsMGDL
	number := func(v float64) *float64 { return &v }
	tests := []struct {
		name  string
		value *float64
		units *consts.Units
		want  string
	}{
		{"mmol/L to one decimal place", number(5.54), &mmol, "5.5"},
		{"mmol/L whole", number(7), &mmol, "7.0"},
		{"mg/dL whole", number(99.6), &mgdl, "100"},
		{"no units as mg/dL", number(120.2), nil, "120"},
		{"none", nil, &mmol, "–"},
	}
	for _, tt := range tests {
		if got := GlucoseOr(tt.value, tt.units); got != tt.want {
			t.Errorf("%s: GlucoseOr() = %q, want %q", tt.name, got, tt.want)
		}
	}
	for units, want := range map[*consts.Units]string{nil: "", &mmol: "mmol/L", &mgdl: "mg/dL"} {
		if got := Units(units); got != want {
			t.Errorf("Units(%v) = %q, want %q", units, got, want)
		}
	}
}

func TestGlucoseRange(t *testing.T) {
	for reading, want := range map[float64]string{3.8: "Low", 3.9: "", 7: "", 10: "", 10.1: "High"} {
		if got := GlucoseRange(reading, 3.9, 10); got != want {
			t.Errorf("GlucoseRange(%v) = %q, want %q", reading, got, want)
		}
	}
}

func TestScale(t *testing.T) {
	for value, want := range map[float64]float64{40: 0, 0: 0, 140: 0.5, 240: 1, 400: 1} {
		if got := Scale(value, 40, 240); got != want {
			t.Errorf("Scale(%v) = %v, want %v", value, got, want)
		}
	}
}
