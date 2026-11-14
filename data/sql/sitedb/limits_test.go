package sitedb_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// Every way in writes through the site DB, so its limits refuse values far outside real use, such as a 1000 unit bolus
// or 10000 g of carbs, while still taking negative insulin and food, which take back too much. Each column's least and
// most are written, and just past them, so the limits values are checked by before they are written are its CHECKs.
func TestSiteDataLimits(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, now, id, by := site.SiteDB.Queries(), time.Now().UnixMilli(), strconv.Itoa, "test"
	// Only an update writes who updated a record, and its limit is the creator's
	const updatedBy = "UpdatedByIdentityName"
	testutils.CheckColumnLimits(t, sitedb.ColumnLimits, map[string]testutils.LimitRecipe{
		"activities": testutils.Recipe(func(i int) sitedb.CreateActivityParams {
			return sitedb.CreateActivityParams{ID: id(i), ActivityType: "exercise", OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateActivity), updatedBy),
		"device_reports": testutils.Recipe(func(i int) sitedb.CreateDeviceReportParams {
			return sitedb.CreateDeviceReportParams{ID: id(i), OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateDeviceReport), updatedBy),
		"event_groups": testutils.Recipe(func(i int) sitedb.CreateEventGroupParams {
			return sitedb.CreateEventGroupParams{ID: id(i), GroupType: "note", OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateEventGroup), updatedBy),
		// Negative carbs, fat, protein and the rest take back too much
		"food_intakes": testutils.Recipe(func(i int) sitedb.CreateFoodIntakeParams {
			return sitedb.CreateFoodIntakeParams{ID: id(i), OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateFoodIntake), updatedBy),
		"food_presets": testutils.Recipe(func(i int) sitedb.CreateFoodPresetParams {
			return sitedb.CreateFoodPresetParams{ID: id(i), Name: id(i), CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateFoodPreset), updatedBy),
		"general_settings": testutils.Recipe(func(i int) sitedb.UpsertGeneralSettingParams {
			return sitedb.UpsertGeneralSettingParams{Key: id(i), Type: "aaps", ValueJson: "{}", CreatedByIdentityName: by}
		}, testutils.NoResult(q.UpsertGeneralSetting)),
		// A reading of 0 mg/dL is none, though one just above is kept
		"glucose_readings": testutils.Recipe(func(i int) sitedb.CreateGlucoseReadingParams {
			return sitedb.CreateGlucoseReadingParams{ID: id(i), Mgdl: 100, ReadingType: "sgv", OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateGlucoseReading), updatedBy),
		// Negative insulin takes back too much, off IOB
		"insulin_doses": testutils.Recipe(func(i int) sitedb.CreateInsulinDoseParams {
			return sitedb.CreateInsulinDoseParams{ID: id(i), Dose: 1, OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateInsulinDose), updatedBy),
		// A set of settings is kept once, seen first and last when it is written
		"loop_settings": testutils.Recipe(func(i int) sitedb.UpsertLoopSettingsParams {
			return sitedb.UpsertLoopSettingsParams{ID: fmt.Sprintf("%064d", i), App: "loop", Source: "profile", SettingsJson: "{}", SeenAt: now}
		}, q.UpsertLoopSettings).As("FirstSeenAt", "SeenAt").As("LastSeenAt", "SeenAt"),
		"notification_events": testutils.Recipe(func(i int) sitedb.CreateNotificationEventParams {
			return sitedb.CreateNotificationEventParams{ID: id(i), Title: "title", NotificationType: "note", OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateNotificationEvent), updatedBy),
		"profile_presets": testutils.Recipe(func(i int) sitedb.CreateProfilePresetParams {
			return sitedb.CreateProfilePresetParams{ID: id(i), Name: id(i), ProfileStoreID: "store", OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateProfilePreset), updatedBy),
		"profile_switches": testutils.Recipe(func(i int) sitedb.CreateProfileSwitchParams {
			return sitedb.CreateProfileSwitchParams{ID: id(i), Name: id(i), OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateProfileSwitch), updatedBy),
		"summary_period": testutils.Recipe(func(i int) sitedb.UpsertSummaryPeriodParams {
			return sitedb.UpsertSummaryPeriodParams{Period: "day", Date: fmt.Sprintf("%010d", i), Timezone: "UTC", Days: 1,
				GlucoseMgdlJson: "[]", LoopIobJson: "[]", LoopIobBasalJson: "[]", LoopIobTreatmentJson: "[]", LoopCobJson: "[]"}
		}, testutils.NoResult(q.UpsertSummaryPeriod)),
		"temp_basals": testutils.Recipe(func(i int) sitedb.CreateTempBasalParams {
			return sitedb.CreateTempBasalParams{ID: id(i), DurationMinutes: 30, Rate: 1, OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateTempBasal), updatedBy),
		// A temp target lasts at most a week
		"temp_targets": testutils.Recipe(func(i int) sitedb.CreateTempTargetParams {
			return sitedb.CreateTempTargetParams{ID: id(i), DurationMinutes: 30, HighTargetMgdl: 140, LowTargetMgdl: 120, OccurredAt: now, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreateTempTarget), updatedBy),
	})
}
