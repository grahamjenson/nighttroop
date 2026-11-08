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

	// The glucose axis runs from 0 to 16 mmol/L whatever the readings, so a high one does not squash the rest: the 20
	// is drawn past the top, but only 12% of the axis past it, and the glucose is not clipped at the axis
	glucoseAxis := option["yAxis"].([]any)[glucoseAxis].(map[string]any)
	require.EqualValues(t, 0, glucoseAxis["min"])
	require.EqualValues(t, 16, glucoseAxis["max"])
	over := 16 * 1.12

	// The readings are joined by a thin line, broken where they are more than 15 minutes apart
	line := []any{}
	for _, point := range series(option, "Glucose")["data"].([]any) {
		line = append(line, point.(map[string]any)["value"])
	}
	require.Equal(t, []any{[]any{float64(now), over}, []any{float64(now - 5*minute), 6.0}, []any{float64(now - 10*minute), 3.0}, "-", []any{float64(now - 30*minute), 12.0}}, line)

	// Readings are split by the summary's low and high alarms, in its units, and each is a dashed line
	tests := []struct {
		name     string
		colour   string
		readings []any
	}{
		{"Low", glucoseLow, []any{3.0}},
		{"In range", glucoseInRange, []any{6.0}},
		{"High", glucoseHigh, []any{over, 12.0}},
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

	// The targets in force are edged by darker lines: the low target as steps, broken where none is in force, then the
	// high target the same way
	target := series(option, "Target")
	require.Equal(t, blue700, target["lineStyle"].(map[string]any)["color"])
	targetPoints := []any{}
	for _, point := range target["data"].([]any) {
		targetPoints = append(targetPoints, point.(map[string]any)["value"])
	}
	start := float64(since)
	require.Equal(t, []any{
		[]any{start, 5.0}, []any{start + hour, 5.0}, []any{start + hour, 6.0}, []any{start + 2*hour, 6.0}, "-", "-",
		[]any{start, 7.0}, []any{start + hour, 7.0}, []any{start + hour, 8.0}, []any{start + 2*hour, 8.0}, "-",
	}, targetPoints)

	// The basal rises to its rate on its own axis, its highest rate as high as 3 mmol/L up the glucose axis and marked,
	// steps to the next rate, and falls back at the graph's edge
	basal := series(option, "Basal")
	require.EqualValues(t, basalAxis, basal["yAxisIndex"])
	require.InDelta(t, 2*16.0/3, option["yAxis"].([]any)[basalAxis].(map[string]any)["max"], 0.001)
	require.EqualValues(t, 2, basal["markLine"].(map[string]any)["data"].([]any)[0].(map[string]any)["yAxis"])
	points := []any{}
	for _, point := range basal["data"].([]any) {
		points = append(points, point.(map[string]any)["value"])
	}
	basalStart, step := float64(now)-hour, float64(now-10*time.Minute.Milliseconds())
	require.Equal(t, []any{[]any{basalStart, 0.0}, []any{basalStart, 1.0}, []any{step, 1.0}, []any{step, 2.0}, []any{float64(now), 2.0}, []any{float64(now), 0.0}}, points)

	// Pointing at a time shows what was then, from the axis, not only on a reading or treatment
	tooltip := option["tooltip"].(map[string]any)
	require.Equal(t, "axis", tooltip["trigger"])
	require.Equal(t, false, tooltip["axisPointer"].(map[string]any)["snap"])
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
	setUp := ""
	for _, fn := range chart.JSFunctions.Fns {
		setUp += string(fn)
	}
	require.Contains(t, setUp, "useUTC: true")
	// The glucose axis is marked in quarters, so 16 is marked and nothing crowds it
	require.Contains(t, setUp, "yAxis: [{interval: 4}]")
	// The glucose is drawn past the axis's top, which go-echarts cannot set
	for _, id := range []string{"low", "high", "in-range", "glucose"} {
		require.Contains(t, setUp, "{id: '"+id+"', clip: false}")
	}
}

// The key under the graph draws each as the graph does: insulin and SMBs as triangles pointing down, carbs as ones
// pointing up, and the basal as an outlined area, not the line through a circle a legend draws for a line by default
func TestGraphLegendDrawsEachAsTheGraphDoes(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC).UnixMilli()
	got := option(t, &sitemodels.Summary{Hours: 6, Since: now - 6*time.Hour.Milliseconds(), Units: new(consts.UnitsMGDL)}, now)
	icons := map[string]any{}
	for _, item := range got["legend"].(map[string]any)["data"].([]any) {
		icons[item.(map[string]any)["name"].(string)] = item.(map[string]any)["icon"]
	}
	require.Equal(t, map[string]any{"Insulin": pointingDown, "SMB": pointingDown, "Carbs": pointingUp, "Basal": "rect", "Target": "path://M0,4L12,4L12,6L0,6Z"}, icons)
	// Each names a series of the graph, so clicking it hides that series
	for name := range icons {
		require.NotNil(t, series(got, name), name)
	}
	// Insulin points down on the graph too, and carbs up
	require.Equal(t, "triangle", series(got, "Insulin")["symbol"])
	require.Equal(t, "triangle", series(got, "Carbs")["symbol"])
}

// The tooltip looks up, for the time pointed at, each reading written as the tables write it, each stretch of a basal
// and of targets, and each treatment, all in the location's time as the graph draws them
func TestGraphHoverData(t *testing.T) {
	mmol := consts.UnitsMMOLL
	minute := time.Minute.Milliseconds()
	since := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC).UnixMilli()
	now := since + 3*minute
	at := func(ms int64) int64 { return ms + 1000 }
	s := &sitemodels.Summary{
		Since: since, Units: &mmol,
		GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: 6.23, Units: &mmol, OccurredAt: since + minute}}},
		Minutes: []sitemodels.SummaryMinute{
			{Basal: new(0.8), LowTarget: new(5.0), HighTarget: new(6.0)},
			{Basal: new(0.8), LowTarget: new(5.0), HighTarget: new(6.0)},
			{Basal: new(1.25)},
			{},
		},
		InsulinDoses: []*sitemodels.SiteInsulinDose{{Dose: 2.5, OccurredAt: since + 2*minute}, {Dose: 0.2, DoseType: new("SMB"), OccurredAt: since}},
		FoodIntakes:  []*sitemodels.SiteFoodIntake{{Carbs: new(15.0), OccurredAt: since + 2*minute}},
	}
	script := hoverTooltip(s, now, at)
	data := strings.SplitN(strings.SplitN(script, "var d = ", 2)[1], ", near", 2)[0]
	var got hover
	require.NoError(t, json.Unmarshal([]byte(strings.ReplaceAll(data, "'", `"`)), &got))
	require.Equal(t, hover{
		Units:    "mmol/L",
		Readings: [][]any{{float64(at(since + minute)), "6.2"}},
		Basal:    [][]any{{float64(at(since)), float64(at(since + 2*minute)), "0.8"}, {float64(at(since + 2*minute)), float64(at(since + 3*minute)), "1.25"}},
		Target:   [][]any{{float64(at(since)), float64(at(since + 2*minute)), "5.0–6.0"}},
		Treatments: [][]any{
			{float64(at(since + 2*minute)), "Insulin: 2.5U"}, {float64(at(since)), "SMB: 0.2U"}, {float64(at(since + 2*minute)), "Carbs: 15g"},
		},
	}, got)
	// The formatter is made once with its data, and reads the time pointed at from the axis
	require.True(t, strings.HasPrefix(script, "(function () {"))
	require.Contains(t, script, "axisValue")

	// The page's script has no escaped double quote, which would stop the browser running it and drawing the graph
	page := graph(s, now, time.UTC).RenderSnippet().Script
	require.Contains(t, page, "var d = {'units':'mmol/L'")
	require.NotContains(t, page, `\"`)
}
