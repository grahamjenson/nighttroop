package site_reports_templ

import (
	"math"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/css"
)

// ECharts draws with these colours from css/theme.css, as it cannot use the Tailwind classes
var (
	brandNight  = css.Colour("brand-night")
	blue500     = css.Colour("blue-500")
	glucoseLow  = css.Colour("glucose-low")
	glucoseHigh = css.Colour("glucose-high")
	brandGold   = css.Colour("brand-gold")
	gray200     = css.Colour("gray-200")
)

// hasGlucose is whether the period has any glucose to draw an AGP of
func hasGlucose(p *sitemodels.SiteSummaryPeriod) bool {
	for _, percentiles := range p.AGP {
		if len(percentiles) == len(sitemodels.AGPPercentiles) {
			return true
		}
	}
	return false
}

// agp is the period's AGP over a day from midnight: the 5th to 95th percentiles shaded, the 25th to 75th shaded
// darker, the median as a line, and the range from low to high behind them, with its bounds dashed. A day's own
// glucose is drawn over its AGP. Each band is stacked on the percentile below it, so its area is between them, and a
// time with no glucose breaks it.
//
// The times are milliseconds from midnight, drawn in UTC, so the axis shows them as the time of day.
func agp(p *sitemodels.SiteSummaryPeriod, low, high float64) *charts.Line {
	chart := charts.NewLine()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "0", Data: []string{"5–95%", "25–75%", "Median", "Glucose"}}),
		charts.WithGridOpts(opts.Grid{Left: "50", Right: "20", Top: "10", Bottom: "56"}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time", Min: 0, Max: 86_400_000,
			AxisLabel: &opts.AxisLabel{Formatter: "{HH}:{mm}"},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}},
		}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: math.Floor(p.Units.FromMGDL(40)), Max: math.Ceil(max(p.Units.FromMGDL(270), percentileMax(p))), SplitLine: &opts.SplitLine{Show: opts.Bool(false)}, AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false)}}),
	)
	chart.AddJSFuncStrs("%MY_ECHARTS%.setOption({useUTC: true});")

	// percentile is the AGP's ith percentile less the one under it, so stacked on that one it reaches the ith
	percentile := func(i, under int) []opts.LineData {
		data := []opts.LineData{}
		for slot, percentiles := range p.AGP {
			value := any("-")
			if len(percentiles) == len(sitemodels.AGPPercentiles) {
				value = percentiles[i]
				if under >= 0 {
					value = percentiles[i] - percentiles[under]
				}
			}
			data = append(data, opts.LineData{Value: []any{int64(slot) * sitemodels.AGPSlotMs, value}})
		}
		return data
	}
	hidden := []charts.SeriesOpts{charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false)}), charts.WithLineStyleOpts(opts.LineStyle{Opacity: opts.Float(0)})}
	band := func(stack string, opacity float32) []charts.SeriesOpts {
		return []charts.SeriesOpts{
			charts.WithLineChartOpts(opts.LineChart{Stack: stack, ShowSymbol: opts.Bool(false)}),
			charts.WithLineStyleOpts(opts.LineStyle{Opacity: opts.Float(0)}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: blue500, Opacity: opts.Float(opacity)}),
			charts.WithAreaStyleOpts(opts.AreaStyle{Color: blue500, Opacity: opts.Float(opacity)}),
		}
	}
	chart.AddSeries("5th", percentile(0, -1), append(hidden, charts.WithLineChartOpts(opts.LineChart{Stack: "outer", ShowSymbol: opts.Bool(false)}))...)
	chart.AddSeries("5–95%", percentile(4, 0), band("outer", 0.15)...)
	chart.AddSeries("25th", percentile(1, -1), append(hidden, charts.WithLineChartOpts(opts.LineChart{Stack: "inner", ShowSymbol: opts.Bool(false)}))...)
	chart.AddSeries("25–75%", percentile(3, 1), band("inner", 0.3)...)

	dashed := func(colour string, at float64) []charts.SeriesOpts {
		return []charts.SeriesOpts{
			charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Type: "dashed", Color: colour, Width: 1}}),
			charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{YAxis: at}),
		}
	}
	median := []charts.SeriesOpts{
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 2}),
		charts.WithItemStyleOpts(opts.ItemStyle{Color: brandNight}),
		charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(false)}}),
		charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Coordinate0: []any{0, high}, Coordinate1: []any{86_400_000, low}, ItemStyle: &opts.ItemStyle{Color: brandNight, Opacity: opts.Float(0.04)}}),
	}
	chart.AddSeries("Median", percentile(2, -1), append(median, dashed(glucoseLow, low)...)...)
	chart.AddSeries("High", nil, dashed(glucoseHigh, high)...)

	if p.Period == "day" {
		var glucose []opts.LineData
		for slot, value := range p.Glucose {
			point := opts.LineData{Value: "-"}
			if value != nil {
				point = opts.LineData{Value: []any{int64(slot) * sitemodels.SummaryDailySlotMs, *value}}
			}
			glucose = append(glucose, point)
		}
		chart.AddSeries("Glucose", glucose, charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false)}),
			charts.WithLineStyleOpts(opts.LineStyle{Color: brandGold, Width: 2}), charts.WithItemStyleOpts(opts.ItemStyle{Color: brandGold}))
	}
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
}

// percentileMax is the period's highest 95th percentile, or a day's highest glucose, so the axis shows all of them
func percentileMax(p *sitemodels.SiteSummaryPeriod) float64 {
	top := 0.0
	for _, percentiles := range p.AGP {
		if len(percentiles) > 0 {
			top = max(top, percentiles[len(percentiles)-1])
		}
	}
	for _, value := range p.Glucose {
		if value != nil && p.Period == "day" {
			top = max(top, *value)
		}
	}
	return top
}
