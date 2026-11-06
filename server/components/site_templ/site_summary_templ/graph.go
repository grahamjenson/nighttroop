package site_summary_templ

import (
	"math"
	"slices"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/css"
	"nighttroop.t1d.nz/utils/consts"
)

const (
	// The glucose axis always shows up to at least this, so a flat day is not stretched
	minAxisMgdl = 270.0
	// The glucose axis starts here, as readings below it are rare
	axisStartMgdl = 40.0
	// Everything is drawn on one grid, as Carer draws it: glucose on the left axis, the basal on its own axis along
	// the bottom, and the treatments on a hidden axis from 0 at the bottom to 1 at the top, in the room left above the
	// highest reading
	glucoseAxis, basalAxis, treatmentAxis = 0, 1, 2
	// The basal's highest rate is drawn a fifth of the way up, and a fifth again of the glucose axis is room above the
	// highest reading for the treatments
	basalShare, treatmentRoom = 5.0, 0.2
	// Carbs are drawn above insulin, near the top of the treatment axis
	carbsHeight, insulinHeight = 0.95, 0.87
	// Readings further apart than this are not joined by the glucose line
	glucoseGapMs = 15 * 60_000
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

// isSMB is whether a dose is a loop's small automatic bolus, which is drawn smaller and fainter than one given by hand
func isSMB(d *sitemodels.SiteInsulinDose) bool {
	return d.DoseType != nil && strings.EqualFold(*d.DoseType, "SMB")
}

// graph is the summary as a go-echarts chart on one grid, as Carer draws it: the urgent range and the range shaded
// across it, the targets in force each minute as bands over those, the glucose as dots joined by a thin line, the basal
// in force each minute along the bottom, and above the readings carbs as triangles pointing up and insulin as
// triangles pointing down, a loop's SMBs smaller and fainter. Treatments are added newest first, as the searches give
// them: where labels of treatments close together overlap, ECharts keeps the first, so the latest can be read.
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
	glucoseTop := math.Ceil(glucoseMax + treatmentRoom*(glucoseMax-glucoseMin))
	for _, m := range s.Minutes {
		if m.Basal != nil {
			basalMax = max(basalMax, *m.Basal)
		}
	}
	since, until := at(s.Since), at(now)

	chart := charts.NewScatter()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "0", Data: []string{"Insulin", "SMB", "Carbs", "Basal"}}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: opts.FuncOpts(tooltip)}),
		// Wide enough on the right for the label of a dose or meal given just now, which is drawn right of its marker,
		// and the basal's highest rate; tall enough at the bottom for the times and the legend
		charts.WithGridOpts(opts.Grid{Left: "60", Right: "70", Top: "10", Bottom: "56"}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time", Min: since, Max: until,
			// The first and last labels would be the exact times the summary covers, down to the millisecond
			AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false), ShowMaxLabel: opts.Bool(false)},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}},
		}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: glucoseMin, Max: glucoseTop, SplitLine: &opts.SplitLine{Show: opts.Bool(false)}, AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false), ShowMaxLabel: opts.Bool(false)}}),
	)
	chart.AddJSFuncStrs("%MY_ECHARTS%.setOption({useUTC: true});")
	chart.ExtendYAxis(
		opts.YAxis{Type: "value", Min: 0, Max: basalMax * basalShare, Show: opts.Bool(false)},
		opts.YAxis{Type: "value", Min: 0, Max: 1, Show: opts.Bool(false)},
	)

	// band is a shaded area behind the readings from start to end and from high to low
	band := func(start, end int64, high, low float64, colour string, opacity float32) charts.SeriesOpts {
		return charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Coordinate0: []any{start, high}, Coordinate1: []any{end, low}, ItemStyle: &opts.ItemStyle{Color: colour, Opacity: opts.Float(opacity)}})
	}
	// Readings are red below the range, amber above and the brand colour in it, with its bounds as dashed lines; behind
	// them the urgent range and the range are shaded across the graph, and each stretch of the same targets over them
	var lows, highs, inRange []opts.ScatterData
	var line []opts.LineData
	for i, r := range s.GlucoseReadings {
		point := opts.ScatterData{Value: []any{at(r.OccurredAt), r.Reading}}
		switch glucoseRange(s, r.Reading) {
		case "Low":
			lows = append(lows, point)
		case "High":
			highs = append(highs, point)
		default:
			inRange = append(inRange, point)
		}
		if i > 0 && s.GlucoseReadings[i-1].OccurredAt-r.OccurredAt > glucoseGapMs {
			line = append(line, opts.LineData{Value: "-"})
		}
		line = append(line, opts.LineData{Value: []any{at(r.OccurredAt), r.Reading}})
	}
	targets := []charts.SeriesOpts{
		charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(false)}}),
		band(since, until, s.RangeUrgentHigh, s.RangeUrgentLow, brandNight, 0.02),
		band(since, until, s.RangeHigh, s.RangeLow, brandNight, 0.04),
	}
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.LowTarget, m.HighTarget} }) {
		if r.values[0] != nil && r.values[1] != nil {
			targets = append(targets, band(at(r.start), at(r.end), *r.values[1], *r.values[0], blue500, 0.2))
		}
	}
	dashed := func(colour string) charts.SeriesOpts {
		return charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Type: "dashed", Color: colour, Width: 1}})
	}
	dot := charts.WithScatterChartOpts(opts.ScatterChart{SymbolSize: 6})
	chart.AddSeries("Low", lows, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseLow}), dashed(glucoseLow), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "Low", YAxis: s.RangeLow}))
	chart.AddSeries("High", highs, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseHigh}), dashed(glucoseHigh), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "High", YAxis: s.RangeHigh}))
	chart.AddSeries("In range", inRange, append(targets, dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseInRange}))...)

	// marker is a row of treatments on the treatment axis, labelled with what they were unless faint, as SMBs are
	// given every few minutes
	marker := func(colour string, size int, faint bool) []charts.SeriesOpts {
		opacity := float32(1)
		if faint {
			opacity = 0.5
		}
		return []charts.SeriesOpts{
			charts.WithScatterChartOpts(opts.ScatterChart{YAxisIndex: treatmentAxis, Symbol: "triangle", SymbolSize: size}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: colour, Opacity: opts.Float(opacity)}),
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(!faint), Position: "right", Formatter: "{b}", Color: gray700}),
			// Treatments a moment apart would have their labels written over each other
			charts.WithLabelLayoutOpts(opts.LabelLayout{HideOverlap: opts.Bool(true)}),
		}
	}
	var insulin, smbs, carbs []opts.ScatterData
	for _, d := range s.InsulinDoses {
		point := opts.ScatterData{Name: value(&d.Dose, "U"), Value: []any{at(d.OccurredAt), insulinHeight}, SymbolRotate: 180}
		if isSMB(d) {
			smbs = append(smbs, point)
		} else {
			insulin = append(insulin, point)
		}
	}
	for _, f := range s.FoodIntakes {
		if f.Carbs != nil {
			carbs = append(carbs, opts.ScatterData{Name: value(f.Carbs, "g"), Value: []any{at(f.OccurredAt), carbsHeight}})
		}
	}
	chart.AddSeries("SMB", smbs, marker(brandNight, 7, true)...)
	chart.AddSeries("Insulin", insulin, marker(brandNight, 10, false)...)
	chart.AddSeries("Carbs", carbs, marker(amber500, 10, false)...)

	// The basal steps to each rate when it comes into force, and falls to nothing where none is in force; its highest
	// rate, if any is in force, is labelled on the right, where Carer labels it, by a mark line not drawn so it is not
	// read as glucose
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
		basals = append(basals, opts.LineData{Value: []any{until, 0}})
	}
	lines := charts.NewLine()
	lines.AddSeries("Glucose", line,
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false), ConnectNulls: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 1, Opacity: opts.Float(0.3)}),
	)
	basal := []charts.SeriesOpts{
		charts.WithLineChartOpts(opts.LineChart{YAxisIndex: basalAxis, ShowSymbol: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 2}),
		charts.WithItemStyleOpts(opts.ItemStyle{Color: brandNight}),
		charts.WithAreaStyleOpts(opts.AreaStyle{Color: brandNight, Opacity: opts.Float(0.1)}),
	}
	if len(basals) > 0 {
		basal = append(basal,
			charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(true), Position: "end", Formatter: "{c} U/hr", Color: gray700}, LineStyle: &opts.LineStyle{Color: "transparent"}}),
			charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "Highest basal", YAxis: consts.Round(basalMax, 2)}),
		)
	}
	lines.AddSeries("Basal", basals, basal...)
	chart.Overlap(lines)
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

// tooltip shows what a reading or treatment is and when it was, in the location's time as the tables show it
const tooltip = `function (p) {
	var at = new Date(p.value[0]).toISOString().slice(0, 19).replace('T', ' ');
	return p.seriesName + ': ' + (p.name || p.value[1]) + '<br/>' + at;
}`
