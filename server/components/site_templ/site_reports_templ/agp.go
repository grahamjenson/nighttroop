package site_reports_templ

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/display"
	"nighttroop.t1d.nz/utils/httputils"
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

// sum is two percents added, or nil when neither is set
func sum(a, b *float64) *float64 {
	if a == nil && b == nil {
		return nil
	}
	total := 0.0
	for _, v := range []*float64{a, b} {
		if v != nil {
			total += *v
		}
	}
	return &total
}

// newAGPGraph is the graph the AGP is drawn on: up to 350 mg/dL, or the highest 95th percentile or glucose above it
func newAGPGraph(p *sitemodels.SiteSummaryPeriod) graph {
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
	g := newGraph(960, 380, p.Units, top)
	g.Max = max(g.Max, p.Units.FromMGDL(350))
	return g
}

// agpX is how far across the graph an AGP or glucose slot, every 5 minutes from midnight, is drawn
func agpX(g graph, slot int) float64 {
	return g.X(float64(int64(slot) * sitemodels.AGPSlotMs))
}

// runs are the runs of slots in a row each of which has a value, as [first, last] pairs
func runs(n int, has func(int) bool) [][2]int {
	var all [][2]int
	for slot := 0; slot < n; slot++ {
		if !has(slot) {
			continue
		}
		start := slot
		for slot+1 < n && has(slot+1) {
			slot++
		}
		all = append(all, [2]int{start, slot})
	}
	return all
}

// bandPath is the SVG path of the area between the AGP's lower and upper percentiles, as their indexes, broken where it
// has none
func bandPath(g graph, agp [][]float64, lower, upper int) string {
	var path strings.Builder
	for _, run := range runs(len(agp), func(slot int) bool { return len(agp[slot]) == len(sitemodels.AGPPercentiles) }) {
		for slot := run[0]; slot <= run[1]; slot++ {
			fmt.Fprintf(&path, "%s%s,%s ", map[bool]string{true: "M", false: "L"}[slot == run[0]], f(agpX(g, slot)), f(g.Y(agp[slot][upper])))
		}
		for slot := run[1]; slot >= run[0]; slot-- {
			fmt.Fprintf(&path, "L%s,%s ", f(agpX(g, slot)), f(g.Y(agp[slot][lower])))
		}
		path.WriteString("Z ")
	}
	return strings.TrimSpace(path.String())
}

// linePath is the SVG path of a line through values, one each 5 minutes from midnight, broken where it has none
func linePath(g graph, values []*float64) string {
	var path strings.Builder
	for _, run := range runs(len(values), func(slot int) bool { return values[slot] != nil }) {
		for slot := run[0]; slot <= run[1]; slot++ {
			fmt.Fprintf(&path, "%s%s,%s ", map[bool]string{true: "M", false: "L"}[slot == run[0]], f(agpX(g, slot)), f(g.Y(*values[slot])))
		}
	}
	return strings.TrimSpace(path.String())
}

// percentile is the AGP's ith percentile each 5 minutes, nil where it has none
func percentile(agp [][]float64, i int) []*float64 {
	values := make([]*float64, len(agp))
	for slot, percentiles := range agp {
		if len(percentiles) == len(sitemodels.AGPPercentiles) {
			values[slot] = &percentiles[i]
		}
	}
	return values
}

// agpPoints are what the AGP's tooltip shows at each 5 minutes with percentiles: how far across the graph it is, its
// time as the settings show times, how far down each percentile is, and each percentile's glucose in the units
func agpPoints(ctx context.Context, g graph, agp [][]float64) [][]any {
	round := func(v float64) float64 { return math.Round(v*10) / 10 }
	points := [][]any{}
	for slot, percentiles := range agp {
		if len(percentiles) != len(sitemodels.AGPPercentiles) {
			continue
		}
		ys, values := []float64{}, []string{}
		for _, value := range percentiles {
			ys, values = append(ys, round(g.Y(value))), append(values, display.Glucose(ctx, value, g.Units))
		}
		points = append(points, []any{round(agpX(g, slot)), timeOfDay(ctx, slot*int(sitemodels.AGPSlotMs/60_000)), ys, values})
	}
	return points
}

// lastPercentiles is the AGP's last 5 minutes with percentiles and their percentiles, where its line labels go, or -1
func lastPercentiles(agp [][]float64) (int, []float64) {
	for slot := len(agp) - 1; slot >= 0; slot-- {
		if len(agp[slot]) == len(sitemodels.AGPPercentiles) {
			return slot, agp[slot]
		}
	}
	return -1, nil
}

// labelGap is how far apart, in SVG units, the AGP's percentile labels are kept, so close percentiles stay readable
const labelGap = 12.0

// percentileLabelYs are how far down the graph each of the percentiles' labels goes, beside its line's end, each
// moved up from the one under it to keep them labelGap apart
func percentileLabelYs(g graph, percentiles []float64) []float64 {
	ys := make([]float64, len(percentiles))
	for i, value := range percentiles {
		ys[i] = g.Y(value) + 4
		if i > 0 {
			ys[i] = min(ys[i], ys[i-1]-labelGap)
		}
	}
	return ys
}

// agpMarks are the glucose the AGP graph marks with a line and its value: very low, the range's low and high, very high,
// and its top
func agpMarks(g graph, low, high float64) []float64 {
	return []float64{g.Units.FromMGDL(sitemodels.VeryLowMgdl), low, high, g.Units.FromMGDL(sitemodels.VeryHighMgdl), g.Max}
}

// regions are the parts of a graph above the range high, in it, and below the range low, which the AGP's bands and a
// day's glucose are clipped to, to colour each, by the ID's suffix and the glucose colour each is filled or stroked with
type region struct {
	Suffix      string
	Colour      site_data_templ.GlucoseColour
	Top, Bottom func(g graph, low, high float64) float64
}

var regions = []region{
	{"above", site_data_templ.GlucoseColours["High"], func(g graph, _, _ float64) float64 { return 0 }, func(g graph, _, high float64) float64 { return g.Y(high) }},
	{"within", inRange, func(g graph, _, high float64) float64 { return g.Y(high) }, func(g graph, low, _ float64) float64 { return g.Y(low) }},
	{"below", site_data_templ.GlucoseColours["Low"], func(g graph, low, _ float64) float64 { return g.Y(low) }, func(g graph, _, _ float64) float64 { return g.Height }},
}

// profileDay is a day of a period's daily glucose profiles: its date, the day's summary if it has one, and whether it is
// in the period, as the days before a month's first in its first week are not
type profileDay struct {
	Date     time.Time
	Day      *sitemodels.SiteSummaryDaily
	InPeriod bool
}

// profileWeeks are the period's days as weeks from the site's week start day, each day with its summary if it has
// one, the days of the first week before the period's first left blank and those of the last after its last too
func profileWeeks(ctx context.Context, p *sitemodels.SiteSummaryPeriod, daily []*sitemodels.SiteSummaryDaily, location *time.Location) [][]profileDay {
	byDate := map[string]*sitemodels.SiteSummaryDaily{}
	for _, day := range daily {
		byDate[day.Date] = day
	}
	first, end := time.UnixMilli(p.Start).In(location), time.UnixMilli(p.End).In(location)
	// A month's first week starts on the site's week start day, and a week or fortnight's on its own first day
	before := 0
	if p.Period == "month" {
		before = (int(first.Weekday()) - weekStartDay(ctx) + 7) % 7
	}
	var weeks [][]profileDay
	for date := first.AddDate(0, 0, -before); date.Before(end); date = date.AddDate(0, 0, 1) {
		if len(weeks) == 0 || len(weeks[len(weeks)-1]) == 7 {
			weeks = append(weeks, []profileDay{})
		}
		inPeriod := !date.Before(first)
		weeks[len(weeks)-1] = append(weeks[len(weeks)-1], profileDay{Date: date, Day: byDate[date.Format(time.DateOnly)], InPeriod: inPeriod})
	}
	return weeks
}

// weekStartDay is the site's week start day, as a time.Weekday
func weekStartDay(ctx context.Context) int {
	if siteContext := httputils.GetSiteContext(ctx); siteContext != nil && siteContext.SiteModel != nil && siteContext.GetSiteWeekStartDay() == "sunday" {
		return int(time.Sunday)
	}
	return int(time.Monday)
}

// agpLegend is the AGP graph's key: its percentiles' bands and median, a day's own glucose over them, and the colours
// above and below the range
func agpLegend(p *sitemodels.SiteSummaryPeriod) []legendItem {
	items := []legendItem{{"h-3 w-3 rounded-sm bg-report-in-range/20", "site.reports.legend.outer_percentiles"},
		{"h-3 w-3 rounded-sm bg-report-in-range/50", "site.reports.legend.inner_percentiles"}, {"h-1 w-4 rounded-full bg-report-in-range", "site.reports.legend.median"}}
	if p.Period == "day" {
		items = append(items, legendItem{"h-1 w-4 rounded-full bg-brand-gold", "site.reports.legend.days_glucose"})
	}
	return append(items, legendItem{"h-3 w-3 rounded-sm bg-glucose-high/50", "site.reports.legend.above_range"}, legendItem{"h-3 w-3 rounded-sm bg-glucose-low/50", "site.reports.legend.below_range"})
}

// goalGlucose is the AGP's goal for the average glucose, below 154 mg/dL, in the units
func goalGlucose(ctx context.Context, units *consts.Units) string {
	return "<" + display.Glucose(ctx, units.FromMGDL(154), units) + " " + display.Units(units)
}
