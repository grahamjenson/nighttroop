package sitedata

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
	"nighttroop.t1d.nz/utils/testutils"
)

// A day's summary built from its cold days is the one built from its rows, for a site whose day is not a UTC
// day
func TestSummaryDayFromColdDays(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	siteContext := &httputils.SiteContext{SiteID: "site", SiteModel: &controldb.Site{ID: "site", SettingsJson: settings.StoredOf(settings.Settings{Timezone: new("Pacific/Auckland")})}}
	// The local day that started 4 midnights ago in Auckland is in two UTC days, both over
	now := time.Now().In(auckland)
	day := time.Date(now.Year(), now.Month(), now.Day()-4, 0, 0, 0, 0, auckland)
	date := day.Format(time.DateOnly)
	at := func(hour, minute int) int64 {
		return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute).UnixMilli()
	}
	must := func(_ any, err error) { require.NoError(t, err) }
	must(q.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: "default", Name: "Default", ProfileStoreID: "store", IsDefault: true, Timezone: new("UTC"),
		BasalRatesJson: new(`[{"time":"00:00","timeAsSeconds":0,"value":1.2}]`), OccurredAt: day.AddDate(0, 0, -10).UnixMilli(), CreatedByIdentityName: "loop"}))
	// A loop's report every 5 minutes, and a CGM reading with each but for the hour from 10:00, from very low to very
	// high, and a reading the day before and the day after, which are not the day's
	for slot := -1; slot <= 288; slot++ {
		when := at(0, slot*5) + 60_000
		if slot < 120 || slot >= 132 {
			must(q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprintf("cgm-%d", slot), Mgdl: float64(40 + slot), ReadingType: "sgv", OccurredAt: when, CreatedByIdentityName: "phone"}))
		}
		must(q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprintf("report-%d", slot), OccurredAt: when, CreatedByIdentityName: "phone", CreatedByDeviceName: new("loop"),
			LoopIob: new(2 + float64(slot)/100), LoopIobBasal: new(0.5), LoopIobTreatment: new(1.5 + float64(slot)/100), LoopCob: new(float64(slot % 30))}))
	}
	must(q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "meter", Mgdl: 111, ReadingType: "mbg", OccurredAt: at(10, 30), CreatedByIdentityName: "phone"}))
	must(q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: "zero", DurationMinutes: 30, Rate: new(0.0), OccurredAt: at(1, 0), CreatedByIdentityName: "loop"}))
	// A bolus, one extended over midnight into the day, long acting insulin, and two meals
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "meal", Dose: 4.5, OccurredAt: at(12, 3), CreatedByIdentityName: "loop"}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "extended", Dose: 1, ExtendedDose: new(2.0), ExtendedDurationMs: new(time.Hour.Milliseconds()), OccurredAt: at(-1, 30), CreatedByIdentityName: "loop"}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "lantus", Dose: 10, IsBasalInsulin: new(true), OccurredAt: at(8, 0), CreatedByIdentityName: "loop"}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "lunch", Carbs: new(40.0), OccurredAt: at(12, 0), CreatedByIdentityName: "phone"}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "tea", Carbs: new(15.5), OccurredAt: at(23, 58), CreatedByIdentityName: "phone"}))

	rows, err := buildSummaryDaily(ctx, q, siteContext, date, time.Now().UnixMilli())
	require.NoError(t, err)
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	// Before the site is packaged, and where it has data written since, the day is built from its rows
	days, err := builtSummaryDays(ctx, q, siteContext, date, date, time.Now(), false)
	require.NoError(t, err)
	require.Equal(t, rows.GlucoseMgdlJson, days[0].GlucoseMgdlJson)
	siteContext.SiteModel.BackedUpAt = new(time.Now().Add(2 * time.Hour).UnixMilli())
	days, err = builtSummaryDays(ctx, q, siteContext, date, date, time.Now(), false)
	require.NoError(t, err)
	require.Len(t, days, 1)
	cold := days[0]

	// The meter's reading is not in the glucose line, and counts as a reading but not in the mean: the rest is the same
	require.Equal(t, int64(277), rows.GlucoseReadings)
	require.Equal(t, rows.GlucoseReadings, cold.GlucoseReadings)
	require.Equal(t, int64(1), cold.MeterReadings)
	require.Equal(t, int64(276), cold.CgmSlots)
	require.Equal(t, rows.CgmSlots, cold.CgmSlots)
	require.InDelta(t, *rows.GlucoseMeanMgdl, *cold.GlucoseMeanMgdl, 0.5)
	require.InDelta(t, *rows.GlucoseSdMgdl, *cold.GlucoseSdMgdl, 0.5)
	for i, column := range [](*float64){rows.TimeVeryLow, rows.TimeLow, rows.TimeInRange, rows.TimeHigh, rows.TimeVeryHigh} {
		require.InDelta(t, *column, *[](*float64){cold.TimeVeryLow, cold.TimeLow, cold.TimeInRange, cold.TimeHigh, cold.TimeVeryHigh}[i], 0.4, i)
	}
	require.Equal(t, rows.TotalDailyCarbs, cold.TotalDailyCarbs)
	require.InDelta(t, rows.TotalDailyDose, cold.TotalDailyDose, 0.011)
	require.InDelta(t, rows.TotalDailyBasal, cold.TotalDailyBasal, 0.011)
	require.Greater(t, cold.TotalDailyBasal, 30.0)
	require.Equal(t, rows.Days, cold.Days)
	require.Equal(t, rows.LoopIobJson, cold.LoopIobJson)
	require.Equal(t, rows.LoopIobBasalJson, cold.LoopIobBasalJson)
	require.Equal(t, rows.LoopIobTreatmentJson, cold.LoopIobTreatmentJson)
	require.Equal(t, rows.LoopCobJson, cold.LoopCobJson)
	// The glucose line is the same but for the meter's 5 minutes
	require.Len(t, cold.GlucoseMgdlJson, len(rows.GlucoseMgdlJson)+len("null")-len("111"))
	require.Equal(t, rows.GlucoseMgdlJson[:400], cold.GlucoseMgdlJson[:400])

	// A day with no data is left out
	none, err := builtSummaryDays(ctx, q, siteContext, day.AddDate(0, 0, -20).Format(time.DateOnly), day.AddDate(0, 0, -19).Format(time.DateOnly), time.Now(), false)
	require.NoError(t, err)
	require.Empty(t, none)
}

// The Summaries page of a month has its days with data, and the weeks, fortnights and months that are over and have
// one, each built from the site's days as it is read
func TestBuiltSummaries(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	siteContext := &httputils.SiteContext{SiteID: "site", SiteModel: &controldb.Site{ID: "site", SettingsJson: settings.StoredOf(settings.Settings{Timezone: new("UTC")})}}
	// Now is Wednesday 18 March 2026. A reading in range on Monday 2 March, one high on Tuesday 10 March, one on
	// 15 February, and one today
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	for id, at := range map[string]time.Time{"feb": time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC), "first": time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),
		"second": time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC), "today": now.Add(-time.Hour)} {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: map[string]float64{"second": 250}[id] + 100, ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	built, err := BuiltSummaries(ctx, q, siteContext, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), now)
	require.NoError(t, err)
	dates := func(period string) (of []string) {
		for _, row := range built[period] {
			of = append(of, row.Date)
		}
		return of
	}
	// Newest first. Today, and the week, fortnight and month it is in, are not over.
	require.Equal(t, []string{"2026-03-10", "2026-03-02"}, dates("day"))
	require.Equal(t, []string{"2026-03-09", "2026-03-02"}, dates("week"))
	// Fortnights are from 2 March, with both days, and 16 March, which is not over
	require.Equal(t, []string{"2026-03-02"}, dates("fortnight"))
	require.Equal(t, 50.0, *built["fortnight"][0].TimeInRange)
	require.Equal(t, []string{"2026-02-01"}, dates("month"))
	require.Equal(t, 100.0, *built["week"][1].TimeInRange)
	require.Equal(t, 0.0, *built["week"][0].TimeInRange)
	require.Equal(t, int64(1), built["month"][0].Days)
}
