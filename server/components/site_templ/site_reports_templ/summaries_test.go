package site_reports_templ

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
)

// The summaries report is of this month, or one of the year before it, in the site's timezone
func TestSummariesMonth(t *testing.T) {
	auckland, _ := time.LoadLocation("Pacific/Auckland")
	// 31 Oct 2026 23:00 UTC is already 1 Nov in Auckland
	now := time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		month string
		want  string
		ok    bool
	}{
		{"", "2026-11-01", true},
		{"2026-11", "2026-11-01", true},
		{"2025-11", "2025-11-01", true},
		{"2025-10", "", false},
		{"2026-12", "", false},
		{"2026-11-01", "", false},
		{"november", "", false},
	} {
		first, ok := SummariesMonth(test.month, now, auckland)
		require.Equal(t, test.ok, ok, test.month)
		if ok {
			require.Equal(t, test.want, first.Format(time.DateOnly), test.month)
			require.Equal(t, auckland, first.Location(), test.month)
		}
	}
}

// A month shows its days, the weeks and fortnights with a day in it, whichever day they start on, and its year's months
func TestSummariesDates(t *testing.T) {
	first := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.Equal(t, map[string][2]string{
		"day":       {"2026-02-01", "2026-02-28"},
		"week":      {"2026-01-26", "2026-02-28"},
		"fortnight": {"2026-01-19", "2026-02-28"},
		"month":     {"2026-01-01", "2026-12-31"},
	}, SummariesDates(first))
}

// The arrows go a month back as far as the year kept and forward to this month, which has no month query
func TestSummariesLinks(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		first          time.Time
		previous, next string
	}{
		{time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), "/reports/summaries?month=2026-10", ""},
		{time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "/reports/summaries?month=2026-09", "/reports/summaries"},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "/reports/summaries?month=2025-12", "/reports/summaries?month=2026-02"},
		{time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC), "", "/reports/summaries?month=2025-12"},
	} {
		previous, next := summariesLinks(test.first, now)
		require.Equal(t, test.previous, previous, test.first)
		require.Equal(t, test.next, next, test.first)
	}
}

// A past month with none of a period says so of the month, or of its year for months
func TestSummariesReportOfAPastMonth(t *testing.T) {
	first := time.Now().UTC().AddDate(0, -1, 0)
	first = time.Date(first.Year(), first.Month(), 1, 0, 0, 0, 0, time.UTC)
	var buf bytes.Buffer
	require.NoError(t, SummariesReport(first, map[string][]*sitemodels.SiteSummaryPeriod{}, false).Render(t.Context(), &buf))
	html := buf.String()
	require.Contains(t, html, `<span id="site--summaries--month" class="min-w-0 text-sm font-medium tabular-nums text-gray-700">`+first.Format("January 2006")+`</span>`)
	require.Contains(t, html, "No days kept in "+first.Format("January 2006")+".")
	require.Contains(t, html, "No months kept in "+first.Format("2006")+".")
	require.Contains(t, html, "Months of "+first.Format("2006"))
	require.NotContains(t, html, "kept yet")
}

// The longest periods come first: months, then fortnights, weeks and days
func TestSummariesReportListsMonthsFirst(t *testing.T) {
	first := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	var buf bytes.Buffer
	require.NoError(t, SummariesReport(first, map[string][]*sitemodels.SiteSummaryPeriod{}, false).Render(t.Context(), &buf))
	html, at := buf.String(), -1
	for _, period := range []string{"month", "fortnight", "week", "day"} {
		next := strings.Index(html, `id="site--summaries--`+period+`"`)
		require.Greater(t, next, at, period)
		at = next
	}
}
