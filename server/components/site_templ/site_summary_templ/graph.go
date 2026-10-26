package site_summary_templ

import (
	"math"
	"slices"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
)

const (
	rangeLowMgdl = 70.0
	rangeTopMgdl = 180.0
	// The glucose axis always shows up to at least this, so a flat day is not stretched
	minAxisMgdl = 270.0
	// The glucose axis starts here, as readings below it are rare
	axisStartMgdl = 40.0
	// ECharts draws with these, as it cannot use the Tailwind classes; brand night is `primary` in css/input.css
	brandNight = "#10203F"
	red600     = "#dc2626"
	amber500   = "#f59e0b"
	blue500    = "#3b82f6"
	gray100    = "#f3f4f6"
	gray200    = "#e5e7eb"
	gray700    = "#374151"
	// The axes of the three grids: glucose on top, then a row each of carbs and insulin, then temp basals
	glucoseX               = 0
	treatmentX, treatmentY = 1, 1
	basalX, basalRate      = 2, 3
)

// graph is the summary as a go-echarts chart on one time axis: glucose dots over the target range with
// temp targets as shaded lines, insulin as triangles pointing down and carbs as triangles pointing up
// under it, and temp basals as a line at the bottom. Treatments are added oldest first, so the latest
// is drawn on top where markers close together overlap.
func graph(s *sitemodels.Summary, now int64) *charts.Scatter {
	glucoseMax, basalMax := s.Units.FromMGDL(minAxisMgdl), 1.0
	for _, r := range s.GlucoseReadings {
		glucoseMax = max(glucoseMax, r.Reading)
	}
	// The glucose axis ends on whole numbers, so its labels are round
	glucoseMin, glucoseMax := math.Floor(s.Units.FromMGDL(axisStartMgdl)), math.Ceil(glucoseMax)
	for _, b := range s.TempBasals {
		if b.Absolute != nil {
			basalMax = max(basalMax, *b.Absolute)
		}
	}
	low, top := s.Units.FromMGDL(rangeLowMgdl), s.Units.FromMGDL(rangeTopMgdl)
	endOf := func(start, durationMinutes int64) int64 { return min(start+durationMinutes*60_000, now) }

	chart := charts.NewScatter()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: opts.FuncOpts(tooltip)}),
		charts.WithGridOpts(
			opts.Grid{Left: "60", Right: "30", Top: "4%", Height: "54%"},
			opts.Grid{Left: "60", Right: "30", Top: "64%", Height: "14%"},
			opts.Grid{Left: "60", Right: "30", Top: "82%", Height: "8%"},
		),
		charts.WithXAxisOpts(timeAxis(glucoseX, s.Since, now, false)),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: glucoseMin, Max: glucoseMax, SplitLine: &opts.SplitLine{Show: opts.Bool(false)}, AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false)}}),
	)
	chart.ExtendXAxis(timeAxis(treatmentX, s.Since, now, false), timeAxis(basalX, s.Since, now, true))
	chart.ExtendYAxis(
		opts.YAxis{Type: "category", Data: []string{"Carbs", "Insulin"}, GridIndex: treatmentX, AxisLine: &opts.AxisLine{Show: opts.Bool(false)}},
		opts.YAxis{Type: "category", Data: []string{"Basal"}, GridIndex: basalX, AxisLine: &opts.AxisLine{Show: opts.Bool(false)}},
		opts.YAxis{Type: "value", Min: 0, Max: basalMax, GridIndex: basalX, Show: opts.Bool(false)},
	)

	// Readings are red below the target range, amber above and the brand colour in it; the range is a band
	// behind them, with its bounds as dashed lines, and each temp target is a blue line over its shaded range
	var lows, highs, inRange []opts.ScatterData
	for _, r := range s.GlucoseReadings {
		point := opts.ScatterData{Value: []any{r.OccurredAt, r.Reading}}
		switch {
		case r.Reading < low:
			lows = append(lows, point)
		case r.Reading > top:
			highs = append(highs, point)
		default:
			inRange = append(inRange, point)
		}
	}
	targets := []charts.SeriesOpts{charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(false)}}), charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Name: "Target range", Coordinate0: []any{s.Since, top}, Coordinate1: []any{now, low}, ItemStyle: &opts.ItemStyle{Color: gray100}})}
	for _, t := range s.TempTargets {
		targets = append(targets,
			charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Name: "Temp target", Coordinate0: []any{t.OccurredAt, t.HighTarget}, Coordinate1: []any{endOf(t.OccurredAt, t.DurationMinutes), t.LowTarget}, ItemStyle: &opts.ItemStyle{Color: blue500, Opacity: opts.Float(0.15)}}),
			charts.WithMarkLineNameCoordItemOpts(opts.MarkLineNameCoordItem{Name: "Temp target", Coordinate0: []any{t.OccurredAt, (t.LowTarget + t.HighTarget) / 2}, Coordinate1: []any{endOf(t.OccurredAt, t.DurationMinutes), (t.LowTarget + t.HighTarget) / 2}}),
		)
	}
	dashed := func(colour string) charts.SeriesOpts {
		return charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Type: "dashed", Color: colour, Width: 2}})
	}
	dot := charts.WithScatterChartOpts(opts.ScatterChart{SymbolSize: 6})
	chart.AddSeries("Low", lows, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: red600}), dashed(red600), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "Low", YAxis: low}))
	chart.AddSeries("High", highs, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: amber500}), dashed(amber500), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "High", YAxis: top}))
	chart.AddSeries("In range", inRange, append(targets, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: brandNight}), charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Color: blue500, Width: 2}}))...)

	marker := func(colour string) []charts.SeriesOpts {
		return []charts.SeriesOpts{
			charts.WithScatterChartOpts(opts.ScatterChart{XAxisIndex: treatmentX, YAxisIndex: treatmentY, Symbol: "triangle", SymbolSize: 10}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: colour}),
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right", Formatter: "{b}", Color: gray700}),
		}
	}
	var insulin, carbs []opts.ScatterData
	for _, d := range slices.Backward(s.InsulinDoses) {
		insulin = append(insulin, opts.ScatterData{Name: value(&d.Dose, "U"), Value: []any{d.OccurredAt, "Insulin"}, SymbolRotate: 180})
	}
	for _, f := range slices.Backward(s.FoodIntakes) {
		if f.Carbs != nil {
			carbs = append(carbs, opts.ScatterData{Name: value(f.Carbs, "g"), Value: []any{f.OccurredAt, "Carbs"}})
		}
	}
	chart.AddSeries("Insulin", insulin, marker(brandNight)...)
	chart.AddSeries("Carbs", carbs, marker(amber500)...)

	// Each temp basal rises to its rate when it starts and falls back when it ends
	var basals []opts.LineData
	for _, b := range slices.Backward(s.TempBasals) {
		if b.Absolute == nil {
			continue
		}
		end := endOf(b.OccurredAt, b.DurationMinutes)
		basals = append(basals, opts.LineData{Value: []any{b.OccurredAt, 0}}, opts.LineData{Value: []any{b.OccurredAt, *b.Absolute}}, opts.LineData{Value: []any{end, *b.Absolute}}, opts.LineData{Value: []any{end, 0}})
	}
	basal := charts.NewLine()
	basal.AddSeries("Temp basal", basals,
		charts.WithLineChartOpts(opts.LineChart{XAxisIndex: basalX, YAxisIndex: basalRate, ShowSymbol: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 2}),
		charts.WithAreaStyleOpts(opts.AreaStyle{Color: brandNight, Opacity: opts.Float(0.1)}),
	)
	chart.Overlap(basal)
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
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

// tooltip shows what a reading or treatment is and when it was, in the browser's time
const tooltip = `function (p) {
	var at = new Date(p.value[0]).toLocaleString();
	return p.seriesName + ': ' + (p.name || p.value[1]) + '<br/>' + at;
}`
