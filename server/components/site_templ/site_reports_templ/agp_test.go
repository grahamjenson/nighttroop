package site_reports_templ

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// The AGP's bands are an area between two percentiles, and its lines a path, each broken where it has none
func TestAGPPaths(t *testing.T) {
	mgdl := consts.UnitsMGDL
	g := graph{Width: 288, Height: 100, Left: 0, Right: 288 * 5, Top: 0, Bottom: 100, Min: 0, Max: 100, Units: &mgdl}
	agp := [][]float64{{10, 20, 30, 40, 50}, {10, 20, 30, 40, 50}, {}, {0, 0, 0, 0, 100}}
	// Each slot is 5 across, as the graph is 5 for each of a day's 288 5 minutes
	require.Equal(t, "M0.0,50.0 L5.0,50.0 L5.0,90.0 L0.0,90.0 Z M15.0,0.0 L15.0,100.0 Z", bandPath(g, agp, 0, 4))
	require.Equal(t, "M0.0,70.0 L5.0,70.0 M15.0,100.0", linePath(g, percentile(agp, 2)))
	require.Equal(t, [][2]int{{0, 1}, {3, 3}}, runs(4, func(slot int) bool { return len(agp[slot]) > 0 }))
	slot, last := lastPercentiles(agp)
	require.Equal(t, 3, slot)
	require.Equal(t, []float64{0, 0, 0, 0, 100}, last)
	slot, _ = lastPercentiles([][]float64{{}})
	require.Equal(t, -1, slot)
	require.True(t, hasGlucose(&sitemodels.SiteSummaryPeriod{AGP: agp}))
	require.False(t, hasGlucose(&sitemodels.SiteSummaryPeriod{AGP: [][]float64{{}, {}}}))
}

// The AGP graph goes up to at least 350 mg/dL, or a day's highest glucose above it, and marks very low, the range,
// very high and its top
func TestAGPGraph(t *testing.T) {
	mgdl := consts.UnitsMGDL
	require.Equal(t, 350.0, newAGPGraph(&sitemodels.SiteSummaryPeriod{Units: &mgdl, AGP: [][]float64{{1, 2, 3, 4, 200}}}).Max)
	day := &sitemodels.SiteSummaryPeriod{Period: "day", Units: &mgdl, Glucose: []*float64{new(390.0)}}
	require.Equal(t, 400.0, newAGPGraph(day).Max)
	require.Equal(t, []float64{54, 70, 180, 250, 400}, agpMarks(newAGPGraph(day), 70, 180))
	// Percentile labels go beside their lines, moved up to stay 12 apart when the lines are close
	g := graph{Top: 0, Bottom: 100, Min: 0, Max: 100, Units: &mgdl}
	require.Equal(t, []float64{94, 82, 50, 38, 26}, percentileLabelYs(g, []float64{10, 11, 54, 55, 56}))
	require.Equal(t, 30.0, *sum(new(10.0), new(20.0)))
	require.Equal(t, 10.0, *sum(new(10.0), nil))
	require.Nil(t, sum(nil, nil))
}

// A period's daily profiles are its days in weeks from Monday, the first week's days before its first blank, each
// with its summary if it has one
func TestProfileWeeks(t *testing.T) {
	location := time.UTC
	// November 2026 starts on a Sunday
	start, end := time.Date(2026, 11, 1, 0, 0, 0, 0, location), time.Date(2026, 12, 1, 0, 0, 0, 0, location)
	month := &sitemodels.SiteSummaryPeriod{Period: "month", Start: start.UnixMilli(), End: end.UnixMilli()}
	weeks := profileWeeks(context.Background(), month, []*sitemodels.SiteSummaryDaily{{Date: "2026-11-02"}}, location)
	require.Len(t, weeks, 6)
	require.Equal(t, "Mon 26 Oct", weeks[0][0].Date.Format("Mon 2 Jan"))
	require.False(t, weeks[0][5].InPeriod)
	require.True(t, weeks[0][6].InPeriod)
	require.Nil(t, weeks[0][6].Day)
	require.Equal(t, "2026-11-02", weeks[1][0].Day.Date)
	require.Len(t, weeks[5], 1)

	// A day is one day, from itself
	day := &sitemodels.SiteSummaryPeriod{Period: "day", Start: start.UnixMilli(), End: start.AddDate(0, 0, 1).UnixMilli()}
	weeks = profileWeeks(context.Background(), day, nil, location)
	require.Len(t, weeks, 1)
	require.Len(t, weeks[0], 1)
	require.True(t, weeks[0][0].InPeriod)
}
