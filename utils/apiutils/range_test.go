package apiutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckRange(t *testing.T) {
	tests := []struct {
		name  string
		value *float64
		hint  string
	}{
		{"not sent", nil, ""},
		{"at min", new(0.0), ""},
		{"at max", new(1000.0), ""},
		{"below min", new(-1.0), "Constraint check failed: carbs, fat, protein and fibre must be from 0 to 1000 g"},
		{"above max", new(1000.5), "Constraint check failed: carbs, fat, protein and fibre must be from 0 to 1000 g"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			herr := CheckRange("errors.data.food", tt.value, 0, 1000)
			if tt.hint == "" {
				require.Nil(t, herr)
				return
			}
			require.Equal(t, tt.hint, herr.Message)
		})
	}
	// Without a unit the hint ends at the max, and a fraction is as it is
	require.Equal(t, "Constraint check failed: severity must be from 0 to 5", CheckRange("errors.data.severity", new(int64(6)), 0, 5).Message)
	require.Equal(t, "Constraint check failed: multiplier must be from 0.1 to 5", CheckRange("errors.data.multiplier", new(0.05), 0.1, 5).Message)
}
