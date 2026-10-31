package sitedb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A profile preset's schedules are JSON, so anything else is refused when written
func TestProfilePresetSchedulesMustBeJSON(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	notJSON, schedule := "not json", `[{"time":"00:00","value":10}]`
	tests := []struct {
		name string
		set  func(p *sitedb.CreateProfilePresetParams, v *string)
	}{
		{"carb ratios", func(p *sitedb.CreateProfilePresetParams, v *string) { p.CarbRatiosJson = v }},
		{"insulin sensitivities", func(p *sitedb.CreateProfilePresetParams, v *string) { p.InsulinSensitivitiesJson = v }},
		{"basal rates", func(p *sitedb.CreateProfilePresetParams, v *string) { p.BasalRatesJson = v }},
		{"targets high", func(p *sitedb.CreateProfilePresetParams, v *string) { p.TargetsHighJson = v }},
		{"targets low", func(p *sitedb.CreateProfilePresetParams, v *string) { p.TargetsLowJson = v }},
	}
	for _, test := range tests {
		params := func(id string, v *string) sitedb.CreateProfilePresetParams {
			p := sitedb.CreateProfilePresetParams{ID: id, Name: id, ProfileStoreID: "store", CreatedByIdentityName: "test"}
			test.set(&p, v)
			return p
		}
		_, err := site.SiteDB.Queries.CreateProfilePreset(t.Context(), params("bad "+test.name, &notJSON))
		require.ErrorContains(t, err, "CHECK constraint failed", test.name)
		_, err = site.SiteDB.Queries.CreateProfilePreset(t.Context(), params("good "+test.name, &schedule))
		require.NoError(t, err, test.name)
	}
}
