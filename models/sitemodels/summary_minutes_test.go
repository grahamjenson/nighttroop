package sitemodels

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

func TestSummaryMinutes(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	until := since + 4*time.Hour.Milliseconds()
	minute := func(m int64) int64 { return since + m*60_000 }
	str := func(s string) *string { return &s }
	targets := func(mgdl string) *string { return str(`[{"time":"00:00","value":` + mgdl + `,"timeAsSeconds":0}]`) }
	// 0.5 U/hr from midnight and 0.8 U/hr from 2am
	basalRates := str(`[{"time":"00:00","value":0.5,"timeAsSeconds":0},{"time":"02:00","value":0.8,"timeAsSeconds":7200}]`)
	defaultProfile := &sitedb.ProfilePreset{OccurredAt: since - 1000, TargetsLowJson: targets("100"), TargetsHighJson: targets("160"), BasalRatesJson: basalRates}
	mgdl, mmol := consts.UnitsMGDL, consts.UnitsMMOLL

	type want struct {
		minute           int64
		low, high, basal *float64
	}
	tests := []struct {
		name     string
		units    *consts.Units
		defaults []*sitedb.ProfilePreset
		switches []*sitedb.ProfileSwitch
		targets  []*SiteTempTarget
		basals   []*SiteTempBasal
		want     []want
	}{
		{
			name:  "with nothing in force every minute is null",
			units: &mgdl,
			want:  []want{{0, nil, nil, nil}, {240, nil, nil, nil}},
		},
		{
			name:     "the default profile's schedules, through the day",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			want:     []want{{0, new(100.0), new(160.0), new(0.5)}, {119, new(100.0), new(160.0), new(0.5)}, {120, new(100.0), new(160.0), new(0.8)}},
		},
		{
			name:     "a default uploaded during the summary applies from then",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile, {OccurredAt: minute(60), TargetsLowJson: targets("90"), TargetsHighJson: targets("150"), BasalRatesJson: basalRates}},
			want:     []want{{59, new(100.0), new(160.0), new(0.5)}, {60, new(90.0), new(150.0), new(0.5)}},
		},
		{
			name:     "targets are in the settings' units",
			units:    &mmol,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			want:     []want{{0, new(mmol.FromMGDL(100)), new(mmol.FromMGDL(160)), new(0.5)}},
		},
		{
			name:     "a switch's schedules, percent and timeshift override the default",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			// Shifted two hours forward, at 00:30 it is 02:30 in the profile, so 0.8 U/hr, at 50%
			switches: []*sitedb.ProfileSwitch{{OccurredAt: minute(30), TargetsLowJson: targets("110"), TargetsHighJson: targets("140"), BasalRatesJson: basalRates, Percent: new(50.0), Timeshift: new(int64(120))}},
			want:     []want{{29, new(100.0), new(160.0), new(0.5)}, {30, new(110.0), new(140.0), new(0.4)}},
		},
		{
			name:     "a switch without schedules keeps the default's, at its percent",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			switches: []*sitedb.ProfileSwitch{{OccurredAt: minute(30), Percent: new(200.0)}},
			want:     []want{{30, new(100.0), new(160.0), new(1.0)}},
		},
		{
			name:     "a switch with a duration hands back to the last switch without one",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			switches: []*sitedb.ProfileSwitch{
				{OccurredAt: minute(10), TargetsLowJson: targets("110")},
				{OccurredAt: minute(30), DurationMinutes: 30, TargetsLowJson: targets("120")},
			},
			want: []want{{10, new(110.0), new(160.0), new(0.5)}, {30, new(120.0), new(160.0), new(0.5)}, {59, new(120.0), new(160.0), new(0.5)}, {60, new(110.0), new(160.0), new(0.5)}},
		},
		{
			name:     "a profile's own timezone reads its schedule",
			units:    &mgdl,
			// 00:00 UTC is 13:00 in Auckland in January, after the 2am change
			defaults: []*sitedb.ProfilePreset{{OccurredAt: since - 1000, Timezone: str("Pacific/Auckland"), BasalRatesJson: basalRates}},
			want:     []want{{0, nil, nil, new(0.8)}},
		},
		{
			name:     "a temp target overrides the targets until it ends, even one started before the summary",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			targets:  []*SiteTempTarget{{OccurredAt: minute(-10), DurationMinutes: 20, LowTarget: 140, HighTarget: 150}, {OccurredAt: minute(60), DurationMinutes: 30, LowTarget: 80, HighTarget: 90}},
			want:     []want{{9, new(140.0), new(150.0), new(0.5)}, {10, new(100.0), new(160.0), new(0.5)}, {60, new(80.0), new(90.0), new(0.5)}, {90, new(100.0), new(160.0), new(0.5)}},
		},
		{
			name:     "a newer temp target replaces an older one, as one of no duration cancels it",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			targets:  []*SiteTempTarget{{OccurredAt: minute(0), DurationMinutes: 120, LowTarget: 140, HighTarget: 150}, {OccurredAt: minute(30), DurationMinutes: 0}},
			want:     []want{{29, new(140.0), new(150.0), new(0.5)}, {30, new(100.0), new(160.0), new(0.5)}},
		},
		{
			name:     "a temp basal's absolute, or its rate times the basal, overrides the basal",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			basals:   []*SiteTempBasal{{OccurredAt: minute(10), DurationMinutes: 20, Absolute: new(0.0)}, {OccurredAt: minute(60), DurationMinutes: 30, Rate: new(0.5)}},
			want:     []want{{10, new(100.0), new(160.0), new(0.0)}, {30, new(100.0), new(160.0), new(0.5)}, {60, new(100.0), new(160.0), new(0.25)}},
		},
		{
			// Its absolute is the basal at its start, 0.5 U/hr at half, but from 2am the basal is 0.8 U/hr
			name:     "a temp basal's rate is times the basal of each minute, over its absolute",
			units:    &mgdl,
			defaults: []*sitedb.ProfilePreset{defaultProfile},
			basals:   []*SiteTempBasal{{OccurredAt: minute(90), DurationMinutes: 60, Absolute: new(0.25), Rate: new(0.5)}},
			want:     []want{{119, new(100.0), new(160.0), new(0.25)}, {120, new(100.0), new(160.0), new(0.4)}, {150, new(100.0), new(160.0), new(0.8)}},
		},
		{
			name:   "without a profile a temp basal's absolute is in force, but not a rate alone",
			units:  &mgdl,
			basals: []*SiteTempBasal{{OccurredAt: minute(0), DurationMinutes: 10, Absolute: new(1.2), Rate: new(2.0)}, {OccurredAt: minute(10), DurationMinutes: 10, Rate: new(1.5)}},
			want:   []want{{0, nil, nil, new(1.2)}, {10, nil, nil, nil}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			minutes := SummaryMinutes(since, until, test.units, time.UTC, test.defaults, test.switches, test.targets, test.basals)
			require.Len(t, minutes, 241)
			for _, w := range test.want {
				got := minutes[w.minute]
				require.Equal(t, w.low, got.LowTarget, "low target at minute %d", w.minute)
				require.Equal(t, w.high, got.HighTarget, "high target at minute %d", w.minute)
				if w.basal == nil {
					require.Nil(t, got.Basal, "basal at minute %d", w.minute)
				} else {
					require.InDelta(t, *w.basal, *got.Basal, 1e-9, "basal at minute %d", w.minute)
				}
			}
		})
	}
}

func TestRateBasal(t *testing.T) {
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	at := func(hour, minute int) int64 {
		return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute).UnixMilli()
	}
	str := func(s string) *string { return &s }
	// 0.5 U/hr from midnight and 0.8 U/hr from 2am
	basalRates := str(`[{"time":"00:00","value":0.5,"timeAsSeconds":0},{"time":"02:00","value":0.8,"timeAsSeconds":7200}]`)
	utc := []*sitedb.ProfilePreset{{OccurredAt: day.UnixMilli() - 1000, BasalRatesJson: basalRates}}
	// Auckland is 13 hours ahead in January, so its 2am is 13:00 UTC
	auckland := []*sitedb.ProfilePreset{{OccurredAt: day.UnixMilli() - 1000, BasalRatesJson: basalRates, Timezone: str("Pacific/Auckland")}}
	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	tests := []struct {
		name     string
		ms       int64
		rate     float64
		location *time.Location
		defaults []*sitedb.ProfilePreset
		switches []*sitedb.ProfileSwitch
		want     *float64
	}{
		{name: "without a profile there is no rate", ms: at(1, 0), rate: 0.5, location: time.UTC},
		{name: "a profile uploaded after the temp basal is not in force", ms: at(1, 0), rate: 0.5, location: time.UTC, defaults: []*sitedb.ProfilePreset{{OccurredAt: at(2, 0), BasalRatesJson: basalRates}}},
		{name: "under 1 lowers the basal", ms: at(1, 0), rate: 0.5, location: time.UTC, defaults: utc, want: new(0.25)},
		{name: "over 1 raises the basal", ms: at(1, 59), rate: 1.5, location: time.UTC, defaults: utc, want: new(0.75)},
		{name: "from a schedule's boundary its next rate", ms: at(2, 0), rate: 1.5, location: time.UTC, defaults: utc, want: new(1.2)},
		{name: "late in the day the last rate", ms: at(23, 59), rate: 1.25, location: time.UTC, defaults: utc, want: new(1.0)},
		{name: "1 keeps the basal", ms: at(3, 0), rate: 1, location: time.UTC, defaults: utc, want: new(0.8)},
		{name: "0 is a zero temp", ms: at(3, 0), rate: 0, location: time.UTC, defaults: utc, want: new(0.0)},
		{name: "in the profile's timezone, before its 2am", ms: at(12, 59), rate: 2, location: time.UTC, defaults: auckland, want: new(1.0)},
		{name: "in the profile's timezone, from its 2am", ms: at(13, 0), rate: 2, location: time.UTC, defaults: auckland, want: new(1.6)},
		{name: "a profile without a timezone is in the site's, before its 2am", ms: at(6, 59), rate: 0.8, location: newYork, defaults: utc, want: new(0.4)},
		{name: "a profile without a timezone is in the site's, from its 2am", ms: at(7, 0), rate: 0.8, location: newYork, defaults: utc, want: new(0.64)},
		{
			// Shifted two hours forward, at 00:30 it is 02:30 in the profile, so 0.8 U/hr, at 200% is 1.6, times 0.75
			name: "a profile switch's percent and timeshift", ms: at(0, 30), rate: 0.75, location: time.UTC, defaults: utc,
			switches: []*sitedb.ProfileSwitch{{OccurredAt: at(0, 10), BasalRatesJson: basalRates, Percent: new(200.0), Timeshift: new(int64(120))}},
			want:     new(1.2),
		},
		{
			name: "a profile switch that has ended hands back to the default", ms: at(1, 0), rate: 0.75, location: time.UTC, defaults: utc,
			switches: []*sitedb.ProfileSwitch{{OccurredAt: at(0, 10), DurationMinutes: 30, Percent: new(200.0)}},
			want:     new(0.375),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RateBasal(test.ms, test.rate, test.location, test.defaults, test.switches)
			if test.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.InDelta(t, *test.want, *got, 1e-9)
		})
	}
}
