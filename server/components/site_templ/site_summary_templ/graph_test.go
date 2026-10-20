package site_summary_templ

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

func TestGraph(t *testing.T) {
	mmol := consts.UnitsMMOLL
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.Local).UnixMilli()
	since := now - 48*time.Hour.Milliseconds()
	rate := 2.0
	g := newGraph(&sitemodels.Summary{
		Since: since, Units: &mmol,
		GlucoseReadings: []*sitemodels.SiteGlucoseReading{{Reading: 20, Units: &mmol}},
		InsulinDoses:    []*sitemodels.SiteInsulinDose{{Dose: 4}},
		TempBasals:      []*sitemodels.SiteTempBasal{{Absolute: &rate, OccurredAt: now - time.Hour.Milliseconds(), DurationMinutes: 120}},
	}, now)

	// Times map across the graph, and times outside it are kept on its edges
	require.Equal(t, graphLeft, g.x(since))
	require.Equal(t, graphRight, g.x(now))
	require.Equal(t, (graphLeft+graphRight)/2, g.x(since+24*time.Hour.Milliseconds()))
	require.Equal(t, graphLeft, g.x(since-time.Hour.Milliseconds()))

	// The glucose axis grows to the highest reading, and values above it are kept on top
	require.Equal(t, glucoseTop, g.glucoseY(20))
	require.Equal(t, glucoseBase, g.glucoseY(0))
	require.Equal(t, glucoseTop, g.glucoseY(30))

	// Readings are coloured by the target range, in the summary's units
	tests := []struct {
		reading float64
		colour  string
	}{
		{3.0, "fill-red-600"},
		{6.0, "fill-brand-night"},
		{12.0, "fill-amber-500"},
	}
	for _, test := range tests {
		require.Equal(t, test.colour, g.glucoseColour(test.reading), test.reading)
	}

	// The biggest dose and rate fill their strips, and a temp basal running past now ends at the graph's edge
	require.Equal(t, insulinTop, g.insulinY(4))
	require.Equal(t, "M970.6 340.0 V300.0 H990.0 V340.0 Z", g.basalPath([]*sitemodels.SiteTempBasal{{Absolute: &rate, OccurredAt: now - time.Hour.Milliseconds(), DurationMinutes: 120}}))

	// 48 hours is labelled every 6 hours
	ticks := g.ticks()
	require.Len(t, ticks, 7)
	require.Equal(t, "18:00", ticks[0].Label)
}
