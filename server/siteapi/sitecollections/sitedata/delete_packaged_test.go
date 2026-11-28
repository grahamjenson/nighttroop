package sitedata

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// Where the server deletes rows, the treatments and events older than the hot days are deleted once their day is
// packaged and has settled, and what they were is in the day's cold days. A profile switch that may still be in force
// is kept, a day whose rows are gone is not packaged again, and a row written for it later stays a row.
func TestDeletingPackagedTreatments(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	settle := db.PackagedSettle
	t.Cleanup(func() { cfg.RetentionDeleteRows, db.PackagedSettle = true, settle })
	// The server deletes rows unless it is told not to: this starts as one that does not
	cfg.RetentionDeleteRows = false
	ctx, q := t.Context(), site.SiteDB.Queries()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	old, recent := today.AddDate(0, 0, -40), today.AddDate(0, 0, -2)
	must := func(_ any, err error) { require.NoError(t, err) }
	by := "phone"
	must(q.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: "default", Name: "Default", ProfileStoreID: "store", IsDefault: true, Timezone: new("UTC"),
		BasalRatesJson: new(`[{"time":"00:00","timeAsSeconds":0,"value":1}]`), OccurredAt: old.AddDate(0, 0, -30).UnixMilli(), CreatedByIdentityName: by}))
	// Profile switches: one with no end 20 days before the old day, replaced by one 10 days before it, which is still
	// in force; one for an hour on the old day, long over; and one for 60 days from the old day, not over
	basal := func(rate string) *string { return new(`[{"time":"00:00","timeAsSeconds":0,"value":` + rate + `}]`) }
	for id, s := range map[string]sitedb.CreateProfileSwitchParams{
		"replaced": {Name: "First", OccurredAt: old.AddDate(0, 0, -20).UnixMilli(), BasalRatesJson: basal("1.1")},
		"in force": {Name: "Second", OccurredAt: old.AddDate(0, 0, -10).UnixMilli(), BasalRatesJson: basal("1.2")},
		"over":     {Name: "Hour", DurationMinutes: 60, OccurredAt: old.Add(time.Hour).UnixMilli(), BasalRatesJson: basal("2")},
		"running":  {Name: "Two months", DurationMinutes: 60 * 24 * 60, OccurredAt: old.Add(2 * time.Hour).UnixMilli(), BasalRatesJson: basal("1.5")}} {
		s.ID, s.CreatedByIdentityName = id, by
		must(q.CreateProfileSwitch(ctx, s))
	}
	// The old day: a meal and a person's bolus in one treatment, a loop's dose and a temp basal with theirs, a temp
	// target, an activity and a note. The recent day: a meal.
	at := old.Add(5 * time.Hour).UnixMilli()
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "meal", Carbs: new(40.0), OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "bolus", Dose: 4, OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "meal-group", GroupType: "Meal Bolus", FoodIntakeID: new("meal"), InsulinDoseID: new("bolus"), OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "smb", Dose: 0.3, Automatic: new(true), OccurredAt: at + 600_000, CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "smb-group", GroupType: "Correction Bolus", InsulinDoseID: new("smb"), OccurredAt: at + 600_000, CreatedByIdentityName: by}))
	must(q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: "temp", DurationMinutes: 30, Rate: new(0.0), OccurredAt: old.Add(8 * time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "temp-group", GroupType: "Temp Basal", TempBasalID: new("temp"), OccurredAt: old.Add(8 * time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	must(q.CreateTempTarget(ctx, sitedb.CreateTempTargetParams{ID: "target", LowTargetMgdl: 140, HighTargetMgdl: 160, DurationMinutes: 60, OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: "run", ActivityType: "run", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: "note", Title: "Note", NotificationType: "note", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "recent meal", Carbs: new(20.0), OccurredAt: recent.Add(time.Hour).UnixMilli(), CreatedByIdentityName: by}))

	// rows are how many rows each table has
	rows := func() map[string]int {
		counts := map[string]int{}
		count := func(table string, n int, err error) {
			require.NoError(t, err)
			counts[table] = n
		}
		groups, err := q.AllEventGroups(ctx, sitedb.AllEventGroupsParams{Limit: 100})
		count("event_groups", len(groups), err)
		doses, err := q.AllInsulinDoses(ctx, sitedb.AllInsulinDosesParams{Limit: 100})
		count("insulin_doses", len(doses), err)
		food, err := q.AllFoodIntakes(ctx, sitedb.AllFoodIntakesParams{Limit: 100})
		count("food_intakes", len(food), err)
		basals, err := q.AllTempBasals(ctx, sitedb.AllTempBasalsParams{Limit: 100})
		count("temp_basals", len(basals), err)
		targets, err := q.AllTempTargets(ctx, sitedb.AllTempTargetsParams{Limit: 100})
		count("temp_targets", len(targets), err)
		activities, err := q.AllActivities(ctx, sitedb.AllActivitiesParams{Limit: 100})
		count("activities", len(activities), err)
		notes, err := q.AllNotificationEvents(ctx, sitedb.AllNotificationEventsParams{Limit: 100})
		count("notification_events", len(notes), err)
		switches, err := q.AllProfileSwitches(ctx, sitedb.AllProfileSwitchesParams{Limit: 100})
		count("profile_switches", len(switches), err)
		return counts
	}
	all := map[string]int{"event_groups": 3, "insulin_doses": 2, "food_intakes": 2, "temp_basals": 1, "temp_targets": 1, "activities": 1, "notification_events": 1, "profile_switches": 4}
	require.Equal(t, all, rows())
	packageDays := func() {
		_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
	}
	total := func(kind string, of time.Time) float64 {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: kind, Date: of.Format(time.DateOnly)})
		require.NoError(t, err, kind)
		values, err := retention.UnpackGrid(row.Body)
		require.NoError(t, err)
		sum := 0.0
		for _, value := range values {
			if !math.IsNaN(value) {
				sum += value
			}
		}
		return sum
	}
	events := func(of time.Time) retention.EventsDay {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: of.Format(time.DateOnly)})
		require.NoError(t, err)
		_, day, err := retention.ReadEvents(row.Body)
		require.NoError(t, err)
		return day
	}

	// A server that does not delete rows packages the days and deletes none
	packageDays()
	require.Equal(t, all, rows())

	// A server that deletes rows does not until a day has settled, a day here
	cfg.RetentionDeleteRows = true
	packageDays()
	require.Equal(t, all, rows())

	// Settled, the old day's treatments and events are deleted, each with its treatment. The recent meal is a hot row.
	// Of the profile switches, the replaced one and the one long over go; the newest with no end and the one still
	// running stay, as the basal of the days that are rows is read from them.
	db.PackagedSettle = 0
	packageDays()
	require.Equal(t, map[string]int{"event_groups": 0, "insulin_doses": 0, "food_intakes": 1, "temp_basals": 0, "temp_targets": 0, "activities": 0, "notification_events": 0, "profile_switches": 2}, rows())
	kept, err := q.AllProfileSwitches(ctx, sitedb.AllProfileSwitchesParams{Limit: 100})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"in force", "running"}, []string{kept[0].ID, kept[1].ID})

	// What they were is in the old day's cold days: its carbs, its boluses, a loop's and a person's, its basal with
	// the half hour at 0, and its events, with the switch that was in force when it started
	require.Equal(t, 40.0, total("carbs", old))
	require.InDelta(t, 4.3, total("insulin_bolus", old), 0.001)
	require.Less(t, total("basal_rate", old)/12, 24*1.5)
	require.Greater(t, total("basal_rate", old)/12, 24.0)
	day := events(old)
	ids := []any{}
	for _, event := range day.Events {
		ids = append(ids, event["id"])
	}
	require.ElementsMatch(t, []any{"over", "running", "meal", "bolus", "target", "run", "note"}, ids)
	require.Equal(t, "in force", day.InForce[0]["id"])
	// The recent day still reads the switch in force from its row
	require.Equal(t, "running", events(recent).InForce[len(events(recent).InForce)-1]["id"])

	// A day whose rows are gone is not packaged again: a meal written for it later stays a row, and the day is as it
	// was, so nothing the day had is lost to a day built from one row
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "late", Carbs: new(15.0), OccurredAt: at + 1000, CreatedByIdentityName: by}))
	packageDays()
	packageDays()
	require.Equal(t, 2, rows()["food_intakes"])
	require.Equal(t, 40.0, total("carbs", old))
	require.Len(t, events(old).Events, 7)
}

// Truncating a site again before its old days have settled, as an admin may, packages them again unchanged, which does
// not restart their wait, so their rows are still deleted once they settle
func TestTruncatingAgainDoesNotHoldBackDeletingPackagedRows(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	settle := db.PackagedSettle
	t.Cleanup(func() { db.PackagedSettle = settle })
	db.PackagedSettle = 200 * time.Millisecond
	ctx, q := t.Context(), site.SiteDB.Queries()
	old := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -40)
	_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "bolus", Dose: 4, OccurredAt: old.Add(5 * time.Hour).UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	doses := func() int {
		_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
		rows, err := q.AllInsulinDoses(ctx, sitedb.AllInsulinDosesParams{Limit: 10})
		require.NoError(t, err)
		return len(rows)
	}

	require.Equal(t, 1, doses(), "packaged, and not yet settled")
	time.Sleep(120 * time.Millisecond)
	require.Equal(t, 1, doses(), "truncated again before it settled")
	time.Sleep(120 * time.Millisecond)
	require.Equal(t, 0, doses(), "settled since it was first packaged")
}

// A day with rows but no event, as one whose only meal was deleted, on a site with no profile, is packaged too, with
// no events, so its rows are deleted: a day with no cold day of events kept its rows for ever
func TestADayWithNoEventsIsPackagedAndItsRowsDeleted(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	settle := db.PackagedSettle
	db.PackagedSettle = 0
	t.Cleanup(func() { db.PackagedSettle = settle })
	ctx, q := t.Context(), site.SiteDB.Queries()
	old := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -40)
	at := old.Add(5 * time.Hour).UnixMilli()
	_, err := q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "meal", Carbs: new(40.0), OccurredAt: at, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	_, err = q.SoftDeleteFoodIntake(ctx, sitedb.SoftDeleteFoodIntakeParams{ID: "meal"})
	require.NoError(t, err)
	// And a temp basal, which is a series and not an event, with no profile to make a basal rate of
	_, err = q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: "temp", DurationMinutes: 30, Multiplier: new(0.5), OccurredAt: at, CreatedByIdentityName: "loop"})
	require.NoError(t, err)

	for range 2 {
		_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
	}
	row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: old.Format(time.DateOnly)})
	require.NoError(t, err, "the day is packaged, though it has no event")
	_, day, err := retention.ReadEvents(row.Body)
	require.NoError(t, err)
	require.Empty(t, day.Events)
	food, err := q.AllFoodIntakes(ctx, sitedb.AllFoodIntakesParams{IncludeDeleted: true, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, food, "the deleted meal's row is deleted")
	basals, err := q.AllTempBasals(ctx, sitedb.AllTempBasalsParams{IncludeDeleted: true, Limit: 10})
	require.NoError(t, err)
	require.Empty(t, basals, "the temp basal's row is deleted")
}
