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
		{"below min", new(-1.0), "Constraint check failed: carbs must be from 0 to 1000 g"},
		{"above max", new(1000.5), "Constraint check failed: carbs must be from 0 to 1000 g"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			herr := CheckRange("carbs", tt.value, 0, 1000, "g")
			if tt.hint == "" {
				require.Nil(t, herr)
				return
			}
			require.Equal(t, tt.hint, herr.Message)
		})
	}
	// Without a unit the hint ends at the max
	require.Equal(t, "Constraint check failed: severity must be from 0 to 5", CheckRange("severity", new(int64(6)), 0, 5, "").Message)
}
