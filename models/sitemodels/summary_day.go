package sitemodels

import "nighttroop.t1d.nz/utils/consts"

// SummaryDay is a day from Start until End, in milliseconds, with its time in range and total daily dose
type SummaryDay struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	// TimeInRange is the percent of the day's readings from the low alarm to the high alarm, null when it has none
	TimeInRange *float64 `json:"time_in_range"`
	// TotalDailyDose is the day's insulin doses and the basal in force through it, in U
	TotalDailyDose float64 `json:"total_daily_dose"`
}

// NewSummaryDay is the day from start until end, from readings, doses and minutes that may run outside it. Readings
// in range are from low to high, and the minutes are each a minute apart from minutesSince.
func NewSummaryDay(start, end int64, low, high float64, readings []*SiteCleanGlucoseReading, doses []*SiteInsulinDose, minutesSince int64, minutes []SummaryMinute) SummaryDay {
	day := SummaryDay{Start: start, End: end}
	within := func(at int64) bool { return at >= start && at < end }
	count, inRange := 0, 0
	for _, r := range readings {
		if within(r.OccurredAt) {
			count++
			if r.Reading >= low && r.Reading <= high {
				inRange++
			}
		}
	}
	if count > 0 {
		day.TimeInRange = new(consts.Round(100*float64(inRange)/float64(count), 1))
	}
	for _, d := range doses {
		if within(d.OccurredAt) {
			day.TotalDailyDose += d.Dose
		}
	}
	for i, m := range minutes {
		if within(minutesSince+int64(i)*60_000) && m.Basal != nil {
			day.TotalDailyDose += *m.Basal / 60
		}
	}
	day.TotalDailyDose = consts.Round(day.TotalDailyDose, 2)
	return day
}
