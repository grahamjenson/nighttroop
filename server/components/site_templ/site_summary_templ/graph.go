package site_summary_templ

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/css"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/display"
)

const (
	// The glucose axis is from 0 to 16 mmol/L (288 mg/dL) whatever the readings, so every day is drawn alike: a
	// reading above it is drawn past its top, up to overflowShare of the axis more, rather than squashing the rest
	glucoseTopMgdl, overflowShare = 288.0, 0.12
	// The basal's highest rate is drawn at 3 mmol/L (54 mg/dL) up the glucose axis
	basalTopMgdl = 54.0
	// Everything is drawn on one grid, as Carer draws it: glucose on the left axis, the basal on its own axis along
	// the bottom, and the treatments on a hidden axis from 0 at the bottom to 1 at the top
	glucoseAxis, basalAxis, treatmentAxis = 0, 1, 2
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
	blue500        = css.Colour("blue-500")
	blue700        = css.Colour("blue-700")
	gray200        = css.Colour("gray-200")
	gray700        = css.Colour("gray-700")
)

// pointingDown and pointingUp are the legend's triangles, insulin's and carbs', the same size as each other
const (
	pointingDown = "path://M0,0L2,0L1,2Z"
	pointingUp   = "path://M1,0L2,2L0,2Z"
)

// legend is the graph's key, each drawn as it is on the graph: a legend draws its series' symbol unturned, so insulin's
// triangles, turned to point down, would point up, and the basal, a line without symbols over a shaded area, would be
// a line through a circle
var legend = []map[string]any{
	{"name": "Insulin", "icon": pointingDown},
	{"name": "SMB", "icon": pointingDown},
	{"name": "Carbs", "icon": pointingUp},
	{"name": "Basal", "icon": "rect", "itemStyle": map[string]any{"color": "transparent", "borderColor": brandNight, "borderWidth": 2}},
	{"name": "Target", "icon": "path://M0,4L12,4L12,6L0,6Z", "itemStyle": map[string]any{"color": blue700}},
}

// isSMB is whether a dose is a loop's small automatic bolus, which is drawn smaller and fainter than one given by hand:
// Trio's and iAPS's are sent as SMB, and Loop's as a bolus that is automatic
func isSMB(d *sitemodels.SiteInsulinDose) bool {
	return (d.DoseType != nil && strings.EqualFold(*d.DoseType, "SMB")) || (d.Automatic != nil && *d.Automatic)
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
	at := inLocation(location)
	// The glucose axis ends on a whole number, 16 or 288, so its labels are round
	glucoseTop, basalMax := math.Round(s.Units.FromMGDL(glucoseTopMgdl)), highestBasal(s)
	chart := newGraphChart(s, now, at, glucoseTop, basalMax)

	// Readings are red below the range, amber above and the brand colour in it, with its bounds as dashed lines; behind
	// them the urgent range and the range are shaded across the graph, and each stretch of the same targets over them
	lows, highs, inRange, glucoseLine := glucosePoints(s, at, glucoseTop)
	bands, targetLines := targets(s, now, at)
	dot := charts.WithScatterChartOpts(opts.ScatterChart{SymbolSize: 6})
	chart.AddSeries("Low", lows, seriesID("low"), dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseLow}), dashed(glucoseLow), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "Low", YAxis: s.RangeLow}))
	chart.AddSeries("High", highs, seriesID("high"), dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseHigh}), dashed(glucoseHigh), charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{Name: "High", YAxis: s.RangeHigh}))
	chart.AddSeries("In range", inRange, append(bands, seriesID("in-range"), dot, charts.WithItemStyleOpts(opts.ItemStyle{Color: glucoseInRange}))...)

	insulin, smbs, carbs := treatmentPoints(s, at)
	chart.AddSeries("SMB", smbs, marker(brandNight, 7, true)...)
	chart.AddSeries("Insulin", insulin, marker(brandNight, 10, false)...)
	chart.AddSeries("Carbs", carbs, marker(amber500, 10, false)...)

	lines := charts.NewLine()
	lines.AddSeries("Target", targetLines,
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false), ConnectNulls: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: blue700, Width: 1.5}),
		charts.WithItemStyleOpts(opts.ItemStyle{Color: blue700}),
	)
	lines.AddSeries("Glucose", glucoseLine, seriesID("glucose"),
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false), ConnectNulls: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Color: brandNight, Width: 1, Opacity: opts.Float(0.3)}),
	)
	basals := basalSteps(s, now, at)
	lines.AddSeries("Basal", basals, basalOpts(basals, basalMax)...)
	chart.Overlap(lines)
	chart.AddJSFuncs("window.addEventListener('resize', function() { %MY_ECHARTS%.resize(); });")
	return chart
}

// localTime is a time in milliseconds moved by its location's offset from UTC then, so drawn in UTC it shows the
// location's time
type localTime func(ms int64) int64

// inLocation is each time as location's
func inLocation(location *time.Location) localTime {
	return func(ms int64) int64 {
		_, offset := time.UnixMilli(ms).In(location).Zone()
		return ms + int64(offset)*1000
	}
}

// highestBasal is the summary's highest basal rate, at least 1 U/hr, which the basal's axis is scaled to
func highestBasal(s *sitemodels.Summary) float64 {
	basalMax := 1.0
	for _, m := range s.Minutes {
		if m.Basal != nil {
			basalMax = max(basalMax, *m.Basal)
		}
	}
	return basalMax
}

// newGraphChart is the graph's chart with its axes, tooltip, grid and legend, from the summary's start to now, and
// no series yet
func newGraphChart(s *sitemodels.Summary, now int64, at localTime, glucoseTop, basalMax float64) *charts.Scatter {
	basalTop := math.Round(s.Units.FromMGDL(basalTopMgdl))
	since, until := at(s.Since), at(now)

	chart := charts.NewScatter()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Renderer: "svg"}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "0", Data: legend}),
		// Pointing at a time shows the glucose, basal, target and treatments then
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis", AxisPointer: &opts.AxisPointer{Type: "line", Snap: opts.Bool(false)},
			Formatter: opts.FuncOpts(hoverTooltip(s, now, at))}),
		// Wide enough on the right for the label of a dose or meal given just now, which is drawn right of its marker,
		// and the basal's highest rate; tall enough at the top for readings drawn past the axis, and at the bottom for
		// the times and the legend
		charts.WithGridOpts(opts.Grid{Left: "60", Right: "70", Top: "40", Bottom: "56"}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time", Min: since, Max: until,
			// The first and last labels would be the exact times the summary covers, down to the millisecond
			AxisLabel: &opts.AxisLabel{ShowMinLabel: opts.Bool(false), ShowMaxLabel: opts.Bool(false)},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Type: "dashed", Color: gray200}},
		}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Min: 0, Max: glucoseTop, SplitLine: &opts.SplitLine{Show: opts.Bool(false)}}),
	)
	// go-echarts has no series' clip or axis interval, so here the glucose is let past the axis's top, and the axis is
	// marked in quarters, 0, 4, 8, 12 and 16 mmol/L, so its top is marked and none is crowded beside it
	chart.AddJSFuncStrs(types.FuncStr(fmt.Sprintf("%%MY_ECHARTS%%.setOption({useUTC: true, yAxis: [{interval: %g}], series: [{id: 'low', clip: false}, {id: 'high', clip: false}, {id: 'in-range', clip: false}, {id: 'glucose', clip: false}]});", glucoseTop/4)))
	chart.ExtendYAxis(
		// The basal's highest rate reaches basalTop up the glucose axis
		opts.YAxis{Type: "value", Min: 0, Max: basalMax * glucoseTop / basalTop, Show: opts.Bool(false)},
		opts.YAxis{Type: "value", Min: 0, Max: 1, Show: opts.Bool(false)},
	)
	return chart
}

// glucosePoints are the readings as dots below, above and in the range, and the line joining them, broken where they
// are more than glucoseGapMs apart. A reading above the axis is drawn past its top, but no further than overflowShare
// of it, so it stays on the page.
func glucosePoints(s *sitemodels.Summary, at localTime, glucoseTop float64) (lows, highs, inRange []opts.ScatterData, line []opts.LineData) {
	drawn := func(reading float64) float64 { return min(reading, glucoseTop*(1+overflowShare)) }
	for i, r := range s.GlucoseReadings {
		point := opts.ScatterData{Value: []any{at(r.OccurredAt), drawn(r.Reading)}}
		switch display.GlucoseRange(r.Reading, s.RangeLow, s.RangeHigh) {
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
		line = append(line, opts.LineData{Value: []any{at(r.OccurredAt), drawn(r.Reading)}})
	}
	return lows, highs, inRange, line
}

// targets are the bands shaded behind the readings, the urgent range and the range across the graph and over them the
// targets in force, with the lines of the low and high target. The targets in force are a light band, edged by darker
// lines of the low and high target, each a step as it changes, broken where there is none.
func targets(s *sitemodels.Summary, now int64, at localTime) (bands []charts.SeriesOpts, targetLines []opts.LineData) {
	since, until := at(s.Since), at(now)
	bands = []charts.SeriesOpts{
		charts.WithMarkAreaStyleOpts(opts.MarkAreaStyle{Label: &opts.Label{Show: opts.Bool(false)}}),
		band(since, until, s.RangeUrgentHigh, s.RangeUrgentLow, brandNight, 0.02),
		band(since, until, s.RangeHigh, s.RangeLow, brandNight, 0.04),
	}
	var highLine []opts.LineData
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.LowTarget, m.HighTarget} }) {
		if r.values[0] == nil || r.values[1] == nil {
			targetLines, highLine = append(targetLines, opts.LineData{Value: "-"}), append(highLine, opts.LineData{Value: "-"})
			continue
		}
		bands = append(bands, band(at(r.start), at(r.end), *r.values[1], *r.values[0], blue500, 0.1))
		targetLines = append(targetLines, opts.LineData{Value: []any{at(r.start), *r.values[0]}}, opts.LineData{Value: []any{at(r.end), *r.values[0]}})
		highLine = append(highLine, opts.LineData{Value: []any{at(r.start), *r.values[1]}}, opts.LineData{Value: []any{at(r.end), *r.values[1]}})
	}
	return bands, append(append(targetLines, opts.LineData{Value: "-"}), highLine...)
}

// band is a shaded area behind the readings from start to end and from high to low
func band(start, end int64, high, low float64, colour string, opacity float32) charts.SeriesOpts {
	return charts.WithMarkAreaNameCoordItemOpts(opts.MarkAreaNameCoordItem{Coordinate0: []any{start, high}, Coordinate1: []any{end, low}, ItemStyle: &opts.ItemStyle{Color: colour, Opacity: opts.Float(opacity)}})
}

// dashed draws a series' mark lines as thin dashed lines of colour, without labels
func dashed(colour string) charts.SeriesOpts {
	return charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Symbol: []string{"none", "none"}, Label: &opts.Label{Show: opts.Bool(false)}, LineStyle: &opts.LineStyle{Type: "dashed", Color: colour, Width: 1}})
}

// seriesID gives a series the ID the chart's script sets its options by
func seriesID(id string) charts.SeriesOpts {
	return charts.WithSeriesOpts(func(s *charts.SingleSeries) { s.Id = id })
}

// marker is a row of treatments on the treatment axis, labelled with what they were unless faint, as SMBs are
// given every few minutes
func marker(colour string, size int, faint bool) []charts.SeriesOpts {
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

// treatmentPoints are the insulin doses given by hand, a loop's SMBs and the carbs, each labelled with how much
func treatmentPoints(s *sitemodels.Summary, at localTime) (insulin, smbs, carbs []opts.ScatterData) {
	for _, d := range s.InsulinDoses {
		point := opts.ScatterData{Name: display.Value(&d.Dose, "U"), Value: []any{at(d.OccurredAt), insulinHeight}, SymbolRotate: 180}
		if isSMB(d) {
			smbs = append(smbs, point)
		} else {
			insulin = append(insulin, point)
		}
	}
	for _, f := range s.FoodIntakes {
		if f.Carbs != nil {
			carbs = append(carbs, opts.ScatterData{Name: display.Value(f.Carbs, "g"), Value: []any{at(f.OccurredAt), carbsHeight}})
		}
	}
	return insulin, smbs, carbs
}

// basalSteps is the basal's line: it steps to each rate when it comes into force, and falls to nothing where none is
// in force
func basalSteps(s *sitemodels.Summary, now int64, at localTime) []opts.LineData {
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
	return basals
}

// basalOpts draw the basal along the bottom on its own axis, shaded beneath. Its highest rate, if any is in force, is
// labelled on the right, where Carer labels it, by a mark line not drawn so it is not read as glucose.
func basalOpts(basals []opts.LineData, basalMax float64) []charts.SeriesOpts {
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
	return basal
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

// hover is what the tooltip looks up for the time pointed at, in the location's time as the graph draws it: each
// reading, each stretch of a basal and of targets, and each treatment, with their values written as the tables write
// them
type hover struct {
	Units      string  `json:"units"`
	Readings   [][]any `json:"readings"`
	Basal      [][]any `json:"basal"`
	Target     [][]any `json:"target"`
	Treatments [][]any `json:"treatments"`
}

// hoverTooltip is the tooltip shown pointing at a time: the time, the reading nearest it within 5 minutes and when it
// was, the basal and the targets in force then, and any treatment within 5 minutes
func hoverTooltip(s *sitemodels.Summary, now int64, at localTime) string {
	h := hover{Units: display.Units(s.Units), Readings: [][]any{}, Basal: [][]any{}, Target: [][]any{}, Treatments: [][]any{}}
	for _, r := range s.GlucoseReadings {
		h.Readings = append(h.Readings, []any{at(r.OccurredAt), display.Glucose(r.Reading, s.Units)})
	}
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.Basal} }) {
		if r.values[0] != nil {
			h.Basal = append(h.Basal, []any{at(r.start), at(r.end), display.Value(r.values[0], "")})
		}
	}
	for _, r := range runs(s, now, func(m sitemodels.SummaryMinute) []*float64 { return []*float64{m.LowTarget, m.HighTarget} }) {
		if r.values[0] != nil && r.values[1] != nil {
			h.Target = append(h.Target, []any{at(r.start), at(r.end), display.Glucose(*r.values[0], s.Units) + "–" + display.Glucose(*r.values[1], s.Units)})
		}
	}
	for _, d := range s.InsulinDoses {
		what := "Insulin"
		if isSMB(d) {
			what = "SMB"
		}
		h.Treatments = append(h.Treatments, []any{at(d.OccurredAt), what + ": " + display.Value(&d.Dose, "U")})
	}
	for _, f := range s.FoodIntakes {
		if f.Carbs != nil {
			h.Treatments = append(h.Treatments, []any{at(f.OccurredAt), "Carbs: " + display.Value(f.Carbs, "g")})
		}
	}
	data, _ := json.Marshal(h)
	// go-echarts writes a function into the page as it is, but leaves a double quote in it escaped, so the data is
	// written with single quotes; its strings are numbers, units and labels written here, with no quote of either kind
	return fmt.Sprintf(hoverScript, strings.ReplaceAll(string(data), `"`, `'`))
}

// hoverScript makes the tooltip's formatter, with the data it looks up made once, not on each move of the pointer
const hoverScript = `(function () {
	var d = %s, near = 5 * 60000;
	function time(t) { return new Date(t).toISOString().slice(11, 16); }
	function during(runs, t) {
		for (var i = 0; i < runs.length; i++) {
			if (runs[i][0] <= t && t < runs[i][1]) return runs[i][2];
		}
		return null;
	}
	return function (p) {
		var t = (Array.isArray(p) ? p[0] : p).axisValue;
		var lines = ['<b>' + new Date(t).toISOString().slice(0, 16).replace('T', ' ') + '</b>'];
		var reading = null;
		d.readings.forEach(function (r) {
			if (Math.abs(r[0] - t) <= near && (reading === null || Math.abs(r[0] - t) < Math.abs(reading[0] - t))) reading = r;
		});
		lines.push(reading === null ? 'Glucose: no reading' : 'Glucose: ' + reading[1] + ' ' + d.units + ' at ' + time(reading[0]));
		var basal = during(d.basal, t), target = during(d.target, t);
		if (basal !== null) lines.push('Basal: ' + basal + ' U/hr');
		if (target !== null) lines.push('Target: ' + target + ' ' + d.units);
		d.treatments.forEach(function (x) {
			if (Math.abs(x[0] - t) <= near) lines.push(x[1] + ' at ' + time(x[0]));
		});
		return lines.join('<br/>');
	};
})()`
