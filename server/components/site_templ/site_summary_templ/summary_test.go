package site_summary_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

// TestGlucoseHeroColoursItsRange checks the latest glucose is red and says Low below the range, amber and says High
// above it, and is the brand colour in it, in either units
func TestGlucoseHeroColoursItsRange(t *testing.T) {
	mmol, mgdl := consts.UnitsMMOLL, consts.UnitsMGDL
	tests := []struct {
		name    string
		reading float64
		units   *consts.Units
		colour  string
		label   string
	}{
		{"low in mmol/L", 3.5, &mmol, "text-red-600", "Low"},
		{"in range in mmol/L", 6.5, &mmol, "text-brand-night", ""},
		{"high in mmol/L", 11.2, &mmol, "text-amber-500", "High"},
		{"low in mg/dL", 63, &mgdl, "text-red-600", "Low"},
		{"the bottom of the range in mg/dL", 70, &mgdl, "text-brand-night", ""},
		{"the top of the range in mg/dL", 180, &mgdl, "text-brand-night", ""},
		{"high in mg/dL", 181, &mgdl, "text-amber-500", "High"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &sitemodels.Summary{Units: tt.units, GlucoseReadings: []*sitemodels.SiteGlucoseReading{{Reading: tt.reading, Units: tt.units}}}
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
