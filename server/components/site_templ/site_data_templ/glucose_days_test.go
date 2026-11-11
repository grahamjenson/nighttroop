package site_data_templ

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The glucose days are Today alone on a new site's first day, not "Last 1 days"
func TestGlucoseDaysTitle(t *testing.T) {
	require.Equal(t, "Today", glucoseDaysTitle(1))
	require.Equal(t, "Last 2 days", glucoseDaysTitle(2))
	require.Equal(t, "Last 8 days", glucoseDaysTitle(8))
}

// A reading's mark is drawn wholly inside its day's graph, from the top edge for the axis's top to the bottom edge for
// its bottom, so a CGM's lowest reading is seen rather than drawn below the graph
func TestGlucoseDaysMarkIsInsideTheGraph(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		want  string
	}{
		{"the axis's bottom sits on the bottom edge", 40, "97.0"},
		{"below the axis is kept on the bottom edge", 20, "97.0"},
		{"the axis's top sits on the top edge", 270, "0.0"},
		{"the middle of the axis", 155, "48.5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := glucoseDaysMarkY(tc.value, 40, 270); got != tc.want {
				t.Errorf("glucoseDaysMarkY(%v) = %s, want %s", tc.value, got, tc.want)
			}
		})
	}
}
