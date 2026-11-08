package site_reports_templ

import (
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// tddRanges are what the TDD report can cover: the week, fortnight or month before today
var tddRanges = []struct {
	Name, Label string
	Days        int
}{{"week", "Week", 7}, {"fortnight", "Fortnight", 14}, {"month", "Month", 30}}

// TDDRangeDays is how many days before today the TDD report's range covers, or 0 when there is no such range
func TDDRangeDays(name string) int {
	for _, r := range tddRanges {
		if r.Name == name {
			return r.Days
		}
	}
	return 0
}

// dates are the days from first to last, oldest first
func dates(first, last time.Time) []time.Time {
	all := []time.Time{}
	for at := first; !at.After(last); at = at.AddDate(0, 0, 1) {
		all = append(all, at)
	}
	return all
}

// tddRow is a day of the TDD report's table: its date and its daily dose, nil when it had no insulin
type tddRow struct {
	Date string
	Dose *float64
}

// tddRows are each day from first to last, newest first, with its daily dose when it had insulin
func tddRows(days []*sitemodels.SiteSummaryDaily, first, last time.Time) []tddRow {
	doses := map[string]float64{}
	for _, day := range dosed(days) {
		doses[day.Date] = day.TotalDailyDose
	}
	rows := []tddRow{}
	for _, at := range slices.Backward(dates(first, last)) {
		row := tddRow{Date: at.Format(time.DateOnly)}
		if dose, ok := doses[row.Date]; ok {
			row.Dose = &dose
		}
		rows = append(rows, row)
	}
	return rows
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

// tddChart is a bar of each day's dose, named by its date for the tooltip, on an axis of every date of the report from
// first to last, with no bar where a day had no insulin
func tddChart(days []*sitemodels.SiteSummaryDaily, first, last time.Time) *charts.Bar {
	doses := map[string]float64{}
	for _, day := range days {
		doses[day.Date] = day.TotalDailyDose
	}
	dates, bars := []string{}, []opts.BarData{}
	for at := first; !at.After(last); at = at.AddDate(0, 0, 1) {
		bar := opts.BarData{Name: at.Format("Mon 2 Jan 2006"), Value: "-"}
		if dose, ok := doses[at.Format(time.DateOnly)]; ok {
			bar.Value = dose
		}
		dates, bars = append(dates, at.Format("2 Jan")), append(bars, bar)
	}
	chart := charts.NewBar()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: opts.FuncOpts("function (p) { return p.data.name + '<br/>' + p.data.value + ' U'; }")}),
		charts.WithGridOpts(opts.Grid{Left: "50", Right: "20", Top: "10", Bottom: "30"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: 0, SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}}, AxisLabel: &opts.AxisLabel{Formatter: "{value} U"}}),
	)
	maxWidth := func(s *charts.SingleSeries) { s.BarMaxWidth = "24" }
	chart.SetXAxis(dates).AddSeries("Daily dose", bars, charts.WithItemStyleOpts(opts.ItemStyle{Color: brandNight}), maxWidth)
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
}
