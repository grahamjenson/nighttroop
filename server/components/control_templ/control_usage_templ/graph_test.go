package control_usage_templ

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

// Each check is placed by its time across the checks kept and by its percent from the bottom to the graph's top,
// leaving out checks without the value; the newest's value is shown beside the graph
func TestResourcePoints(t *testing.T) {
	at := func(p float64) *float64 { return &p }
	model := controlmodels.Usage{Now: 3000, ResourcesSince: 0, ResourceReadings: []controlmodels.ResourceReading{
		{At: 0, CPUPercent: 0, MemoryPercent: at(100)},
		{At: 1500, CPUPercent: 50},
		{At: 3000, CPUPercent: 150, MemoryPercent: at(25)},
	}}
	require.Equal(t, "0.0,60.0 150.0,30.0 300.0,0.0", resourcePoints(model, cpuPercent, 100), "over the top is drawn at the top")
	require.Equal(t, "0.0,60.0 150.0,45.0 300.0,15.0", resourcePoints(model, cpuPercent, 200), "the CPU of 2 cores, up to 200%")
	require.Equal(t, "0.0,0.0 300.0,45.0", resourcePoints(model, memoryPercent, 100))
	require.Equal(t, "", resourcePoints(model, diskPercent, 100))
	require.Equal(t, "25%", latestValue(model, memoryPercent, percent))
	require.Equal(t, "-", latestValue(model, diskPercent, percent))
	require.Equal(t, "-", latestValue(controlmodels.Usage{}, cpuPercent, percent))
}

// Of many checks in one unit across, only the lowest and the highest are drawn, in time order, so a graph of 3 days of
// checks keeps its peaks in a few hundred points
func TestResourcePointsKeepEachUnitsLowestAndHighest(t *testing.T) {
	model := controlmodels.Usage{Now: 3000, ResourcesSince: 0}
	for i, cpu := range []float64{50, 60, 10, 40, 90, 20} { // all in the unit at 0, then one at 300
		model.ResourceReadings = append(model.ResourceReadings, controlmodels.ResourceReading{At: int64(i), CPUPercent: cpu})
	}
	model.ResourceReadings = append(model.ResourceReadings, controlmodels.ResourceReading{At: 3000, CPUPercent: 30})
	require.Equal(t, "0.2,54.0 0.4,6.0 300.0,42.0", resourcePoints(model, cpuPercent, 100), "10% then 90%, then the next unit's")
}

// Hovering a graph's column shows its check nearest the column's middle, and how long ago it was, and nothing where the
// column has no check
func TestHoverValues(t *testing.T) {
	now := int64(60 * 60 * 1000)
	model := controlmodels.Usage{Now: now, ResourcesSince: 0, ResourceReadings: []controlmodels.ResourceReading{
		{At: 0, CPUPercent: 10}, {At: now / hoverColumns / 2, CPUPercent: 20}, {At: now, CPUPercent: 30},
	}}
	shown := hoverValues(t.Context(), model, cpuPercent, percent)
	require.Len(t, shown, hoverColumns)
	require.Equal(t, "20% · 59m ago", shown[0], "the check at the column's middle")
	require.Equal(t, "", shown[1])
	require.Equal(t, "30% · 0m ago", shown[hoverColumns-1], "the newest is in the last column")
}
