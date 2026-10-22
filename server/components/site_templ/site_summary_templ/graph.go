package site_summary_templ

import (
	"fmt"
	"strings"
	"time"

	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// The graph is drawn in a 1000 wide viewBox: glucose on top, then a row of insulin and one of carbs, then temp basals, then the time axis
const (
	graphWidth   = 1000.0
	graphLeft    = 60.0
	graphRight   = 960.0
	glucoseTop   = 10.0
	glucoseBase  = 250.0
	insulinRow   = 268.0
	carbsRow     = 290.0
	basalTop     = 324.0
	basalBase    = 340.0
	graphHeight  = 362.0
	rangeLowMgdl = 70.0
	rangeTopMgdl = 180.0
	// The glucose axis always shows up to at least this, so a flat day is not stretched
	minAxisMgdl = 270.0
	// The glucose axis starts here, as readings below it are rare
	axisStartMgdl = 40.0
)

// graph maps times and values in a summary to points on the SVG
type graph struct {
	since, until int64
	units        *consts.Units
	glucoseMin   float64
	glucoseMax   float64
	basalMax     float64
}

func newGraph(s *sitemodels.Summary, now int64) graph {
	g := graph{since: s.Since, until: now, units: s.Units, glucoseMin: s.Units.FromMGDL(axisStartMgdl), glucoseMax: s.Units.FromMGDL(minAxisMgdl), basalMax: 1}
	for _, r := range s.GlucoseReadings {
		g.glucoseMax = max(g.glucoseMax, r.Reading)
	}
	for _, b := range s.TempBasals {
		if b.Absolute != nil {
			g.basalMax = max(g.basalMax, *b.Absolute)
		}
	}
	return g
}

func (g graph) x(ms int64) float64 {
	ms = min(max(ms, g.since), g.until)
	return graphLeft + float64(ms-g.since)/float64(g.until-g.since)*(graphRight-graphLeft)
}

// scale maps 0 to base and top value to top, keeping values outside them on the edges
func scale(v, maxV, top, base float64) float64 {
	return base - min(max(v/maxV, 0), 1)*(base-top)
}

func (g graph) glucoseY(v float64) float64 {
	return scale(v-g.glucoseMin, g.glucoseMax-g.glucoseMin, glucoseTop, glucoseBase)
}

func (g graph) basalY(v float64) float64 { return scale(v, g.basalMax, basalTop, basalBase) }

func (g graph) rangeLow() float64 { return g.units.FromMGDL(rangeLowMgdl) }
func (g graph) rangeTop() float64 { return g.units.FromMGDL(rangeTopMgdl) }

// glucoseColour is red below range, amber above and the brand colour in it
func (g graph) glucoseColour(v float64) string {
	switch {
	case v < g.rangeLow():
		return "fill-red-600"
	case v > g.rangeTop():
		return "fill-amber-500"
	}
	return "fill-brand-night"
}

// endOf is when something lasting durationMinutes from start ends, no later than the graph
func (g graph) endOf(start, durationMinutes int64) int64 {
	return min(start+durationMinutes*60_000, g.until)
}

// basalPath is a step line along the bottom for each temp basal, as high as its rate, for as long as it runs
func (g graph) basalPath(basals []*sitemodels.SiteTempBasal) string {
	var path strings.Builder
	for _, b := range basals {
		if b.Absolute == nil {
			continue
		}
		fmt.Fprintf(&path, "M%s %s V%s H%s V%s ", f(g.x(b.OccurredAt)), f(basalBase), f(g.basalY(*b.Absolute)), f(g.x(g.endOf(b.OccurredAt, b.DurationMinutes))), f(basalBase))
	}
	return strings.TrimSpace(path.String())
}

// triangle is a small marker centred on x and y, pointing up or down
func triangle(x, y float64, up bool) string {
	tip, side := y-5, y+4
	if !up {
		tip, side = y+5, y-4
	}
	return fmt.Sprintf("M%s %s L%s %s L%s %s Z", f(x), f(tip), f(x+5), f(side), f(x-5), f(side))
}

type tick struct {
	X     float64
	Label string
}

// ticks are the time axis labels, 6 to 8 across whatever the graph covers
func (g graph) ticks() []tick {
	step := time.Hour
	for _, s := range []time.Duration{2 * time.Hour, 4 * time.Hour, 6 * time.Hour, 12 * time.Hour, 24 * time.Hour} {
		if time.Duration(g.until-g.since)*time.Millisecond/step <= 8 {
			break
		}
		step = s
	}
	layout := "15:04"
	if step >= 12*time.Hour {
		layout = "Mon 15:04"
	}
	// Ticks start from local midnight, so they land on round hours in any time zone
	start := time.UnixMilli(g.since)
	t := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	for t.UnixMilli() <= g.since {
		t = t.Add(step)
	}
	ticks := []tick{}
	for ; t.UnixMilli() < g.until; t = t.Add(step) {
		ticks = append(ticks, tick{g.x(t.UnixMilli()), t.Format(layout)})
	}
	return ticks
}

// f is a coordinate to one decimal place, plenty for a 1000 wide SVG
func f(v float64) string {
	return fmt.Sprintf("%.1f", v)
}
