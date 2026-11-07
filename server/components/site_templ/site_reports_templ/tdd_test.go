package site_reports_templ

import (
	"encoding/json"
	"testing"

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

// The chart has a bar of each day with insulin, named by its date, on an axis of every date from the first to the last,
// with no bar where a day has none
func TestTDDChart(t *testing.T) {
	days := dosed(tddDays(map[string]float64{"2026-10-31": 20, "2026-11-02": 31.5}, "2026-10-31", "2026-11-01", "2026-11-02"))
	chart := tddChart(days)
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))

	xAxis := option["xAxis"].([]any)[0].(map[string]any)
	require.Equal(t, []any{"31 Oct", "1 Nov", "2 Nov"}, xAxis["data"])
	series := option["series"].([]any)[0].(map[string]any)
	require.Equal(t, "bar", series["type"])
	require.Equal(t, []any{
		map[string]any{"name": "Sat 31 Oct 2026", "value": 20.0},
		map[string]any{"name": "Sun 1 Nov 2026", "value": "-"},
		map[string]any{"name": "Mon 2 Nov 2026", "value": 31.5},
	}, series["data"])
}
