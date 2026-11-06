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
	minute := time.Minute.Milliseconds()
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
		Since: since, Units: &mmol, RangeLow: 3.9, RangeHigh: 10, RangeUrgentLow: 3, RangeUrgentHigh: 14,
		// The last two readings are more than 15 minutes apart
		GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 20, Units: &mmol, OccurredAt: now}}, {SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 6, Units: &mmol, OccurredAt: now - 5*minute}}, {SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 3, Units: &mmol, OccurredAt: now - 10*minute}}, {SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 12, Units: &mmol, OccurredAt: now - 30*minute}}},
		Minutes:         minutes,
	}, now)

	// Everything is drawn on one grid, as Carer draws it, its time axis from the start of the summary to now
	xAxes := option["xAxis"].([]any)
	require.Len(t, xAxes, 1)
	require.Equal(t, "time", xAxes[0].(map[string]any)["type"])
	require.EqualValues(t, since, xAxes[0].(map[string]any)["min"])
	require.EqualValues(t, now, xAxes[0].(map[string]any)["max"])

	// The glucose axis runs from 2 mmol/L (just under 40 mg/dL) to the highest reading, with a fifth again above it
	// for the treatments
	glucoseAxis := option["yAxis"].([]any)[glucoseAxis].(map[string]any)
	require.EqualValues(t, 2, glucoseAxis["min"])
	require.EqualValues(t, 24, glucoseAxis["max"])

	// The readings are joined by a thin line, broken where they are more than 15 minutes apart
	line := []any{}
	for _, point := range series(option, "Glucose")["data"].([]any) {
		line = append(line, point.(map[string]any)["value"])
	}
	require.Equal(t, []any{[]any{float64(now), 20.0}, []any{float64(now - 5*minute), 6.0}, []any{float64(now - 10*minute), 3.0}, "-", []any{float64(now - 30*minute), 12.0}}, line)

	// Readings are split by the summary's low and high alarms, in its units, and each is a dashed line
	tests := []struct {
		name     string
		colour   string
		readings []any
	}{
		{"Low", glucoseLow, []any{3.0}},
		{"In range", glucoseInRange, []any{6.0}},
		{"High", glucoseHigh, []any{20.0, 12.0}},
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

	// The urgent range and the range are bands across the graph, and each stretch of the same targets is a band over
	// them, from its high to its low, none drawn where no target is in force
	areas := series(option, "In range")["markArea"].(map[string]any)["data"].([]any)
	require.Len(t, areas, 4)
	hour := float64(time.Hour.Milliseconds())
	for i, want := range [][]any{{float64(since), 14.0, float64(now), 3.0}, {float64(since), 10.0, float64(now), 3.9}, {float64(since), 7.0, float64(since) + hour, 5.0}, {float64(since) + hour, 8.0, float64(since) + 2*hour, 6.0}} {
		area := areas[i].([]any)
		require.Equal(t, []any{want[0], want[1]}, area[0].(map[string]any)["coord"])
		require.Equal(t, []any{want[2], want[3]}, area[1].(map[string]any)["coord"])
	}

	// The basal rises to its rate on its own axis, its highest rate a fifth of the way up and marked, steps to the next
	// rate, and falls back at the graph's edge
	basal := series(option, "Basal")
	require.EqualValues(t, basalAxis, basal["yAxisIndex"])
	require.EqualValues(t, 10, option["yAxis"].([]any)[basalAxis].(map[string]any)["max"])
	require.EqualValues(t, 2, basal["markLine"].(map[string]any)["data"].([]any)[0].(map[string]any)["yAxis"])
	points := []any{}
	for _, point := range basal["data"].([]any) {
		points = append(points, point.(map[string]any)["value"])
	}
	start, step := float64(now)-hour, float64(now-10*time.Minute.Milliseconds())
	require.Equal(t, []any{[]any{start, 0.0}, []any{start, 1.0}, []any{step, 1.0}, []any{step, 2.0}, []any{float64(now), 2.0}, []any{float64(now), 0.0}}, points)
}

// Treatments a moment apart overlap on a long graph, and ECharts keeps the first of labels drawn over each other, so the
// latest is first, where it can be read
func TestGraphLabelsTheLatestTreatment(t *testing.T) {
	mgdl := consts.UnitsMGDL
	now := time.Now().UnixMilli()
	older, newer := 15.0, 40.0
	// Searches give the newest first
	summary := &sitemodels.Summary{
		Since: now - 48*time.Hour.Milliseconds(), Units: &mgdl,
		InsulinDoses: []*sitemodels.SiteInsulinDose{{Dose: 2.5, OccurredAt: now}, {Dose: 0.2, DoseType: new("smb"), OccurredAt: now - time.Minute.Milliseconds()}, {Dose: 1, OccurredAt: now - time.Minute.Milliseconds()}},
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
	require.Equal(t, []any{"2.5U", "1U"}, names("Insulin"))
	require.Equal(t, []any{"40g", "15g"}, names("Carbs"))
	require.Equal(t, true, series(option, "Carbs")["labelLayout"].(map[string]any)["hideOverlap"])

	// A loop's SMB is drawn smaller and fainter than insulin given by hand, and not labelled, as they come every few
	// minutes; carbs are drawn above insulin
	require.Equal(t, []any{"0.2U"}, names("SMB"))
	require.Equal(t, false, series(option, "SMB")["label"].(map[string]any)["show"])
	require.Equal(t, true, series(option, "Insulin")["label"].(map[string]any)["show"])
	require.Less(t, series(option, "SMB")["symbolSize"].(float64), series(option, "Insulin")["symbolSize"].(float64))
	height := func(name string) float64 {
		return series(option, name)["data"].([]any)[0].(map[string]any)["value"].([]any)[1].(float64)
	}
	require.Greater(t, height("Carbs"), height("Insulin"))
	require.EqualValues(t, treatmentAxis, series(option, "Carbs")["yAxisIndex"])

	// The page gets the chart's element and the script that draws it into that element
	var html strings.Builder
	require.NoError(t, Graph(summary).Render(context.Background(), &html))
	require.Contains(t, html.String(), `id="site--summary--graph"`)
	require.Contains(t, html.String(), "echarts.init(document.getElementById(")
	require.Contains(t, html.String(), `"name":"2.5U"`)

	// With no basal in force, no highest rate is labelled
	require.Nil(t, series(option, "Basal")["markLine"])
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
		Since: since, Units: &mmol, RangeLow: 3.9, RangeHigh: 10,
		GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 6, Units: &mmol, OccurredAt: now}}},
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
