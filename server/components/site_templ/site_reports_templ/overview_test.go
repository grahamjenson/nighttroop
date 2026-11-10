package site_reports_templ

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

func TestPercentText(t *testing.T) {
	for _, tt := range []struct {
		value *float64
		want  string
	}{{nil, "–"}, {new(0.0), "0%"}, {new(0.4), "<1%"}, {new(0.5), "1%"}, {new(79.6), "80%"}, {new(100.0), "100%"}} {
		require.Equal(t, tt.want, percentText(tt.value))
	}
}

// The best day is the whole day with the most time in range, the later of two the same, never today
func TestBestDay(t *testing.T) {
	now := time.UnixMilli(10 * 86_400_000)
	day := func(date string, end int64, inRange *float64) *sitemodels.SiteSummaryDaily {
		return &sitemodels.SiteSummaryDaily{Date: date, End: end, TimeInRange: inRange}
	}
	require.Nil(t, bestDay(nil, now))
	require.Nil(t, bestDay([]*sitemodels.SiteSummaryDaily{day("none", 1, nil)}, now))
	days := []*sitemodels.SiteSummaryDaily{day("first", 1, new(80.0)), day("second", 2, new(90.0)), day("third", 3, new(90.0)),
		day("fourth", 4, new(50.0)), day("today", now.UnixMilli()+1, new(100.0))}
	require.Equal(t, "third", bestDay(days, now).Date)
}

// Each bar is 15 minutes: the 25th and 75th percentiles of the AGP at its middle 5 minutes, and its glucose's mean
func TestOverviewBars(t *testing.T) {
	agp := make([][]float64, 288)
	agp[1], agp[4] = []float64{3, 4, 5, 6, 7}, []float64{}
	glucose := make([]*float64, 288)
	glucose[0], glucose[2], glucose[5] = new(5.0), new(6.0), new(9.0)
	bars := overviewBars(&sitemodels.SiteSummaryPeriod{AGP: agp, Glucose: glucose})
	require.Len(t, bars, 96)
	require.Equal(t, overviewBar{Slot: 0, Lower: 4, Upper: 6, Mean: new(5.5), HasPercentiles: true}, bars[0])
	require.Equal(t, overviewBar{Slot: 1, Mean: new(9.0)}, bars[1])
	require.Equal(t, overviewBar{Slot: 2}, bars[2])
	require.Equal(t, 6.0, overviewTop(bars))
}

// A period's days so far are all of them once it is over, and up to today, which has begun, while it is not
func TestPeriodDays(t *testing.T) {
	location := time.UTC
	start := time.Date(2026, 11, 2, 0, 0, 0, 0, location)
	week := &sitemodels.SiteSummaryPeriod{Start: start.UnixMilli(), End: start.AddDate(0, 0, 7).UnixMilli()}
	require.Equal(t, 7, periodDays(week, start.AddDate(0, 1, 0), location))
	require.Equal(t, 3, periodDays(week, start.AddDate(0, 0, 2).Add(time.Hour), location))
	require.Equal(t, 0, periodDays(week, start.Add(-time.Hour), location))
}

// A graph marks every 50 mg/dL or 2.5 mmol/L, from 40 mg/dL to at least 300 mg/dL
func TestGraphTicks(t *testing.T) {
	mgdl, mmol := consts.UnitsMGDL, consts.UnitsMMOLL
	require.Equal(t, []float64{50, 100, 150, 200, 250, 300}, newGraph(960, 300, &mgdl, 0).Ticks())
	require.Equal(t, []float64{50, 100, 150, 200, 250, 300, 350}, newGraph(960, 300, &mgdl, 320).Ticks())
	require.Equal(t, []float64{2.5, 5, 7.5, 10, 12.5, 15}, newGraph(960, 300, &mmol, 0).Ticks())
	g := newGraph(960, 300, &mgdl, 0)
	require.Equal(t, g.Bottom, g.Y(10))
	require.Equal(t, g.Top, g.Y(400))
	require.Equal(t, g.Right, g.X(24*60*60_000))
}

// The next week or fortnight is its 7 or 14 days after, or the days to today when fewer are over, with none after
// today's; the next day or month is the one after
func TestNext(t *testing.T) {
	today := time.Now().UTC()
	date := func(days int) string { return today.AddDate(0, 0, days).Format(time.DateOnly) }
	week, fortnight := &sitemodels.SiteSummaryPeriod{Period: "week"}, &sitemodels.SiteSummaryPeriod{Period: "fortnight"}
	require.Equal(t, "/reports/agp/week-"+date(-13), next("agp", week, today.AddDate(0, 0, -20)))
	require.Equal(t, "/reports/overview/week-"+date(0), next("overview", week, today.AddDate(0, 0, -3)))
	require.Equal(t, "/reports/agp/fortnight-"+date(-6), next("agp", fortnight, today.AddDate(0, 0, -20)))
	require.Equal(t, "", next("agp", week, today))
	require.Equal(t, "/reports/agp/day-"+date(-1), next("agp", &sitemodels.SiteSummaryPeriod{Period: "day"}, today.AddDate(0, 0, -2)))
	require.Equal(t, "", next("agp", &sitemodels.SiteSummaryPeriod{Period: "day"}, today))
}

// A number's change from the report before is up or down by how much, to its decimal places, or none when the same
// or when either is missing; the report before is named by its kind
func TestChange(t *testing.T) {
	require.Equal(t, "▲ 0.3 mmol/L", change(new(8.4), new(8.1), 1, " mmol/L"))
	require.Equal(t, "▼ 4%", change(new(60.0), new(63.6), 0, "%"))
	require.Equal(t, "No change", change(new(6.9), new(6.94), 1, "%"))
	require.Equal(t, "", change(new(6.9), nil, 1, "%"))
	require.Equal(t, "", change(nil, new(6.9), 1, "%"))
	require.Equal(t, "the previous 14 days", previousName(&sitemodels.SiteSummaryPeriod{Period: "fortnight"}))
	require.Equal(t, "the previous month", previousName(&sitemodels.SiteSummaryPeriod{Period: "month"}))
	require.Equal(t, "the day before", previousName(&sitemodels.SiteSummaryPeriod{Period: "day"}))
	mmol := consts.UnitsMMOLL
	require.Equal(t, "▲ 0.5 mmol/L", glucoseChange(&sitemodels.SiteSummaryPeriod{Units: &mmol, Mean: new(7.5)}, &sitemodels.SiteSummaryPeriod{Mean: new(7.0)}))
	require.Equal(t, "", glucoseChange(&sitemodels.SiteSummaryPeriod{Units: &mmol, Mean: new(7.5)}, nil))
}
