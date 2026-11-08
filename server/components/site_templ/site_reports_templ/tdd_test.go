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
		days = append(days, &sitemodels.SiteSummaryDaily{Date: date, TotalDailyDose: doses[date]})
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

// The chart has a bar of each day with insulin, named by its date, on an axis of every date of the report, with no bar
// where a day has none, even before the first day with insulin
func TestTDDChart(t *testing.T) {
	days := dosed(tddDays(map[string]float64{"2026-10-31": 20, "2026-11-02": 31.5}, "2026-10-31", "2026-11-01", "2026-11-02"))
	chart := tddChart(days, date("2026-10-30"), date("2026-11-02"))
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))

	xAxis := option["xAxis"].([]any)[0].(map[string]any)
	require.Equal(t, []any{"30 Oct", "31 Oct", "1 Nov", "2 Nov"}, xAxis["data"])
	series := option["series"].([]any)[0].(map[string]any)
	require.Equal(t, "bar", series["type"])
	require.Equal(t, []any{
		map[string]any{"name": "Fri 30 Oct 2026", "value": "-"},
		map[string]any{"name": "Sat 31 Oct 2026", "value": 20.0},
		map[string]any{"name": "Sun 1 Nov 2026", "value": "-"},
		map[string]any{"name": "Mon 2 Nov 2026", "value": 31.5},
	}, series["data"])
}

func date(day string) time.Time {
	at, _ := time.Parse(time.DateOnly, day)
	return at
}

// A report ends yesterday unless its end names another day of the past year: today is not over, and older days are not
// kept
func TestTDDEnd(t *testing.T) {
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	// 11pm UTC on 6 November is already the 7th in Auckland, so yesterday there is the 6th
	now := time.Date(2026, 11, 6, 23, 0, 0, 0, time.UTC)
	tests := []struct {
		end  string
		want string
		ok   bool
	}{
		{"", "2026-11-06", true},
		{"2026-11-06", "2026-11-06", true},
		{"2026-10-01", "2026-10-01", true},
		{"2025-11-06", "2025-11-06", true},
		{"2026-11-07", "", false},
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

// The arrows go back and forward a month, the next no later than yesterday, which has no end query, and none goes past
// yesterday or before a year ago
func TestTDDLinks(t *testing.T) {
	yesterday := date("2026-11-06")
	tests := []struct {
		last           string
		previous, next string
	}{
		{"2026-11-06", "/reports/tdd?end=2026-10-07", ""},
		{"2026-10-07", "/reports/tdd?end=2026-09-07", "/reports/tdd"},
		{"2026-10-20", "/reports/tdd?end=2026-09-20", "/reports/tdd"},
		{"2026-09-07", "/reports/tdd?end=2026-08-08", "/reports/tdd?end=2026-10-07"},
		{"2025-11-20", "", "/reports/tdd?end=2025-12-20"},
	}
	for _, tt := range tests {
		previous, next := tddLinks(date(tt.last), yesterday)
		require.Equal(t, tt.previous, previous, tt.last)
		require.Equal(t, tt.next, next, tt.last)
	}
}

// The bars fill their days but for a gap of a couple of pixels, however many days there are
func TestTDDChartBarGaps(t *testing.T) {
	chart := tddChart(nil, date("2026-10-08"), date("2026-11-06"))
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))
	series := option["series"].([]any)[0].(map[string]any)
	require.Equal(t, "2", series["barCategoryGap"])
	require.Nil(t, series["barMaxWidth"])
	require.Len(t, option["xAxis"].([]any)[0].(map[string]any)["data"], TDDDays)
}

// The table has every day of the report, newest first, each with its daily dose, or none for a day without insulin or
// not kept
func TestTDDRows(t *testing.T) {
	days := tddDays(map[string]float64{"2026-10-31": 20, "2026-11-01": 0, "2026-11-02": 31.5}, "2026-10-31", "2026-11-01", "2026-11-02")
	require.Equal(t, []tddRow{
		{Date: "2026-11-02", Dose: new(31.5)},
		{Date: "2026-11-01"},
		{Date: "2026-10-31", Dose: new(20.0)},
		{Date: "2026-10-30"},
	}, tddRows(days, date("2026-10-30"), date("2026-11-02")))
}
