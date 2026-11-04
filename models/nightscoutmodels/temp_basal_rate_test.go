package nightscoutmodels

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A Nightscout percent is kept as a rate of the basal and returned as the same percent
func TestPercentToRate(t *testing.T) {
	tests := []struct {
		name          string
		percent, rate float64
	}{
		{name: "half", percent: -50, rate: 0.5},
		{name: "a zero temp", percent: -100, rate: 0},
		{name: "the basal", percent: 0, rate: 1},
		{name: "over the basal", percent: 20, rate: 1.2},
		{name: "a fraction", percent: -33.333, rate: 0.66667},
		{name: "the most", percent: 1000, rate: 11},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			percent := types.FlexibleFloat64(test.percent)
			rate := PercentToRate(&percent)
			require.Equal(t, test.rate, *rate)
			require.Equal(t, percent, *RateToPercent(rate))
		})
	}
	require.Nil(t, PercentToRate(nil))
	require.Nil(t, RateToPercent(nil))
}
