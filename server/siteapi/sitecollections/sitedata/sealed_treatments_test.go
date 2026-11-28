package sitedata

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A day's treatments are sealed as series every 5 minutes: the basal rate in force, from the profile and the temp
// basals, the insulin of its boluses, an extended one as it is given and over midnight, the basal insulin injected,
// and the carbs. Their totals are the day's.
func TestSealingTreatments(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	day := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -3)
	at := func(hour, minute int) int64 {
		return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute).UnixMilli()
	}
	// A profile of 1.2 U/hr all day. A temp basal of 0 for 30 minutes from 01:00, and one of 2.4 U/hr for 5 minutes
	// from 02:02, which is not on a step.
	_, err := q.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: "default", Name: "Default", ProfileStoreID: "store", IsDefault: true, Timezone: new("UTC"),
		BasalRatesJson: new(`[{"time":"00:00","timeAsSeconds":0,"value":1.2}]`), OccurredAt: day.AddDate(0, 0, -10).UnixMilli(), CreatedByIdentityName: "loop"})
	require.NoError(t, err)
	for id, basal := range map[string]sitedb.CreateTempBasalParams{
		"zero": {DurationMinutes: 30, Rate: new(0.0), OccurredAt: at(1, 0)}, "high": {DurationMinutes: 5, Rate: new(2.4), OccurredAt: at(2, 2)}} {
		basal.ID, basal.CreatedByIdentityName = id, "loop"
		_, err := q.CreateTempBasal(ctx, basal)
		require.NoError(t, err)
	}
	// A loop's 0.5 U and a person's 2 U in the same 5 minutes; 1 U with 2 U more over the hour from 23:30, half of it
	// the day after; and 10 U of long acting insulin
	for id, dose := range map[string]sitedb.CreateInsulinDoseParams{
		"smb": {Dose: 0.5, Automatic: new(true), OccurredAt: at(3, 1)}, "meal": {Dose: 2, OccurredAt: at(3, 3)},
		"extended": {Dose: 1, ExtendedDose: new(2.0), ExtendedDurationMs: new(time.Hour.Milliseconds()), OccurredAt: at(23, 30)},
		"lantus":   {Dose: 10, IsBasalInsulin: new(true), OccurredAt: at(8, 0)}} {
		dose.ID, dose.CreatedByIdentityName = id, "loop"
		_, err := q.CreateInsulinDose(ctx, dose)
		require.NoError(t, err)
	}
	for id, food := range map[string]sitedb.CreateFoodIntakeParams{"lunch": {Carbs: new(40.0), OccurredAt: at(12, 0)}, "more": {Carbs: new(15.5), OccurredAt: at(12, 2)}} {
		food.ID, food.CreatedByIdentityName = id, "phone"
		_, err := q.CreateFoodIntake(ctx, food)
		require.NoError(t, err)
	}
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)

	// series is a day of a series: its values, how many steps have one, and their total
	series := func(kind string, of time.Time) ([]float64, int, float64) {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: kind, Date: of.Format(time.DateOnly)})
		if err != nil {
			return nil, 0, 0
		}
		values, err := retention.UnpackGrid(row.Body)
		require.NoError(t, err)
		require.Len(t, values, 288)
		with, total := 0, 0.0
		for _, value := range values {
			if !math.IsNaN(value) {
				with, total = with+1, total+value
			}
		}
		return values, with, total
	}

	// The basal rate each 5 minutes: the profile's, 0 through the half hour, and the mean of a step's minutes where a
	// temp basal starts in it. The day's basal is what was given: 24 h of 1.2, less 0.6, and 0.1 more.
	rate, with, total := series("basal_rate", day)
	require.Equal(t, 288, with)
	require.Equal(t, 1.2, rate[0])
	require.Equal(t, []float64{0, 0, 0, 0, 0, 0, 1.2}, rate[12:19])
	require.Greater(t, rate[24], 1.2)
	require.Less(t, rate[24], 2.4)
	require.InDelta(t, 28.3, total/12, 0.001)

	// The boluses, a loop's and a person's together, and the extended one as it is given, with its half the day after
	bolus, with, total := series("insulin_bolus", day)
	require.Equal(t, 2.5, bolus[36])
	require.InDelta(t, 1+2.0/12, bolus[282], 0.001)
	require.InDelta(t, 2.0/12, bolus[287], 0.001)
	require.Equal(t, 7, with)
	require.InDelta(t, 4.5, total, 0.01)
	after, with, total := series("insulin_bolus", day.AddDate(0, 0, 1))
	require.Equal(t, 6, with)
	require.InDelta(t, 1.0, total, 0.01)
	require.True(t, math.IsNaN(after[6]))

	// The long acting insulin is basal, not a bolus, and the carbs of a step are added
	injected, with, _ := series("insulin_basal", day)
	require.Equal(t, 1, with)
	require.Equal(t, 10.0, injected[96])
	carbs, with, _ := series("carbs", day)
	require.Equal(t, 1, with)
	require.Equal(t, 55.5, carbs[144])

	// The day after has the profile's basal rate, though no treatment was written in it, and no carbs
	_, with, total = series("basal_rate", day.AddDate(0, 0, 1))
	require.Equal(t, 288, with)
	require.InDelta(t, 28.8, total/12, 0.001)
	values, _, _ := series("carbs", day.AddDate(0, 0, 1))
	require.Nil(t, values)

	// A dose deleted empties its step when the day is sealed again
	_, err = q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{ID: "lantus"})
	require.NoError(t, err)
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	_, with, _ = series("insulin_basal", day)
	require.Zero(t, with)
}
