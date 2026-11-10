package site_reports_templ

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
)

func tddDays(doses map[string]float64, dates ...string) []*sitemodels.SiteSummaryDaily {
	days := []*sitemodels.SiteSummaryDaily{}
	for _, date := range dates {
		days = append(days, &sitemodels.SiteSummaryDaily{Date: date, TotalDailyDose: doses[date], TotalDailyBasal: doses[date] / 4})
	}
	return days
}

func TestTDDStats(t *testing.T) {
	tests := []struct {
		name                  string
		doses                 map[string]float64
		dates                 []string
		days                  int
		mean, lowest, highest *float64
	}{
		{name: "no days", days: 0},
		{name: "days without insulin are left out", doses: map[string]float64{"2026-11-02": 0}, dates: []string{"2026-11-02"}, days: 0},
		{name: "one day", doses: map[string]float64{"2026-11-02": 30.25}, dates: []string{"2026-11-02"}, days: 1,
			mean: new(30.3), lowest: new(30.25), highest: new(30.25)},
		{name: "mean, lowest and highest of the days with insulin", doses: map[string]float64{"2026-11-01": 20, "2026-11-03": 31, "2026-11-04": 25},
			dates: []string{"2026-11-01", "2026-11-02", "2026-11-03", "2026-11-04"}, days: 3, mean: new(25.3), lowest: new(20.0), highest: new(31.0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days := dosed(tddDays(tt.doses, tt.dates...))
			require.Len(t, days, tt.days)
			mean, lowest, highest := tddStats(days)
			require.Equal(t, tt.mean, mean)
			require.Equal(t, tt.lowest, lowest)
			require.Equal(t, tt.highest, highest)
		})
	}
}

// The chart has a bar of each day with insulin, its basal under its boluses, named by its date, on an axis of every
// date of the report, with no bar where a day has none, even before the first day with insulin
func TestTDDChart(t *testing.T) {
	days := dosed(tddDays(map[string]float64{"2026-10-31": 20, "2026-11-02": 31.5}, "2026-10-31", "2026-11-01", "2026-11-02"))
	chart := tddChart(days, date("2026-10-30"), date("2026-11-02"), date("2026-11-03"))
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))

	xAxis := option["xAxis"].([]any)[0].(map[string]any)
	require.Equal(t, []any{"30 Oct", "31 Oct", "1 Nov", "2 Nov"}, xAxis["data"])
	basal, boluses := option["series"].([]any)[0].(map[string]any), option["series"].([]any)[1].(map[string]any)
	require.Equal(t, []any{"Basal", "Bolus", "bar", "dose"}, []any{basal["name"], boluses["name"], basal["type"], basal["stack"]})
	require.Equal(t, basal["stack"], boluses["stack"])
	// Without today there is no so far label, so no room is kept for it
	require.Equal(t, "20", option["grid"].([]any)[0].(map[string]any)["right"])
	require.Equal(t, []any{
		map[string]any{"name": "Fri 30 Oct 2026", "value": "-"},
		map[string]any{"name": "Sat 31 Oct 2026", "value": 5.0},
		map[string]any{"name": "Sun 1 Nov 2026", "value": "-"},
		map[string]any{"name": "Mon 2 Nov 2026", "value": 7.875},
	}, basal["data"])
	require.Equal(t, []any{
		map[string]any{"name": "Fri 30 Oct 2026", "value": "-"},
		map[string]any{"name": "Sat 31 Oct 2026", "value": 15.0},
		map[string]any{"name": "Sun 1 Nov 2026", "value": "-"},
		map[string]any{"name": "Mon 2 Nov 2026", "value": 23.63},
	}, boluses["data"])
}

// Today's bar, not over, is paler and labelled so far, and named so for its tooltip; the days before it are not
func TestTDDChartToday(t *testing.T) {
	days := dosed(tddDays(map[string]float64{"2026-11-01": 20, "2026-11-02": 8}, "2026-11-01", "2026-11-02"))
	chart := tddChart(days, date("2026-11-01"), date("2026-11-02"), date("2026-11-02"))
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))
	basal, boluses := option["series"].([]any)[0].(map[string]any)["data"].([]any), option["series"].([]any)[1].(map[string]any)["data"].([]any)
	require.Equal(t, map[string]any{"name": "Sun 1 Nov 2026", "value": 5.0}, basal[0])
	require.Equal(t, map[string]any{"name": "Sun 1 Nov 2026", "value": 15.0}, boluses[0])
	require.Equal(t, map[string]any{"name": "Mon 2 Nov 2026, so far", "value": 2.0, "itemStyle": map[string]any{"opacity": 0.5}}, basal[1])
	require.Equal(t, map[string]any{"name": "Mon 2 Nov 2026, so far", "value": 6.0, "itemStyle": map[string]any{"opacity": 0.5},
		"label": map[string]any{"show": true, "position": "right", "formatter": "so far", "color": gray500}}, boluses[1])
	require.Equal(t, "52", option["grid"].([]any)[0].(map[string]any)["right"])
}

// A day's boluses are what of its daily dose is not its basal, to 2 decimal places, and never below none
func TestBolus(t *testing.T) {
	tests := []struct {
		name               string
		dose, basal, bolus float64
	}{
		{"no insulin", 0, 0, 0},
		{"only basal", 12.5, 12.5, 0},
		{"only boluses", 20, 0, 20},
		{"basal and boluses", 31.5, 7.875, 23.63},
		{"rounded", 10.333, 3.111, 7.22},
		{"basal more than the dose, as rounded", 4.99, 5, 0},
	}
	for _, tt := range tests {
		require.Equal(t, tt.bolus, bolus(tt.dose, tt.basal), tt.name)
	}
}

// The days over are those before today, so today so far, or a day after it, is left out of the report's numbers
func TestTDDOver(t *testing.T) {
	tests := []struct {
		name  string
		dates []string
		today string
		want  []string
	}{
		{"no days", nil, "2026-11-02", []string{}},
		{"all over", []string{"2026-10-31", "2026-11-01"}, "2026-11-02", []string{"2026-10-31", "2026-11-01"}},
		{"today left out", []string{"2026-11-01", "2026-11-02"}, "2026-11-02", []string{"2026-11-01"}},
		{"only today", []string{"2026-11-02"}, "2026-11-02", []string{}},
		{"a month before today", []string{"2026-10-01", "2026-10-02"}, "2026-11-02", []string{"2026-10-01", "2026-10-02"}},
	}
	for _, tt := range tests {
		got := []string{}
		for _, day := range tddOver(tddDays(map[string]float64{}, tt.dates...), date(tt.today)) {
			got = append(got, day.Date)
		}
		require.Equal(t, tt.want, got, tt.name)
	}
}

func date(day string) time.Time {
	at, _ := time.Parse(time.DateOnly, day)
	return at
}

// A report ends today, so far, unless its end names another day of the past year, as older days are not kept
func TestTDDEnd(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	// 11pm UTC on 6 November is already the 7th in Auckland, so today there is the 7th
	now := time.Date(2026, 11, 6, 23, 0, 0, 0, time.UTC)
	tests := []struct {
		end  string
		want string
		ok   bool
	}{
		{"", "2026-11-07", true},
		{"2026-11-07", "2026-11-07", true},
		{"2026-11-06", "2026-11-06", true},
		{"2026-10-01", "2026-10-01", true},
		{"2025-11-06", "2025-11-06", true},
		{"2026-11-08", "", false},
		{"2025-11-05", "", false},
		{"6 Nov 2026", "", false},
	}
	for _, tt := range tests {
		last, ok := TDDEnd(tt.end, now, auckland)
		require.Equal(t, tt.ok, ok, tt.end)
		if ok {
			require.Equal(t, tt.want, last.Format(time.DateOnly), tt.end)
			require.Equal(t, auckland, last.Location(), tt.end)
		}
	}
}

// The arrows go back and forward a month, the next no later than today, which has no end query, and none goes past
// today or before a year before yesterday
func TestTDDLinks(t *testing.T) {
	today := date("2026-11-07")
	tests := []struct {
		last           string
		previous, next string
	}{
		{"2026-11-07", "/reports/tdd?end=2026-10-08", ""},
		{"2026-11-06", "/reports/tdd?end=2026-10-07", "/reports/tdd"},
		{"2026-10-08", "/reports/tdd?end=2026-09-08", "/reports/tdd"},
		{"2026-10-20", "/reports/tdd?end=2026-09-20", "/reports/tdd"},
		{"2026-09-08", "/reports/tdd?end=2026-08-09", "/reports/tdd?end=2026-10-08"},
		{"2025-12-06", "/reports/tdd?end=2025-11-06", "/reports/tdd?end=2026-01-05"},
		{"2025-11-20", "", "/reports/tdd?end=2025-12-20"},
	}
	for _, tt := range tests {
		previous, next := tddLinks(date(tt.last), today)
		require.Equal(t, tt.previous, previous, tt.last)
		require.Equal(t, tt.next, next, tt.last)
	}
}

// The bars fill their days but for a gap of a couple of pixels, however many days there are
func TestTDDChartBarGaps(t *testing.T) {
	chart := tddChart(nil, date("2026-10-08"), date("2026-11-06"), date("2026-11-07"))
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))
	series := option["series"].([]any)[0].(map[string]any)
	require.Equal(t, "2", series["barCategoryGap"])
	require.Nil(t, series["barMaxWidth"])
	require.Len(t, option["xAxis"].([]any)[0].(map[string]any)["data"], TDDDays)
}

// The table has every day of the report, newest first, each with its daily dose and its basal and boluses, or none for
// a day without insulin or not kept, and today, with insulin or not, marked so far
func TestTDDRows(t *testing.T) {
	days := tddDays(map[string]float64{"2026-10-31": 20, "2026-11-01": 0, "2026-11-02": 31.5}, "2026-10-31", "2026-11-01", "2026-11-02")
	tests := []struct {
		name  string
		today string
		want  []tddRow
	}{
		{"today after the report", "2026-11-03", []tddRow{
			{Date: "2026-11-02", Dose: new(31.5), Basal: new(7.875), Bolus: new(23.63)},
			{Date: "2026-11-01"},
			{Date: "2026-10-31", Dose: new(20.0), Basal: new(5.0), Bolus: new(15.0)},
			{Date: "2026-10-30"},
		}},
		{"today with insulin so far", "2026-11-02", []tddRow{
			{Date: "2026-11-02", Dose: new(31.5), Basal: new(7.875), Bolus: new(23.63), SoFar: true},
			{Date: "2026-11-01"},
			{Date: "2026-10-31", Dose: new(20.0), Basal: new(5.0), Bolus: new(15.0)},
			{Date: "2026-10-30"},
		}},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tddRows(days, date("2026-10-30"), date("2026-11-02"), date(tt.today)), tt.name)
	}
	// Today without insulin so far is still marked
	require.Equal(t, tddRow{Date: "2026-11-01", SoFar: true}, tddRows(days, date("2026-10-30"), date("2026-11-01"), date("2026-11-01"))[0])
}
