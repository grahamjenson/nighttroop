package sitedata

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A summary's basal and targets read a profile sent without its own timezone in the site's, whatever the timezone of
// whoever views it, as the daily summaries do
func TestSummaryMinutesInSiteTimezone(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx := site.SiteDB.Queries(), t.Context()
	// Each hour's basal is its hour of the day
	schedule := []map[string]any{}
	for hour := range 24 {
		schedule = append(schedule, map[string]any{"time": fmt.Sprintf("%02d:00", hour), "timeAsSeconds": hour * 3600, "value": hour})
	}
	basal, err := json.Marshal(schedule)
	require.NoError(t, err)
	now := time.Now()
	_, err = q.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: "default", Name: "Default", ProfileStoreID: "store", IsDefault: true,
		BasalRatesJson: new(string(basal)), OccurredAt: now.Add(-48 * time.Hour).UnixMilli(), CreatedByIdentityName: "loop"})
	require.NoError(t, err)

	siteContext := &httputils.SiteContext{SiteID: "site", SiteModel: &controldb.Site{ID: "site", Timezone: new("Pacific/Auckland")},
		IdentityModel: &controldb.SiteIdentity{Name: "viewer", Timezone: new("America/Los_Angeles")}}
	summary := &sitemodels.Summary{Since: now.Add(-10 * time.Minute).UnixMilli(), Units: new(consts.UnitsMGDL)}
	require.Nil(t, addMinutes(ctx, q, summary, siteContext))

	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	require.NotEmpty(t, summary.Minutes)
	require.Equal(t, float64(time.UnixMilli(summary.Since).In(auckland).Hour()), *summary.Minutes[0].Basal)
}
