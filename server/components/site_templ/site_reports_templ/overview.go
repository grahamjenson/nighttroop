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

// gmiDays and cgmActiveGoal are the days with data a GMI needs, and the percent of the time the CGM must read, for a
// report's numbers to be reliable, as the international consensus on CGM metrics has them
const (
	gmiDays       = 14
	cgmActiveGoal = 70.0
)

// gmiNote cautions that a GMI of fewer than gmiDays days with data is less reliable
func gmiNote(p *sitemodels.SiteSummaryPeriod) string {
	if p.GMI == nil || p.Days >= gmiDays {
		return ""
	}
	return fmt.Sprintf("Reliable from %d days of data", gmiDays)
}

// cgmActiveNote cautions that numbers from a CGM that read for less than cgmActiveGoal of the time may not be typical
func cgmActiveNote(p *sitemodels.SiteSummaryPeriod) string {
	if p.CGMActive == nil || *p.CGMActive >= cgmActiveGoal {
		return ""
	}
	return fmt.Sprintf("Under %.0f%%, so these numbers may not be typical", cgmActiveGoal)
}

// previousOf is the report before, or an empty one when there is none, so its numbers are all missing
func previousOf(previous *sitemodels.SiteSummaryPeriod) *sitemodels.SiteSummaryPeriod {
	if previous == nil {
		return &sitemodels.SiteSummaryPeriod{}
	}
	return previous
}

// previousName is what the report before p is, as the previous 7 days or the previous month
func previousName(p *sitemodels.SiteSummaryPeriod) string {
	if n, ok := sitemodels.ReportDays[p.Period]; ok {
		return fmt.Sprintf("the previous %d days", n)
	}
	return map[string]string{"day": "the day before", "month": "the previous month"}[p.Period]
}

// change is how much a number of the report rose or fell from the one before, to decimals places, with what it is
// in, as ▲ 0.3 mmol/L or ▼ 4%, "No change" when the same, or "" when either is missing
func change(now, before *float64, decimals int, unit string) string {
	if now == nil || before == nil {
		return ""
	}
	difference := consts.Round(*now-*before, decimals)
	switch {
	case difference > 0:
		return fmt.Sprintf("▲ %.*f%s", decimals, difference, unit)
	case difference < 0:
		return fmt.Sprintf("▼ %.*f%s", decimals, -difference, unit)
	}
	return "No change"
}

// glucoseChange is how much the average glucose rose or fell from the report before, in its units, mmol/L to one
// decimal place and mg/dL whole
func glucoseChange(p, previous *sitemodels.SiteSummaryPeriod) string {
	decimals := 0
	if p.Units != nil && *p.Units == consts.UnitsMMOLL {
		decimals = 1
	}
	return change(p.Mean, previousOf(previous).Mean, decimals, " "+display.Units(p.Units))
}

// oneDecimal is a number to one decimal place, or "–" for none
func oneDecimal(v *float64) string {
	if v == nil {
		return "–"
	}
	return fmt.Sprintf("%.1f", *v)
}

// band is a range of glucose a report counts readings in: its ID, its name, the percent of readings in it, and the
// classes it is drawn with
type band struct {
	ID, Label string
	Percent   *float64
	Class     string
}

// bands are the period's very high, high, in range, low and very low, top down as they are drawn
func bands(p *sitemodels.SiteSummaryPeriod) []band {
	colours := site_data_templ.GlucoseColours
	return []band{
		{"very-high", "Very high", p.TimeVeryHigh, colours["Very high"].Background},
		{"high", "High", p.TimeHigh, colours["High"].Background},
		{"in-range", "In range", p.TimeInRange, colours[""].Background},
		{"low", "Low", p.TimeLow, colours["Low"].Background},
		{"very-low", "Very low", p.TimeVeryLow, colours["Very low"].Background},
	}
}

// bandHeight is how much of the time in range bar a band takes, at least a sliver for one with any readings, so it
// shows
func bandHeight(percent *float64) string {
	if percent == nil || *percent <= 0 {
		return "0"
	}
	return fmt.Sprintf("max(%.1f%%, 3px)", *percent)
}

// periodDays is how many of the period's days have begun by now: all of them once it is over
func periodDays(p *sitemodels.SiteSummaryPeriod, now time.Time, location *time.Location) int {
	days, end := 0, min(p.End, now.UnixMilli())
	for day := time.UnixMilli(p.Start).In(location); day.UnixMilli() < end; day = day.AddDate(0, 0, 1) {
		days++
	}
	return days
}

// perDay is a count spread over the days, to one decimal place, or "–" with none
func perDay(count int64, days int) string {
	if days == 0 {
		return "–"
	}
	return fmt.Sprintf("%.1f", float64(count)/float64(days))
}

// bestDay is the whole day with the most time in range, the last of those with the same, or nil when none had
// readings; today is left out, as it is not over, and a day the CGM read for under cgmActiveGoal of is only when none
// read for more
func bestDay(days []*sitemodels.SiteSummaryDaily, now time.Time) *sitemodels.SiteSummaryDaily {
	var best *sitemodels.SiteSummaryDaily
	for _, worn := range []bool{true, false} {
		for _, day := range days {
			if day.TimeInRange != nil && day.End <= now.UnixMilli() && (!worn || cgmWorn(day)) && (best == nil || *day.TimeInRange >= *best.TimeInRange) {
				best = day
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

// cgmWorn is whether the day has glucose in at least cgmActiveGoal of its 5 minutes, so its time in range is of the
// whole day rather than a few hours of it, as a backfill's first day can be
func cgmWorn(day *sitemodels.SiteSummaryDaily) bool {
	read := 0
	for _, value := range day.Glucose {
		if value != nil {
			read++
		}
	}
	return len(day.Glucose) > 0 && 100*float64(read)/float64(len(day.Glucose)) >= cgmActiveGoal
}

// overviewBar is 15 minutes of the period's day: the 25th and 75th percentiles of its glucose, and its mean
type overviewBar struct {
	Slot           int
	Lower, Upper   float64
	Mean           *float64
	HasPercentiles bool
}

// overviewSlotsPerBar is how many of the AGP's and glucose's 5 minutes each overview bar is of
const overviewSlotsPerBar = 3

// overviewBars are each 15 minutes of the period's day from midnight, 96 of them: the 25th and 75th percentiles of the
// AGP at the middle 5 minutes of each, and the mean of its glucose in them
func overviewBars(p *sitemodels.SiteSummaryPeriod) []overviewBar {
	bars := make([]overviewBar, 24*60*60_000/(overviewSlotsPerBar*sitemodels.SummaryDailySlotMs))
	for i := range bars {
		bars[i].Slot = i
		if middle := i*overviewSlotsPerBar + 1; middle < len(p.AGP) && len(p.AGP[middle]) == len(sitemodels.AGPPercentiles) {
			bars[i].Lower, bars[i].Upper, bars[i].HasPercentiles = p.AGP[middle][1], p.AGP[middle][3], true
		}
		sum, count := 0.0, 0
		for slot := i * overviewSlotsPerBar; slot < (i+1)*overviewSlotsPerBar && slot < len(p.Glucose); slot++ {
			if p.Glucose[slot] != nil {
				sum, count = sum+*p.Glucose[slot], count+1
			}
		}
		if count > 0 {
			bars[i].Mean = new(sum / float64(count))
		}
	}
	return bars
}

// graph is where a report's graph of glucose over a day from midnight is drawn, in SVG units: its plot from Left to
// Right and Top to Bottom, of glucose from Min to Max
type graph struct {
	Width, Height            float64
	Left, Right, Top, Bottom float64
	Min, Max                 float64
	Units                    *consts.Units
}

// newGraph is a graph Width wide and Height high, of glucose from 40 mg/dL to at least 300 mg/dL, or above top, in units
func newGraph(width, height float64, units *consts.Units, top float64) graph {
	step := units.FromMGDL(50)
	return graph{Width: width, Height: height, Left: 48, Right: width - 48, Top: 24, Bottom: height - 28,
		Min: units.FromMGDL(40), Max: max(units.FromMGDL(300), math.Ceil(top/step)*step), Units: units}
}

// X is how far across the graph a time of day, in milliseconds from midnight, is drawn
func (g graph) X(ms float64) float64 {
	return g.Left + (g.Right-g.Left)*ms/(24*60*60_000)
}

// Y is how far down the graph glucose is drawn, kept inside it
func (g graph) Y(value float64) float64 {
	return g.Bottom - (g.Bottom-g.Top)*display.Scale(value, g.Min, g.Max)
}

// Ticks are the glucose the graph marks, every 50 mg/dL, or 2.5 mmol/L, from its bottom to its top
func (g graph) Ticks() []float64 {
	step := 50.0
	if g.Units != nil && *g.Units == consts.UnitsMMOLL {
		step = 2.5
	}
	var ticks []float64
	for tick := math.Ceil(g.Min/step) * step; tick <= g.Max; tick += step {
		ticks = append(ticks, tick)
	}
	return ticks
}

// Hours are the times of day the graph marks, every 3 hours from midnight to midnight
func (g graph) Hours() []int {
	return []int{0, 3, 6, 9, 12, 15, 18, 21, 24}
}

// f is an SVG number, to one decimal place
func f(v float64) string {
	return fmt.Sprintf("%.1f", v)
}

// hourLabel is an hour of the day as the settings show times: 12am, 3am and 12pm for a 12 hour clock, else 00:00
func hourLabel(ctx context.Context, hour int) string {
	if !twelveHour(ctx) {
		return fmt.Sprintf("%02d:00", hour%24)
	}
	return strings.ToLower(time.Date(2000, 1, 1, hour%24, 0, 0, 0, time.UTC).Format("3PM"))
}

// timeOfDay is minutes from midnight as the settings show times: 2:05pm for a 12 hour clock, else 14:05
func timeOfDay(ctx context.Context, minutes int) string {
	if !twelveHour(ctx) {
		return fmt.Sprintf("%02d:%02d", minutes/60%24, minutes%60)
	}
	return strings.ToLower(time.Date(2000, 1, 1, 0, minutes, 0, 0, time.UTC).Format("3:04PM"))
}

// twelveHour is whether the settings show times on a 12 hour clock
func twelveHour(ctx context.Context) bool {
	layout := time.TimeOnly
	if siteContext := httputils.GetSiteContext(ctx); siteContext != nil {
		layout = siteContext.GetSettings().TimeLayout()
	}
	return strings.Contains(layout, "PM")
}

// overviewLegend is the overview graph's key: its bars above and below the range and between its 25th and 75th
// percentiles, and its average
var overviewLegend = []legendItem{{"h-3 w-2 rounded-sm bg-glucose-high", "Above the range"}, {"h-3 w-2 rounded-sm bg-gray-300", "25th to 75th percentile"},
	{"h-1.5 w-1.5 rounded-full bg-gray-900", "Average"}, {"h-3 w-2 rounded-sm bg-glucose-low", "Below the range"}}

// overviewTop is the highest 75th percentile of the period, so its graph shows all its bars
func overviewTop(bars []overviewBar) float64 {
	top := 0.0
	for _, bar := range bars {
		top = max(top, bar.Upper)
	}
	return top
}

// tddLink is the TDD report of the month ending on the day, or the latest, to today so far, for today or later
func tddLink(day, now time.Time) string {
	if day.Format(time.DateOnly) >= now.In(day.Location()).Format(time.DateOnly) {
		return "/reports/tdd"
	}
	return "/reports/tdd?end=" + day.Format(time.DateOnly)
}
