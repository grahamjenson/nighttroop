package site_reports_templ

import (
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/css"
	"nighttroop.t1d.nz/utils/consts"
)

// ECharts draws the TDD chart with these colours from css/theme.css, as it cannot use the Tailwind classes
var (
	brandNight = css.Colour("brand-night")
	brandGold  = css.Colour("brand-gold")
	gray200    = css.Colour("gray-200")
)

// TDDDays is how many days the TDD report covers, a month, as a week or fortnight is too few to see a pattern in
const TDDDays = 30

// tddYesterday is the newest day a TDD report can end on, as today is not over
func tddYesterday(now time.Time, location *time.Location) time.Time {
	y, m, d := now.In(location).AddDate(0, 0, -1).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, location)
}

// tddKept is whether a TDD report can end on last: no later than yesterday and no earlier than a year before it, as
// older days are not kept
func tddKept(last, yesterday time.Time) bool {
	return !last.After(yesterday) && !last.Before(yesterday.AddDate(-1, 0, 0))
}

// TDDEnd is the last day of the TDD report its end query names, as 2026-01-31, or yesterday when it names none, and
// whether a report can end on it
func TDDEnd(end string, now time.Time, location *time.Location) (time.Time, bool) {
	yesterday := tddYesterday(now, location)
	if end == "" {
		return yesterday, true
	}
	last, err := time.ParseInLocation(time.DateOnly, end, location)
	return last, err == nil && tddKept(last, yesterday)
}

// tddLinks are the TDD reports of the month before and after the one ending on last, "" where there is none; the one
// after ends on yesterday at the latest, the one ending yesterday having no end query
func tddLinks(last, yesterday time.Time) (previous, next string) {
	link := func(at time.Time) string {
		switch {
		case !tddKept(at, yesterday):
			return ""
		case at.Equal(yesterday):
			return "/reports/tdd"
		}
		return "/reports/tdd?end=" + at.Format(time.DateOnly)
	}
	previous = link(last.AddDate(0, 0, -TDDDays))
	if last.Before(yesterday) {
		after := last.AddDate(0, 0, TDDDays)
		if after.After(yesterday) {
			after = yesterday
		}
		next = link(after)
	}
	return previous, next
}

// dates are the days from first to last, oldest first
func dates(first, last time.Time) []time.Time {
	all := []time.Time{}
	for at := first; !at.After(last); at = at.AddDate(0, 0, 1) {
		all = append(all, at)
	}
	return all
}

// tddRow is a day of the TDD report's table: its date and its daily dose, with the basal and boluses of it, nil when
// it had no insulin
type tddRow struct {
	Date               string
	Dose, Basal, Bolus *float64
}

// tddRows are each day from first to last, newest first, with its daily dose and the basal and boluses of it when it
// had insulin
func tddRows(days []*sitemodels.SiteSummaryDaily, first, last time.Time) []tddRow {
	byDate := map[string]*sitemodels.SiteSummaryDaily{}
	for _, day := range dosed(days) {
		byDate[day.Date] = day
	}
	rows := []tddRow{}
	for _, at := range slices.Backward(dates(first, last)) {
		row := tddRow{Date: at.Format(time.DateOnly)}
		if day, ok := byDate[row.Date]; ok {
			row.Dose, row.Basal, row.Bolus = &day.TotalDailyDose, &day.TotalDailyBasal, new(bolus(day.TotalDailyDose, day.TotalDailyBasal))
		}
		rows = append(rows, row)
	}
	return rows
}

// bolus is the boluses of a daily dose, what is not its basal, to 2 decimal places
func bolus(dose, basal float64) float64 {
	return consts.Round(max(dose-basal, 0), 2)
}

// dosed are the days with insulin, as a day without any has no doses recorded, rather than none given
func dosed(days []*sitemodels.SiteSummaryDaily) []*sitemodels.SiteSummaryDaily {
	return slices.DeleteFunc(slices.Clone(days), func(day *sitemodels.SiteSummaryDaily) bool { return day.TotalDailyDose <= 0 })
}

// tddStats are the days' mean, lowest and highest daily dose, in U, or nil when there are no days
func tddStats(days []*sitemodels.SiteSummaryDaily) (mean, lowest, highest *float64) {
	if len(days) == 0 {
		return nil, nil, nil
	}
	sum, low, high := 0.0, days[0].TotalDailyDose, days[0].TotalDailyDose
	for _, day := range days {
		sum, low, high = sum+day.TotalDailyDose, min(low, day.TotalDailyDose), max(high, day.TotalDailyDose)
	}
	return new(consts.Round(sum/float64(len(days)), 1)), &low, &high
}

// tddChart is a bar of each day's dose, its basal under its boluses, each named by its date for the tooltip, on an
// axis of every date of the report from first to last, with no bar where a day had no insulin
func tddChart(days []*sitemodels.SiteSummaryDaily, first, last time.Time) *charts.Bar {
	byDate := map[string]*sitemodels.SiteSummaryDaily{}
	for _, day := range days {
		byDate[day.Date] = day
	}
	dates, basals, boluses := []string{}, []opts.BarData{}, []opts.BarData{}
	for at := first; !at.After(last); at = at.AddDate(0, 0, 1) {
		basal, boluss := opts.BarData{Name: at.Format("Mon 2 Jan 2006"), Value: "-"}, opts.BarData{Name: at.Format("Mon 2 Jan 2006"), Value: "-"}
		if day, ok := byDate[at.Format(time.DateOnly)]; ok {
			basal.Value, boluss.Value = day.TotalDailyBasal, bolus(day.TotalDailyDose, day.TotalDailyBasal)
		}
		dates, basals, boluses = append(dates, at.Format("2 Jan")), append(basals, basal), append(boluses, boluss)
	}
	chart := charts.NewBar()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "0"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: opts.FuncOpts("function (p) { return p.data.name + '<br/>' + p.seriesName + ' ' + p.data.value + ' U'; }")}),
		charts.WithGridOpts(opts.Grid{Left: "50", Right: "20", Top: "10", Bottom: "56"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: 0, SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}}, AxisLabel: &opts.AxisLabel{Formatter: "{value} U"}}),
	)
	// The bars fill their days, but for a gap of a couple of pixels, so a month of them reads at a glance
	narrowGaps := func(s *charts.SingleSeries) { s.BarCategoryGap = "2" }
	stacked := charts.WithBarChartOpts(opts.BarChart{Stack: "dose"})
	chart.SetXAxis(dates).
		AddSeries("Basal", basals, charts.WithItemStyleOpts(opts.ItemStyle{Color: brandGold}), stacked, narrowGaps).
		AddSeries("Bolus", boluses, charts.WithItemStyleOpts(opts.ItemStyle{Color: brandNight}), stacked, narrowGaps)
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
}
