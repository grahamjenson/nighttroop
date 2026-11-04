package nightscoutmodels

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A profile switch's Nightscout percentage is kept as a multiplier of its insulin and returned as the same percentage
func TestPercentageToMultiplier(t *testing.T) {
	tests := []struct {
		name                   string
		percentage, multiplier float64
	}{
		{name: "as it is", percentage: 100, multiplier: 1},
		{name: "more", percentage: 120, multiplier: 1.2},
		{name: "less", percentage: 90, multiplier: 0.9},
		{name: "a fraction", percentage: 33.333, multiplier: 0.33333},
		{name: "none", percentage: 0, multiplier: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			percentage := types.FlexibleFloat64(test.percentage)
			multiplier := PercentageToMultiplier(&percentage)
			require.Equal(t, test.multiplier, *multiplier)
			require.Equal(t, percentage, *MultiplierToPercentage(multiplier))
		})
	}
	require.Nil(t, PercentageToMultiplier(nil))
	require.Nil(t, MultiplierToPercentage(nil))
}
