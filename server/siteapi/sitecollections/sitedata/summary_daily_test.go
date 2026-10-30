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
		// 2026-10-26 is an even number of weeks from 5 Jan 1970, the first Monday, so starts a fortnight, as did 12 Oct
		{name: "in a fortnight's first week last fortnight ended last Sunday", period: "fortnight", now: at("2026-10-30 12:00", time.UTC), location: time.UTC, want: "2026-10-12"},
		{name: "in a fortnight's second week last fortnight ended the Sunday before last", period: "fortnight", now: at("2026-10-23 12:00", time.UTC), location: time.UTC, want: "2026-09-28"},
		{name: "on the Monday a fortnight starts last fortnight ended yesterday", period: "fortnight", now: at("2026-11-09 00:00", time.UTC), location: time.UTC, want: "2026-10-26"},
		// 2026-10-25 is an even number of weeks from 4 Jan 1970, the first Sunday
		{name: "from Sunday, fortnights start on Sundays", period: "fortnight", now: at("2026-10-30 12:00", time.UTC), location: time.UTC, weekStart: "sunday", want: "2026-10-11"},
		{name: "from Sunday, on the Sunday a fortnight starts last fortnight ended yesterday", period: "fortnight", now: at("2026-11-08 00:00", time.UTC), location: time.UTC, weekStart: "sunday", want: "2026-10-25"},
		{name: "the fortnight is in the site's timezone", period: "fortnight", now: at("2026-11-08 12:00", time.UTC), location: auckland, want: "2026-10-26"},
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
