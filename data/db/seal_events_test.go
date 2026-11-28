package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// What a person did in a day is sealed as one list, the oldest first: each with the columns its table's policy keeps
// that have a value, and what its treatment says of it. What a loop did is not in it: its doses and temp basals are
// series, as a CGM's readings are.
func TestSealingEvents(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	day := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -3)
	at := func(hour int) int64 { return day.Add(time.Duration(hour) * time.Hour).UnixMilli() }
	must := func(_ any, err error) { require.NoError(t, err) }
	by := "phone"
	// Events: a meter's reading, a CGM reading with a note, a meal with its treatment, a person's dose, a temp target,
	// a profile switch, an activity, a note, and a site change, which is a treatment with no event
	must(q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "meter", Mgdl: 101, ReadingType: "mbg", OccurredAt: at(1), CreatedByIdentityName: by}))
	must(q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "noted", Mgdl: 64, ReadingType: "sgv", Notes: new("felt low"), OccurredAt: at(2), CreatedByIdentityName: by}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "lunch", Carbs: new(40.0), Name: new("Pasta"), OccurredAt: at(3), CreatedByIdentityName: by, CreatedByAppName: new("AAPS")}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "lunch-group", GroupType: "Meal Bolus", FoodIntakeID: new("lunch"), EnteredBy: new("openaps://AAPS"), Notes: new("from the group"), OccurredAt: at(3), CreatedByIdentityName: by}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "bolus", Dose: 4.5, OccurredAt: at(4), CreatedByIdentityName: by}))
	must(q.CreateTempTarget(ctx, sitedb.CreateTempTargetParams{ID: "target", LowTargetMgdl: 140, HighTargetMgdl: 160, DurationMinutes: 60, Reason: new("Exercise"), OccurredAt: at(5), CreatedByIdentityName: by}))
	must(q.CreateProfileSwitch(ctx, sitedb.CreateProfileSwitchParams{ID: "switch", Name: "Sick day", Multiplier: new(1.2), OccurredAt: at(6), CreatedByIdentityName: by,
		BasalRatesJson: new(`[{"time":"00:00","value":1}]`)}))
	must(q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: "run", ActivityType: "run", DurationMinutes: new(int64(30)), OccurredAt: at(7), CreatedByIdentityName: by}))
	must(q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: "note", Title: "Announcement", NotificationType: "note", OccurredAt: at(8), CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "site-change", GroupType: "Site Change", EnteredBy: new("xDrip"), OccurredAt: at(9), CreatedByIdentityName: by}))
	// Not events: a CGM reading, a loop's dose and its treatment, a temp basal and its treatment, and a deleted meal
	must(q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "cgm", Mgdl: 120, ReadingType: "sgv", OccurredAt: at(10), CreatedByIdentityName: by}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "smb", Dose: 0.2, Automatic: new(true), OccurredAt: at(11), CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "smb-group", GroupType: "Correction Bolus", InsulinDoseID: new("smb"), OccurredAt: at(11), CreatedByIdentityName: by}))
	must(q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: "temp", DurationMinutes: 30, Rate: new(0.5), OccurredAt: at(12), CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "temp-group", GroupType: "Temp Basal", TempBasalID: new("temp"), OccurredAt: at(12), CreatedByIdentityName: by}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "gone", Carbs: new(10.0), OccurredAt: at(13), CreatedByIdentityName: by}))
	must(q.SoftDeleteFoodIntake(ctx, sitedb.SoftDeleteFoodIntakeParams{ID: "gone"}))

	events := func() []map[string]any {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: day.Format(time.DateOnly)})
		require.NoError(t, err)
		require.Equal(t, retention.EventsShape, retention.ShapeOf(row.Body))
		header, list, err := retention.ReadEvents(row.Body)
		require.NoError(t, err)
		require.Equal(t, retention.Header{Version: 1, Shape: retention.EventsShape, Deflated: true, Steps: len(list)}, header)
		return list
	}
	seal := func() {
		_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
	}
	seal()
	sealed := events()
	ids := []any{}
	for _, event := range sealed {
		ids = append(ids, event["id"])
	}
	require.Equal(t, []any{"meter", "noted", "lunch", "bolus", "target", "switch", "run", "note", "site-change"}, ids)

	// An event has its table, and the columns kept that have a value: not the app that sent it, or a column with none
	ms := func(hour int) float64 { return float64(at(hour)) }
	require.Equal(t, map[string]any{"k": "glucose_readings", "id": "meter", "occurred_at": ms(1), "mgdl": 101.0, "reading_type": "mbg", "created_by_identity_name": by}, sealed[0])
	require.Equal(t, "felt low", sealed[1]["notes"])
	// The meal has what its treatment says of it, but keeps its own ID and has the treatment's notes, as it has none
	require.Equal(t, map[string]any{"k": "food_intakes", "id": "lunch", "occurred_at": ms(3), "carbs": 40.0, "name": "Pasta", "created_by_identity_name": by,
		"group_type": "Meal Bolus", "entered_by": "openaps://AAPS", "notes": "from the group"}, sealed[2])
	require.Equal(t, 4.5, sealed[3]["dose"])
	require.Equal(t, []any{140.0, 160.0, 60.0, "Exercise"}, []any{sealed[4]["low_target_mgdl"], sealed[4]["high_target_mgdl"], sealed[4]["duration_minutes"], sealed[4]["reason"]})
	// The switch has its name and percent, not the profile it was made with
	require.Equal(t, []any{"Sick day", 1.2}, []any{sealed[5]["name"], sealed[5]["multiplier"]})
	require.NotContains(t, sealed[5], "basal_rates_json")
	require.Equal(t, map[string]any{"k": "event_groups", "id": "site-change", "occurred_at": ms(9), "group_type": "Site Change", "entered_by": "xDrip", "created_by_identity_name": by}, sealed[8])

	// An event deleted is gone when its day is sealed again, and a day left with none is emptied
	must(q.SoftDeleteInsulinDose(ctx, sitedb.SoftDeleteInsulinDoseParams{ID: "bolus"}))
	seal()
	require.Len(t, events(), 8)
	_, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: day.AddDate(0, 0, 1).Format(time.DateOnly)})
	require.Error(t, err, "a day with no event has no list")
}

// Each column a policy keeps is a column of its table, so a name spelt wrong does not lose what it was to keep
func TestPoliciesKeepRealColumns(t *testing.T) {
	rows := map[string]any{"glucose_readings": sitedb.GlucoseReading{}, "insulin_doses": sitedb.InsulinDose{}, "food_intakes": sitedb.FoodIntake{}, "activities": sitedb.Activity{},
		"temp_targets": sitedb.TempTarget{}, "profile_switches": sitedb.ProfileSwitch{}, "notification_events": sitedb.NotificationEvent{}, "event_groups": sitedb.EventGroup{}}
	for table, policy := range retention.Policies {
		if len(policy.Keep) == 0 {
			continue
		}
		row, ok := rows[table]
		require.True(t, ok, "%s keeps columns, so its row is here", table)
		require.ElementsMatch(t, policy.Keep, testutils.ColumnsOf(table, row, policy.Keep), table)
	}
}
