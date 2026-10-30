package sitedata

import (
	"testing"
	"time"
)

func TestLastPeriodStart(t *testing.T) {
	auckland, _ := time.LoadLocation("Pacific/Auckland")
	at := func(value string, location *time.Location) time.Time {
		parsed, _ := time.ParseInLocation("2006-01-02 15:04", value, location)
		return parsed
	}
	tests := []struct {
		name     string
		period   string
		now      time.Time
		location *time.Location
		// weekStart is monday unless given
		weekStart string
		want      string
	}{
		{name: "on a Friday last week is from the Monday before this one", period: "week", now: at("2026-10-30 12:00", time.UTC), location: time.UTC, want: "2026-10-19"},
		{name: "on a Monday last week ended yesterday", period: "week", now: at("2026-10-26 00:00", time.UTC), location: time.UTC, want: "2026-10-19"},
		{name: "on a Sunday last week is still the one before", period: "week", now: at("2026-11-01 23:59", time.UTC), location: time.UTC, want: "2026-10-19"},
		{name: "the week is in the site's timezone", period: "week", now: at("2026-10-25 12:00", time.UTC), location: auckland, want: "2026-10-19"},
		{name: "from Sunday, on a Friday last week is from the Sunday before this one", period: "week", now: at("2026-10-30 12:00", time.UTC), location: time.UTC, weekStart: "sunday", want: "2026-10-18"},
		{name: "from Sunday, on a Sunday last week ended yesterday", period: "week", now: at("2026-11-01 00:00", time.UTC), location: time.UTC, weekStart: "sunday", want: "2026-10-25"},
		{name: "from Sunday, on a Saturday last week is still the one before", period: "week", now: at("2026-10-31 23:59", time.UTC), location: time.UTC, weekStart: "sunday", want: "2026-10-18"},
		{name: "last month is the one before this", period: "month", now: at("2026-10-30 12:00", time.UTC), location: time.UTC, want: "2026-09-01"},
		{name: "in January last month is last year's December", period: "month", now: at("2026-01-15 12:00", time.UTC), location: time.UTC, want: "2025-12-01"},
		{name: "the month is in the site's timezone", period: "month", now: at("2026-10-31 12:00", time.UTC), location: auckland, want: "2026-10-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weekStart := tt.weekStart
			if weekStart == "" {
				weekStart = "monday"
			}
			start := LastPeriodStart(tt.period, tt.now, tt.location, weekStart)
			if got := start.Format(time.DateOnly); got != tt.want || start.Location() != tt.location || start.Hour() != 0 {
				t.Errorf("last %s starts %v, want midnight on %s in %s", tt.period, start, tt.want, tt.location)
			}
		})
	}
}
