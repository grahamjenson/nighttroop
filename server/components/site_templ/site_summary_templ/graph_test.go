package site_summary_templ

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// option is the chart as ECharts gets it, so the test reads what the browser would draw
func option(t *testing.T, s *sitemodels.Summary, now int64) map[string]any {
	chart := graph(s, now)
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))
	return option
}

func series(option map[string]any, name string) map[string]any {
	for _, s := range option["series"].([]any) {
		if s.(map[string]any)["name"] == name {
			return s.(map[string]any)
		}
	}
	return nil
}

func TestGraph(t *testing.T) {
	mmol := consts.UnitsMMOLL
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.Local).UnixMilli()
	since := now - 48*time.Hour.Milliseconds()
	rate := 2.0
	option := option(t, &sitemodels.Summary{
		Since: since, Units: &mmol,
		GlucoseReadings: []*sitemodels.SiteGlucoseReading{{Reading: 20, Units: &mmol}, {Reading: 6, Units: &mmol}, {Reading: 3, Units: &mmol}, {Reading: 12, Units: &mmol}},
		TempTargets:     []*sitemodels.SiteTempTarget{{LowTarget: 5, HighTarget: 7, OccurredAt: since, DurationMinutes: 60}},
		TempBasals:      []*sitemodels.SiteTempBasal{{Absolute: &rate, OccurredAt: now - time.Hour.Milliseconds(), DurationMinutes: 120}},
	}, now)

	// Every grid's time axis runs from the start of the summary to now, labelled only at the bottom
	xAxes := option["xAxis"].([]any)
	require.Len(t, xAxes, 3)
	for _, axis := range xAxes {
		require.Equal(t, "time", axis.(map[string]any)["type"])
		require.EqualValues(t, since, axis.(map[string]any)["min"])
		require.EqualValues(t, now, axis.(map[string]any)["max"])
	}
	require.Equal(t, false, xAxes[0].(map[string]any)["axisLabel"].(map[string]any)["show"])
	require.Equal(t, true, xAxes[2].(map[string]any)["axisLabel"].(map[string]any)["show"])

	// The glucose axis runs from 2 mmol/L (just under 40 mg/dL) to the highest reading
	glucoseAxis := option["yAxis"].([]any)[0].(map[string]any)
	require.EqualValues(t, 2, glucoseAxis["min"])
	require.EqualValues(t, 20, glucoseAxis["max"])

	// Readings are split by the target range, in the summary's units, and each bound is a dashed line
	tests := []struct {
		name     string
		colour   string
		readings []any
	}{
		{"Low", red600, []any{3.0}},
		{"In range", brandNight, []any{6.0}},
		{"High", amber500, []any{20.0, 12.0}},
	}
	for _, test := range tests {
		s := series(option, test.name)
		require.Equal(t, test.colour, s["itemStyle"].(map[string]any)["color"], test.name)
		readings := []any{}
		for _, point := range s["data"].([]any) {
			readings = append(readings, point.(map[string]any)["value"].([]any)[1])
		}
		require.Equal(t, test.readings, readings, test.name)
	}
	require.InDelta(t, 3.9, series(option, "Low")["markLine"].(map[string]any)["data"].([]any)[0].(map[string]any)["yAxis"], 0.05)
	require.InDelta(t, 10, series(option, "High")["markLine"].(map[string]any)["data"].([]any)[0].(map[string]any)["yAxis"], 0.05)

	// The target range is a band behind the readings, and a temp target shades its range for as long as it runs
	areas := series(option, "In range")["markArea"].(map[string]any)["data"].([]any)
	require.Len(t, areas, 2)
	band, target := areas[0].([]any), areas[1].([]any)
	require.Equal(t, []any{float64(since), float64(now)}, []any{band[0].(map[string]any)["coord"].([]any)[0], band[1].(map[string]any)["coord"].([]any)[0]})
	require.InDelta(t, 10, band[0].(map[string]any)["coord"].([]any)[1], 0.05)
	require.InDelta(t, 3.9, band[1].(map[string]any)["coord"].([]any)[1], 0.05)
	require.Equal(t, []any{float64(since), 7.0}, target[0].(map[string]any)["coord"])
	require.Equal(t, []any{float64(since + time.Hour.Milliseconds()), 5.0}, target[1].(map[string]any)["coord"])

	// A temp basal rises to its rate on its own axis, and one running past now ends at the graph's edge
	basal := series(option, "Temp basal")
	require.EqualValues(t, basalRate, basal["yAxisIndex"])
	require.EqualValues(t, 2, option["yAxis"].([]any)[basalRate].(map[string]any)["max"])
	points := []any{}
	for _, point := range basal["data"].([]any) {
		points = append(points, point.(map[string]any)["value"])
	}
	start := float64(now - time.Hour.Milliseconds())
	require.Equal(t, []any{[]any{start, 0.0}, []any{start, 2.0}, []any{float64(now), 2.0}, []any{float64(now), 0.0}}, points)
}

// Treatments a moment apart overlap on a long graph, so the latest is drawn last, on top where it can be read
func TestGraphDrawsTheLatestTreatmentOnTop(t *testing.T) {
	mgdl := consts.UnitsMGDL
	now := time.Now().UnixMilli()
	older, newer := 15.0, 40.0
	// Searches give the newest first
	summary := &sitemodels.Summary{
		Since: now - 48*time.Hour.Milliseconds(), Units: &mgdl,
		InsulinDoses: []*sitemodels.SiteInsulinDose{{Dose: 2.5, OccurredAt: now}, {Dose: 1, OccurredAt: now - time.Minute.Milliseconds()}},
		FoodIntakes:  []*sitemodels.SiteFoodIntake{{Carbs: &newer, OccurredAt: now}, {Carbs: &older, OccurredAt: now - time.Minute.Milliseconds()}},
	}
	option := option(t, summary, now)
	names := func(name string) []any {
		names := []any{}
		for _, point := range series(option, name)["data"].([]any) {
			names = append(names, point.(map[string]any)["name"])
		}
		return names
	}
	require.Equal(t, []any{"1U", "2.5U"}, names("Insulin"))
	require.Equal(t, []any{"15g", "40g"}, names("Carbs"))

	// The page gets the chart's element and the script that draws it into that element
	var html strings.Builder
	require.NoError(t, Graph(summary).Render(context.Background(), &html))
	require.Contains(t, html.String(), `id="site--summary--graph"`)
	require.Contains(t, html.String(), "echarts.init(document.getElementById(")
	require.Contains(t, html.String(), `"name":"2.5U"`)
}
