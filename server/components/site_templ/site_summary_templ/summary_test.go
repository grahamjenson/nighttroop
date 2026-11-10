package site_summary_templ

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// TestGlucoseHeroColoursItsRange checks the latest glucose is the theme's low colour and says Low below the low alarm,
// its high colour and says High above the high alarm, and its in range colour between them, in either units
func TestGlucoseHeroColoursItsRange(t *testing.T) {
	mmol, mgdl := consts.UnitsMMOLL, consts.UnitsMGDL
	tests := []struct {
		name    string
		reading float64
		units   *consts.Units
		colour  string
		label   string
	}{
		{"low in mmol/L", 3.5, &mmol, "text-glucose-low", "Low"},
		{"in range in mmol/L", 6.5, &mmol, "text-glucose-in-range", ""},
		{"high in mmol/L", 11.2, &mmol, "text-glucose-high-text", "High"},
		{"low in mg/dL", 63, &mgdl, "text-glucose-low", "Low"},
		{"at the low alarm in mg/dL", 70, &mgdl, "text-glucose-in-range", ""},
		{"at the high alarm in mg/dL", 180, &mgdl, "text-glucose-in-range", ""},
		{"high in mg/dL", 181, &mgdl, "text-glucose-high-text", "High"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The alarms are 70 and 180 mg/dL, as the summary has them in its units
			summary := &sitemodels.Summary{Units: tt.units, RangeLow: tt.units.FromMGDL(70), RangeHigh: tt.units.FromMGDL(180), GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: tt.reading, Units: tt.units, OccurredAt: time.Now().UnixMilli()}}}}
			var buf bytes.Buffer
			require.NoError(t, glucoseHero(summary).Render(context.Background(), &buf))
			html := buf.String()
			require.Contains(t, html, `id="site--summary--glucose--reading" class="text-6xl font-extrabold tracking-tight `+tt.colour+`"`)
			if tt.label == "" {
				require.NotContains(t, html, "site--summary--glucose--range")
			} else {
				require.Contains(t, html, `id="site--summary--glucose--range" class="ml-2 text-2xl font-bold `+tt.colour+`">`+tt.label+`</span>`)
			}
		})
	}
}

// TestGlucoseHeroGreysAnOldReading checks a reading older than 15 minutes, without settings, is grey with a badge
// saying how long there has been none, and a newer one is coloured as its range with no badge
func TestGlucoseHeroGreysAnOldReading(t *testing.T) {
	mmol := consts.UnitsMMOLL
	tests := []struct {
		name  string
		age   time.Duration
		badge string
	}{
		{"recent", 14 * time.Minute, ""},
		{"at the threshold", 15 * time.Minute, "No reading for 15 min"},
		{"old", 45*time.Minute + 30*time.Second, "No reading for 45 min"},
		{"hours old", 3*time.Hour + 20*time.Minute, "No reading for 3 h 20 min"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := sitemodels.SiteGlucoseReading{Reading: 6.5, Units: &mmol, OccurredAt: time.Now().Add(-tt.age).UnixMilli()}
			summary := &sitemodels.Summary{Units: &mmol, RangeLow: 3.9, RangeHigh: 10, GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: reading}}}
			var buf bytes.Buffer
			require.NoError(t, glucoseHero(summary).Render(context.Background(), &buf))
			html := buf.String()
			if tt.badge == "" {
				require.NotContains(t, html, "site--summary--glucose--stale")
				require.Contains(t, html, `class="text-6xl font-extrabold tracking-tight text-glucose-in-range"`)
			} else {
				require.Regexp(t, `id="site--summary--glucose--stale" class="[^"]*\btext-amber-800\b[^"]*">`+tt.badge+`</span>`, html)
				require.Contains(t, html, `class="text-6xl font-extrabold tracking-tight text-gray-400"`)
			}
		})
	}
}

// TestSummaryReloadsWhenItsReadingTurnsStale checks the summary's content is reloaded a second after its latest
// reading turns stale, after 15 minutes without settings, and not when it has no reading or already is; a reading from
// the future waits only the 15 minutes, not longer than the browser's timer can
func TestSummaryReloadsWhenItsReadingTurnsStale(t *testing.T) {
	mmol, now := consts.UnitsMMOLL, time.UnixMilli(time.Now().UnixMilli())
	tests := []struct {
		name    string
		age     time.Duration
		trigger string
	}{
		{"new", 0, ", load delay:901000ms"},
		{"10 minutes old", 10 * time.Minute, ", load delay:301000ms"},
		{"just under", 15*time.Minute - 500*time.Millisecond, ", load delay:1500ms"},
		{"stale", 15 * time.Minute, ""},
		{"from the future", -100 * 24 * time.Hour, ", load delay:901000ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := sitemodels.SiteGlucoseReading{Reading: 6.5, Units: &mmol, OccurredAt: now.Add(-tt.age).UnixMilli()}
			summary := &sitemodels.Summary{Units: &mmol, GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: reading}}}
			require.Equal(t, tt.trigger, staleTrigger(context.Background(), summary, now))
		})
	}
	require.Empty(t, staleTrigger(context.Background(), &sitemodels.Summary{Units: &mmol}, now))

	// The rendered summary schedules it after its stream's triggers
	reading := sitemodels.SiteGlucoseReading{Reading: 6.5, Units: &mmol, OccurredAt: now.Add(-10 * time.Minute).UnixMilli()}
	summary := &sitemodels.Summary{Hours: 6, Units: &mmol, GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: reading}}}
	var buf bytes.Buffer
	require.NoError(t, Summary(summary).Render(context.Background(), &buf))
	require.Regexp(t, `reconnected from:#site--summary--live, load delay:30\d{4}ms"`, buf.String())
}
