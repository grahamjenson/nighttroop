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
			summary := &sitemodels.Summary{Units: tt.units, RangeLow: tt.units.FromMGDL(70), RangeHigh: tt.units.FromMGDL(180), CurrentGlucose: &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: tt.reading, Units: tt.units, OccurredAt: time.Now().UnixMilli()}}}
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
			summary := &sitemodels.Summary{Units: &mmol, RangeLow: 3.9, RangeHigh: 10, CurrentGlucose: &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: reading}}
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
			summary := &sitemodels.Summary{Units: &mmol, CurrentGlucose: &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: reading}}
			require.Equal(t, tt.trigger, staleTrigger(context.Background(), summary, now))
		})
	}
	require.Empty(t, staleTrigger(context.Background(), &sitemodels.Summary{Units: &mmol}, now))

	// The rendered summary schedules it after its stream's triggers
	reading := sitemodels.SiteGlucoseReading{Reading: 6.5, Units: &mmol, OccurredAt: now.Add(-10 * time.Minute).UnixMilli()}
	summary := &sitemodels.Summary{Hours: 6, Units: &mmol, CurrentGlucose: &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: reading}}
	var buf bytes.Buffer
	require.NoError(t, Summary(summary).Render(context.Background(), &buf))
	require.Regexp(t, `reconnected from:#site--summary--live, load delay:30\d{4}ms"`, buf.String())
}

// A dose in the insulin and carbs table shows its extended dose and the time it is given over, after the dose given at
// once unless that is 0, and an extended dose with no time is given at once
func TestDoseText(t *testing.T) {
	tests := []struct {
		name string
		dose *sitemodels.SiteInsulinDose
		text string
	}{
		{"a bolus", &sitemodels.SiteInsulinDose{Dose: 2.5}, "2.5 U"},
		{"a combo bolus", &sitemodels.SiteInsulinDose{Dose: 1.2, ExtendedDose: new(1.8), ExtendedDurationMinutes: new(int64(120))}, "1.2 U + 1.8 U over 2 h"},
		{"an extended bolus", &sitemodels.SiteInsulinDose{Dose: 0, ExtendedDose: new(2.0), ExtendedDurationMinutes: new(int64(90))}, "2 U over 1 h 30 min"},
		{"an extended dose with no time", &sitemodels.SiteInsulinDose{Dose: 1, ExtendedDose: new(2.0)}, "3 U"},
		{"an extended dose of 0", &sitemodels.SiteInsulinDose{Dose: 1, ExtendedDose: new(0.0), ExtendedDurationMinutes: new(int64(120))}, "1 U"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.text, doseText(context.Background(), tt.dose, " U"))
		})
	}
	// On the graph its unit follows the number, as in 2.5U
	combo := &sitemodels.SiteInsulinDose{Dose: 0.3, ExtendedDose: new(0.7), ExtendedDurationMinutes: new(int64(60))}
	require.Equal(t, "0.3U + 0.7U over 1 h", doseText(context.Background(), combo, "U"))
}

// TestInsulinCarbsTableSaysWhenAPlannedIntakeIs checks carbs planned an hour ahead, as Trio's from fat and protein, say
// "in 58 minutes", not "just now", while one only minutes ahead, as from a phone's clock running fast, is "just now"
func TestInsulinCarbsTableSaysWhenAPlannedIntakeIs(t *testing.T) {
	now := time.Now()
	render := func(at time.Time) string {
		var buf bytes.Buffer
		require.NoError(t, insulinCarbsTable([]intake{{At: at.UnixMilli(), Carbs: "17 g"}}).Render(context.Background(), &buf))
		return buf.String()
	}
	require.Contains(t, render(now.Add(58*time.Minute+30*time.Second)), ">in 58 minutes</span>")
	require.Contains(t, render(now.Add(5*time.Minute)), ">just now</span>")
	require.Contains(t, render(now.Add(-2*time.Minute-time.Second)), ">2 minutes ago</span>")
}

// TestGlucoseHeroShowsTheSensorNotAFingerstick checks the tile shows the current glucose, the newest sensor reading,
// with its arrow, and a fingerstick newer than it under it, not in its place; with only fingersticks the current
// glucose is the newest of them, and none shows under it
func TestGlucoseHeroShowsTheSensorNotAFingerstick(t *testing.T) {
	mmol, now := consts.UnitsMMOLL, time.Now()
	reading := func(value float64, readingType string, ago time.Duration, direction *string) *sitemodels.SiteCleanGlucoseReading {
		return &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: value, Units: &mmol, ReadingType: readingType, OccurredAt: now.Add(-ago).UnixMilli(), Direction: direction}}
	}
	falling := "SingleDown"
	render := func(current *sitemodels.SiteCleanGlucoseReading, readings ...*sitemodels.SiteCleanGlucoseReading) string {
		summary := &sitemodels.Summary{Units: &mmol, RangeLow: 3.9, RangeHigh: 10, CurrentGlucose: current, GlucoseReadings: readings}
		var buf bytes.Buffer
		require.NoError(t, glucoseHero(summary).Render(context.Background(), &buf))
		return buf.String()
	}

	sensor := reading(4.7, "sgv", 3*time.Minute, &falling)
	html := render(sensor, reading(6.2, "mbg", time.Minute, nil), sensor, reading(5.0, "sgv", 8*time.Minute, nil))
	require.Regexp(t, `id="site--summary--glucose--reading"[^>]*>4.7</span>`, html)
	require.Contains(t, html, "↓")
	require.Regexp(t, `id="site--summary--glucose--meter"[^>]*>\s*Meter 6.2 mmol/L, `, html)

	sensor = reading(4.7, "sgv", time.Minute, &falling)
	html = render(sensor, sensor, reading(6.2, "mbg", 3*time.Minute, nil))
	require.Regexp(t, `id="site--summary--glucose--reading"[^>]*>4.7</span>`, html)
	require.NotContains(t, html, "site--summary--glucose--meter", "an older fingerstick is not shown")

	meter := reading(6.2, "mbg", time.Minute, nil)
	html = render(meter, meter)
	require.Regexp(t, `id="site--summary--glucose--reading"[^>]*>6.2</span>`, html)
	require.NotContains(t, html, "site--summary--glucose--meter")
}
