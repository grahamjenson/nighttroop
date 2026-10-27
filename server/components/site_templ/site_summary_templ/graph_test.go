package site_summary_templ

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// option is the chart as ECharts gets it, so the test reads what the browser would draw
func option(t *testing.T, s *sitemodels.Summary, now int64) map[string]any {
	chart := graph(s, now, time.UTC)
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
	// Targets of 5 to 7 for the first hour and 6 to 8 for the next, then none; a basal of 1 U/hr for most of the
	// last hour then 2 U/hr, after none
	minutes := make([]sitemodels.SummaryMinute, 48*60+1)
	for i := range minutes {
		switch {
		case i < 60:
			minutes[i].LowTarget, minutes[i].HighTarget = new(5.0), new(7.0)
		case i < 120:
			minutes[i].LowTarget, minutes[i].HighTarget = new(6.0), new(8.0)
		}
		switch {
		case i >= 48*60-10:
			minutes[i].Basal = new(2.0)
		case i >= 47*60:
			minutes[i].Basal = new(1.0)
		}
	}
	option := option(t, &sitemodels.Summary{
		Since: since, Units: &mmol,
		GlucoseReadings: []*sitemodels.SiteGlucoseReading{{Reading: 20, Units: &mmol}, {Reading: 6, Units: &mmol}, {Reading: 3, Units: &mmol}, {Reading: 12, Units: &mmol}},
		Minutes:         minutes,
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

	// Each stretch of the same targets is a band behind the readings, from its high to its low, and none is drawn
	// where no target is in force
	areas := series(option, "In range")["markArea"].(map[string]any)["data"].([]any)
	require.Len(t, areas, 2)
	hour := float64(time.Hour.Milliseconds())
	for i, want := range [][]any{{float64(since), 7.0, float64(since) + hour, 5.0}, {float64(since) + hour, 8.0, float64(since) + 2*hour, 6.0}} {
		area := areas[i].([]any)
		require.Equal(t, []any{want[0], want[1]}, area[0].(map[string]any)["coord"])
		require.Equal(t, []any{want[2], want[3]}, area[1].(map[string]any)["coord"])
	}

	// The basal rises to its rate on its own axis, steps to the next rate, and falls back at the graph's edge
	basal := series(option, "Basal")
	require.EqualValues(t, basalRate, basal["yAxisIndex"])
	require.EqualValues(t, 2, option["yAxis"].([]any)[basalRate].(map[string]any)["max"])
	points := []any{}
	for _, point := range basal["data"].([]any) {
		points = append(points, point.(map[string]any)["value"])
	}
	start, step := float64(now)-hour, float64(now-10*time.Minute.Milliseconds())
	require.Equal(t, []any{[]any{start, 0.0}, []any{start, 1.0}, []any{step, 1.0}, []any{step, 2.0}, []any{float64(now), 2.0}, []any{float64(now), 0.0}}, points)
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

// TestGraphInTheLocationsTime checks each time is moved by the location's offset at that time, which ECharts,
// drawing in UTC, shows as the location's time
func TestGraphInTheLocationsTime(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	mmol := consts.UnitsMMOLL
	// New Zealand's daylight saving starts at 2am on 27 September 2026, from +12 to +13
	since := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC).UnixMilli()
	now := since + 24*time.Hour.Milliseconds()
	chart := graph(&sitemodels.Summary{
		Since: since, Units: &mmol,
		GlucoseReadings: []*sitemodels.SiteGlucoseReading{{Reading: 6, Units: &mmol, OccurredAt: now}},
	}, now, auckland)
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))

	axis := option["xAxis"].([]any)[0].(map[string]any)
	require.EqualValues(t, since+12*time.Hour.Milliseconds(), axis["min"])
	require.EqualValues(t, now+13*time.Hour.Milliseconds(), axis["max"])
	point := series(option, "In range")["data"].([]any)[0].(map[string]any)["value"].([]any)
	require.EqualValues(t, now+13*time.Hour.Milliseconds(), point[0])
	require.Contains(t, chart.JSFunctions.Fns, types.FuncStr("%MY_ECHARTS%.setOption({useUTC: true});"))
}
