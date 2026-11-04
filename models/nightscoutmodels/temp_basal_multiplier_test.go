package nightscoutmodels

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A Nightscout percent is kept as a multiplier of the basal and returned as the same percent
func TestPercentToMultiplier(t *testing.T) {
	tests := []struct {
		name                string
		percent, multiplier float64
	}{
		{name: "half", percent: -50, multiplier: 0.5},
		{name: "a zero temp", percent: -100, multiplier: 0},
		{name: "the basal", percent: 0, multiplier: 1},
		{name: "over the basal", percent: 20, multiplier: 1.2},
		{name: "a fraction", percent: -33.333, multiplier: 0.66667},
		{name: "the most", percent: 1000, multiplier: 11},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			percent := types.FlexibleFloat64(test.percent)
			multiplier := PercentToMultiplier(&percent)
			require.Equal(t, test.multiplier, *multiplier)
			require.Equal(t, percent, *MultiplierToPercent(multiplier))
		})
	}
	require.Nil(t, PercentToMultiplier(nil))
	require.Nil(t, MultiplierToPercent(nil))
}
