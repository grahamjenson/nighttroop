package site_reports_templ

import (
	"time"
)

// SummariesMonths is how many months before this one the summaries report goes back, the year regenerating builds
const SummariesMonths = 12

// SummariesMonth is the first day of the month the summaries report shows, in location: month, as 2026-10, or this
// month when it is "", and whether it is this month or one of the SummariesMonths before it
func SummariesMonth(month string, now time.Time, location *time.Location) (time.Time, bool) {
	local := now.In(location)
	this := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location)
	if month == "" {
		return this, true
	}
	first, err := time.ParseInLocation("2006-01", month, location)
	return first, err == nil && summariesKept(first, this)
}

// summariesKept is whether the month starting first is this one or one of the SummariesMonths before it
func summariesKept(first, this time.Time) bool {
	return !first.After(this) && !first.Before(this.AddDate(0, -SummariesMonths, 0))
}

// SummariesDates are the first and last dates of each period's summaries the report of the month starting first
// shows: its days, the weeks and fortnights with a day in it, as they start up to 6 and 13 days before it, and the
// months of its year
func SummariesDates(first time.Time) map[string][2]string {
	last := first.AddDate(0, 1, -1).Format(time.DateOnly)
	year := time.Date(first.Year(), 1, 1, 0, 0, 0, 0, first.Location())
	return map[string][2]string{
		"day":       {first.Format(time.DateOnly), last},
		"week":      {first.AddDate(0, 0, -6).Format(time.DateOnly), last},
		"fortnight": {first.AddDate(0, 0, -13).Format(time.DateOnly), last},
		"month":     {year.Format(time.DateOnly), year.AddDate(1, 0, -1).Format(time.DateOnly)},
	}
}

// summariesLinks are the summaries reports of the month before and after the one starting first, "" where there is
// none; this month's has no month query
func summariesLinks(first, now time.Time) (previous, next string) {
	local := now.In(first.Location())
	this := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, first.Location())
	link := func(at time.Time) string {
		switch {
		case !summariesKept(at, this):
			return ""
		case at.Equal(this):
			return "/reports/summaries"
		}
		return "/reports/summaries?month=" + at.Format("2006-01")
	}
	return link(first.AddDate(0, -1, 0)), link(first.AddDate(0, 1, 0))
}

// summariesTitle names a table of a month's summaries of period, as Weeks, or Months of 2026 for its year's months
func summariesTitle(period, label string, first time.Time) string {
	if period == "month" {
		return label + "s of " + first.Format("2006")
	}
	return label + "s"
}

// emptyIn is when a past month's report has no summaries of period: in the month, or in its year for months
func emptyIn(period string, first time.Time) string {
	if period == "month" {
		return first.Format("2006")
	}
	return first.Format("January 2006")
}
