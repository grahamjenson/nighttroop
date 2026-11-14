package sitecollections

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestCheckLimits checks a site API write is refused for a value its column, narrowed by its override, does not take,
// saying the JSON field and what it takes, with glucose in the settings' units
func TestCheckLimits(t *testing.T) {
	mgdl := &httputils.SiteContext{SiteID: "site", SiteModel: &controldb.Site{ID: "site", Units: new(consts.UnitsMGDL)}}
	mmol := &httputils.SiteContext{SiteID: "site", SiteModel: &controldb.Site{ID: "site", Units: new(consts.UnitsMMOLL)}}
	tests := []struct {
		name        string
		collection  string
		model       any
		siteContext *httputils.SiteContext
		update      bool
		want        string
	}{
		{"in range", "insulin_doses", &sitemodels.SiteInsulinDose{Dose: 5, SplitNow: new(0.5)}, mgdl, false, ""},
		{"out of its column's range", "insulin_doses", &sitemodels.SiteInsulinDose{Dose: 5, SplitNow: new(1.5)}, mgdl, false,
			"split_now must be from 0 to 1"},
		{"out of its override, narrower than its column", "insulin_doses", &sitemodels.SiteInsulinDose{Dose: 200}, mgdl, false,
			"dose must be from -100 to 100"},
		{"a drift fixed: food's energy at least 0", "food", &sitemodels.SiteFoodIntake{Energy: new(-1.0)}, mgdl, false,
			"energy must be from 0 to 20000"},
		{"a drift fixed: a temp basal of at least 1%", "temp_basals", &sitemodels.SiteTempBasal{DurationMinutes: 30, Rate: new(1.0), Multiplier: new(0.001)}, mgdl, false,
			"multiplier must be from 0.01 to 11"},
		{"a zero created is checked", "temp_basals", &sitemodels.SiteTempBasal{Rate: new(1.0)}, mgdl, false,
			"duration_minutes must be from 1 to 1440"},
		{"a zero updated is not sent", "temp_basals", &sitemodels.SiteTempBasal{Rate: new(1.0)}, mgdl, true, ""},
		{"a nil is not sent", "activities", &sitemodels.SiteActivity{}, mgdl, false, ""},
		{"too long", "notification_events", &sitemodels.SiteNotificationEvent{Title: strings.Repeat("a", 129)}, mgdl, false,
			"title must be at most 128 characters"},
		{"a text of runes, not bytes", "notification_events", &sitemodels.SiteNotificationEvent{Title: strings.Repeat("é", 128)}, mgdl, false, ""},
		{"glucose in mg/dL", "glucose", &sitemodels.SiteGlucoseReading{Reading: 1001}, mgdl, false,
			"reading must be more than 0 and at most 1000 mg/dL"},
		{"glucose in mmol/L, rounded in", "glucose", &sitemodels.SiteGlucoseReading{Reading: 56}, mmol, false,
			"reading must be more than 0 and at most 55.5 mmol/L"},
		{"glucose in mmol/L in range", "glucose", &sitemodels.SiteGlucoseReading{Reading: 55.5}, mmol, false, ""},
		{"a glucose override in mmol/L", "temp_targets", &sitemodels.SiteTempTarget{DurationMinutes: 60, LowTarget: 0.001, HighTarget: 6}, mmol, false,
			"low_target must be from 0.1 to 55.5 mmol/L"},
		{"a collection with no table", "settings", &sitemodels.Settings{}, mgdl, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			herr := checkLimits(tt.collection, tt.model, tt.siteContext, tt.update)
			if tt.want == "" {
				require.Nil(t, herr)
				return
			}
			require.NotNil(t, herr)
			require.Equal(t, 400, herr.Status)
			require.Equal(t, "Constraint check failed: "+tt.want, herr.Message)
		})
	}
}
