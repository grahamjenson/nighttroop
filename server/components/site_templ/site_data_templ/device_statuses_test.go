package site_data_templ

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A device's values, like an IOB summed from basal and bolus IOB, are shown to at most two decimal places
func TestWithUnit(t *testing.T) {
	tests := []struct {
		value *float64
		want  string
	}{
		{nil, "-"},
		{new(1.65 + 0.0000000000000001), "1.65 U"},
		{new(0.5), "0.5 U"},
		{new(128.0), "128 U"},
		{new(1.006), "1.01 U"},
	}
	for _, test := range tests {
		require.Equal(t, test.want, withUnit(test.value, " U"))
	}
}
