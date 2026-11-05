package site_summary_templ

import (
	"bytes"
	"context"
	"testing"

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
		{"high in mmol/L", 11.2, &mmol, "text-glucose-high", "High"},
		{"low in mg/dL", 63, &mgdl, "text-glucose-low", "Low"},
		{"at the low alarm in mg/dL", 70, &mgdl, "text-glucose-in-range", ""},
		{"at the high alarm in mg/dL", 180, &mgdl, "text-glucose-in-range", ""},
		{"high in mg/dL", 181, &mgdl, "text-glucose-high", "High"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The alarms are 70 and 180 mg/dL, as the summary has them in its units
			summary := &sitemodels.Summary{Units: tt.units, RangeLow: tt.units.FromMGDL(70), RangeHigh: tt.units.FromMGDL(180), GlucoseReadings: []*sitemodels.SiteCleanGlucoseReading{{SiteGlucoseReading: sitemodels.SiteGlucoseReading{Reading: tt.reading, Units: tt.units}}}}
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

func TestNumbersShowUnitsSmallerAndGrey(t *testing.T) {
	const unit = `<span class="text-sm font-medium text-gray-500">`
	tests := []struct{ text, html string }{
		{"0.7 U", "0.7" + unit + " U</span>"},
		{"84%", "84" + unit + "%</span>"},
		{"2 days 1 hour ago", "2" + unit + " days </span>1" + unit + " hour ago</span>"},
		{"just now", "just now"},
		{"–", "–"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		require.NoError(t, numbers(tt.text, "text-sm font-medium text-gray-500").Render(context.Background(), &buf))
		require.Equal(t, tt.html, buf.String(), tt.text)
	}
}
