package site_summary_templ

import (
	"math"
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/css"
)

const (
	// The glucose axis always shows up to at least this, so a flat day is not stretched
	minAxisMgdl = 270.0
	// The glucose axis starts here, as readings below it are rare
	axisStartMgdl = 40.0
	// The axes of the three grids: glucose on top, then a row each of carbs and insulin, then temp basals
	glucoseX               = 0
	treatmentX, treatmentY = 1, 1
	basalX, basalRate      = 2, 3
)

// ECharts draws with these colours from css/theme.css, as it cannot use the Tailwind classes
var (
	brandNight     = css.Colour("brand-night")
	amber500       = css.Colour("amber-500")
	glucoseLow     = css.Colour("glucose-low")
	glucoseHigh    = css.Colour("glucose-high")
	glucoseInRange = css.Colour("glucose-in-range")
	blue500    = css.Colour("blue-500")
	gray200    = css.Colour("gray-200")
	gray700    = css.Colour("gray-700")
)

// glucoseRange is a reading in the summary's units below its low alarm, "Low", above its high alarm, "High", or
// between them, ""
func glucoseRange(s *sitemodels.Summary, reading float64) string {
	switch {
	case reading < s.RangeLow:
		return "Low"
	case reading > s.RangeHigh:
		return "High"
	}
	return ""
}

// graph is the summary as a go-echarts chart on one time axis: glucose dots over the targets in force each minute,
// shaded, insulin as triangles pointing down and carbs as triangles pointing up under it, and the basal rate in force
// each minute as a line at the bottom. Treatments are added oldest first, so the latest is drawn on top where markers
// close together overlap.
//
// ECharts shows times in the browser's timezone, or in UTC, so each time is moved by the location's offset
// from UTC at that time, and drawn in UTC: the axis and tooltips then show the location's time.
func graph(s *sitemodels.Summary, now int64, location *time.Location) *charts.Scatter {
	at := func(ms int64) int64 {
		_, offset := time.UnixMilli(ms).In(location).Zone()
		return ms + int64(offset)*1000
	}
	glucoseMax, basalMax := s.Units.FromMGDL(minAxisMgdl), 1.0
	for _, r := range s.GlucoseReadings {
		glucoseMax = max(glucoseMax, r.Reading)
	}
	// The glucose axis ends on whole numbers, so its labels are round
	glucoseMin, glucoseMax := math.Floor(s.Units.FromMGDL(axisStartMgdl)), math.Ceil(glucoseMax)
	for _, m := range s.Minutes {
		if m.Basal != nil {
			basalMax = max(basalMax, *m.Basal)
		}
	}

	chart := charts.NewScatter()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: opts.FuncOpts(tooltip)}),
		charts.WithGridOpts(
			// Wide enough on the right for the label of a dose or meal given just now, which is drawn right of its marker
			opts.Grid{Left: "60", Right: "60", Top: "4%", Height: "54%"},
			opts.Grid{Left: "60", Right: "60", Top: "64%", Height: "14%"},
			opts.Grid{Left: "60", Right: "60", Top: "82%", Height: "8%"},
		),
		charts.WithXAxisOpts(timeAxis(glucoseX, at(s.Since), at(now), false)),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: glucoseMin, Max: glucoseMax, SplitLine: &opts.SplitLine{Show: opts.Bool(false)}, AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false)}}),
	)
	chart.AddJSFuncStrs("%MY_ECHARTS%.setOption({useUTC: true});")
	chart.ExtendXAxis(timeAxis(treatmentX, at(s.Since), at(now), false), timeAxis(basalX, at(s.Since), at(now), true))
	chart.ExtendYAxis(
		opts.YAxis{Type: "category", Data: []string{"Carbs", "Insulin"}, GridIndex: treatmentX, AxisLine: &opts.AxisLine{Show: opts.Bool(false)}},
		opts.YAxis{Type: "category", Data: []string{"Basal"}, GridIndex: basalX, AxisLine: &opts.AxisLine{Show: opts.Bool(false)}},
		opts.YAxis{Type: "value", Min: 0, Max: basalMax, GridIndex: basalX, Show: opts.Bool(false)},
	)

	// Readings are red below the range, amber above and the brand colour in it, with its bounds as dashed lines; behind
	// them each stretch of the same targets is a shaded band
	var lows, highs, inRange []opts.ScatterData
	for _, r := range s.GlucoseReadings {
		point := opts.ScatterData{Value: []any{at(r.OccurredAt), r.Reading}}
		switch glucoseRange(s, r.Reading) {
		case "Low":
			lows = append(lows, point)
		case "High":
			highs = append(highs, point)
		default:
			inRange = append(inRange, point)
		}
	}
	targets := []charts.SeriesOpts{charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(false)}})}
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.LowTarget, m.HighTarget} }) {
		if r.values[0] != nil && r.values[1] != nil {
			targets = append(targets, charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Name: "Target", Coordinate0: []any{at(r.start), *r.values[1]}, Coordinate1: []any{at(r.end), *r.values[0]}, ItemStyle: &opts.ItemStyle{Color: blue500, Opacity: opts.Float(0.15)}}))
		}
	}
	dashed := func(colour string) charts.SeriesOpts {
		return charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Type: "dashed", Color: colour, Width: 2}})
	}
	dot := charts.WithScatterChartOpts(opts.ScatterChart{SymbolSize: 6})
	chart.AddSeries("Low", lows, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseLow}), dashed(glucoseLow), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "Low", YAxis: s.RangeLow}))
	chart.AddSeries("High", highs, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseHigh}), dashed(glucoseHigh), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "High", YAxis: s.RangeHigh}))
	chart.AddSeries("In range", inRange, append(targets, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseInRange}))...)

	marker := func(colour string) []charts.SeriesOpts {
		return []charts.SeriesOpts{
			charts.WithScatterChartOpts(opts.ScatterChart{XAxisIndex: treatmentX, YAxisIndex: treatmentY, Symbol: "triangle", SymbolSize: 10}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: colour}),
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right", Formatter: "{b}", Color: gray700}),
		}
	}
	var insulin, carbs []opts.ScatterData
	for _, d := range slices.Backward(s.InsulinDoses) {
		insulin = append(insulin, opts.ScatterData{Name: value(&d.Dose, "U"), Value: []any{at(d.OccurredAt), "Insulin"}, SymbolRotate: 180})
	}
	for _, f := range slices.Backward(s.FoodIntakes) {
		if f.Carbs != nil {
			carbs = append(carbs, opts.ScatterData{Name: value(f.Carbs, "g"), Value: []any{at(f.OccurredAt), "Carbs"}})
		}
	}
	chart.AddSeries("Insulin", insulin, marker(brandNight)...)
	chart.AddSeries("Carbs", carbs, marker(amber500)...)

	// The basal steps to each rate when it comes into force, and falls to nothing where none is in force
	var basals []opts.LineData
	var last *float64
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.Basal} }) {
		rate := r.values[0]
		switch {
		case rate != nil && last == nil:
			basals = append(basals, opts.LineData{Value: []any{at(r.start), 0}}, opts.LineData{Value: []any{at(r.start), *rate}}, opts.LineData{Value: []any{at(r.end), *rate}})
		case rate != nil:
			basals = append(basals, opts.LineData{Value: []any{at(r.start), *rate}}, opts.LineData{Value: []any{at(r.end), *rate}})
		case last != nil:
			basals = append(basals, opts.LineData{Value: []any{at(r.start), 0}})
		}
		last = rate
	}
	if last != nil {
		basals = append(basals, opts.LineData{Value: []any{at(now), 0}})
	}
	basal := charts.NewLine()
	basal.AddSeries("Basal", basals,
		charts.WithLineChartOpts(opts.LineChart{XAxisIndex: basalX, YAxisIndex: basalRate, ShowSymbol: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 2}),
		charts.WithAreaStyleOpts(opts.AreaStyle{Color: brandNight, Opacity: opts.Float(0.1)}),
	)
	chart.Overlap(basal)
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
}

// run is a stretch of minutes from start to end with the same values
type run struct {
	start, end int64
	values     []*float64
}

// runs are the summary's minutes up to now joined into stretches whose values are the same
func runs(s *sitemodels.Summary, now int64, values func(sitemodels.SummaryMinute) []*float64) []run {
	var all []run
	for i, m := range s.Minutes {
		start := s.Since + int64(i)*60_000
		if start >= now {
			break
		}
		end := min(start+60_000, now)
		v := values(m)
		if len(all) > 0 && slices.EqualFunc(all[len(all)-1].values, v, func(a, b *float64) bool { return (a == nil) == (b == nil) && (a == nil || *a == *b) }) {
			all[len(all)-1].end = end
			continue
		}
		all = append(all, run{start: start, end: end, values: v})
	}
	return all
}

// timeAxis is the time axis of a grid, from the start of the summary to now, labelled only on the bottom grid
func timeAxis(gridIndex int, since, until int64, labelled bool) opts.XAxis {
	return opts.XAxis{
		Type: "time", Min: since, Max: until, GridIndex: gridIndex,
		// The first and last labels would be the exact times the summary covers, down to the millisecond
		AxisLabel: &opts.AxisLabel{Show: opts.Bool(labelled), ShowMinLabel: opts.Bool(false), ShowMaxLabel: opts.Bool(false)},
		AxisTick:  &opts.AxisTick{Show: opts.Bool(labelled)},
		AxisLine:  &opts.AxisLine{Show: opts.Bool(labelled)},
		SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}},
	}
}

// tooltip shows what a reading or treatment is and when it was, in the location's time as the tables show it
const tooltip = `function (p) {
	var at = new Date(p.value[0]).toISOString().slice(0, 19).replace('T', ' ');
	return p.seriesName + ': ' + (p.name || p.value[1]) + '<br/>' + at;
}`
