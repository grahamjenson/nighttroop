package display

import (
	"context"
	"testing"

	"nighttroop.t1d.nz/utils/consts"
)

func TestNumber(t *testing.T) {
	mmol := consts.UnitsMMOLL
	tests := []struct {
		format string
		number string
		want   string
	}{
		{"", "1234.5", "1,234.5"},
		{"unknown", "1234.5", "1,234.5"},
		{"1,234.5", "-1234567.25", "-1,234,567.25"},
		{"1.234,5", "1234.5", "1.234,5"},
		{"1.234,5", "5.4", "5,4"},
		{"1 234,5", "-1234.5", "-1 234,5"},
		{"1 234,5", "123", "123"},
		{"1.234,5", "1000", "1.000"},
	}
	for _, tt := range tests {
		if got := Number(WithNumberFormat(context.Background(), tt.format), tt.number); got != tt.want {
			t.Errorf("Number(%q, %q) = %q, want %q", tt.format, tt.number, got, tt.want)
		}
	}
	// Glucose and values are shown in the number format, and a glucose's value as it is typed
	comma := WithNumberFormat(context.Background(), "1.234,5")
	if got := Glucose(comma, 5.44, &mmol); got != "5,4" {
		t.Errorf("Glucose() = %q, want 5,4", got)
	}
	if got := GlucoseValue(5.44, &mmol); got != "5.4" {
		t.Errorf("GlucoseValue() = %q, want 5.4", got)
	}
	if got := PerDay(comma, 7, 1234.567, " U"); got != "1.234,57 U" {
		t.Errorf("PerDay() = %q, want 1.234,57 U", got)
	}
}
