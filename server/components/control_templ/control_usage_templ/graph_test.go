package control_usage_templ

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

// Each check is placed by its time across the checks kept and by its percent from the bottom, leaving out checks
// without the value; the newest's value is shown beside the graph
func TestResourcePoints(t *testing.T) {
	percent := func(p float64) *float64 { return &p }
	model := controlmodels.Usage{Now: 3000, ResourcesSince: 0, ResourceReadings: []controlmodels.ResourceReading{
		{At: 0, CPUPercent: 0, MemoryPercent: percent(100)},
		{At: 1500, CPUPercent: 50},
		{At: 3000, CPUPercent: 150, MemoryPercent: percent(25)},
	}}
	require.Equal(t, "0.0,60.0 150.0,30.0 300.0,0.0", resourcePoints(model, cpuPercent), "over 100% is drawn at the top")
	require.Equal(t, "0.0,0.0 300.0,45.0", resourcePoints(model, memoryPercent))
	require.Equal(t, "", resourcePoints(model, diskPercent))
	require.Equal(t, "25%", latestPercent(model, memoryPercent))
	require.Equal(t, "-", latestPercent(model, diskPercent))
	require.Equal(t, "-", latestPercent(controlmodels.Usage{}, cpuPercent))
}
