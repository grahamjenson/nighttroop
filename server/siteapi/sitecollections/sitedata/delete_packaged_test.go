package sitedata

import (
	"database/sql"
	"fmt"
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

	// A server that deletes rows deletes the old day's activity, target and note at once, as they are only events, which
	// the day packaged again keeps. The rest waits until the day has settled, a day here, as their series are built
	// from rows: the treatments, a loop's dose, the meal, the temp basal and the switches.
	cfg.RetentionDeleteRows = true
	packageDays()
	atOnce := map[string]int{}
	for table, n := range all {
		atOnce[table] = n
	}
	atOnce["activities"], atOnce["temp_targets"], atOnce["notification_events"] = 0, 0, 0
	require.Equal(t, atOnce, rows())
	packageDays()
	require.Equal(t, atOnce, rows())
	require.Subset(t, eventIDs(events(old)), []any{"target", "run", "note"}, "packaged again, the day keeps their events")

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
	require.ElementsMatch(t, []any{"over", "running", "meal", "bolus", "target", "run", "note"}, eventIDs(day))
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

// eventIDs are the IDs of a day's events
func eventIDs(day retention.EventsDay) []any {
	ids := []any{}
	for _, event := range day.Events {
		ids = append(ids, event["id"])
	}
	return ids
}

// A day packaged again keeps the events of its rows deleted at once (activities, temp targets, notes and treatments of
// only them, or of none, as a site change), merged by ID with those built from its rows that are left. One deleted by
// a person before is not brought back, and a temp target still in force at the next day's start stays in force there.
func TestEventsDeletedAtOnceStayInTheirDay(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.RetentionDeleteRows = true })
	cfg.RetentionDeleteRows = false
	ctx, q := t.Context(), site.SiteDB.Queries()
	must := func(_ any, err error) { require.NoError(t, err) }
	old := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -40)
	next, at, by := old.AddDate(0, 0, 1), old.Add(5*time.Hour).UnixMilli(), "phone"
	must(q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: "walk", ActivityType: "walk", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: "memo", Title: "Memo", NotificationType: "note", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: "mistake", Title: "Mistake", NotificationType: "note", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateTempTarget(ctx, sitedb.CreateTempTargetParams{ID: "overnight", LowTargetMgdl: 140, HighTargetMgdl: 160, DurationMinutes: 10 * 60, OccurredAt: old.Add(20 * time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "site", GroupType: "Site Change", OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "bolus", Dose: 4, OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "bolus-group", GroupType: "Correction Bolus", InsulinDoseID: new("bolus"), OccurredAt: at, CreatedByIdentityName: by}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "breakfast", Carbs: new(30.0), OccurredAt: next.Add(time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	truncate := func() {
		_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
	}
	day := func(of time.Time) retention.EventsDay {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: of.Format(time.DateOnly)})
		require.NoError(t, err)
		_, events, err := retention.ReadEvents(row.Body)
		require.NoError(t, err)
		return events
	}

	// Packaged with every row, then a person deletes a note, as can happen before the server deletes rows
	truncate()
	require.Contains(t, eventIDs(day(old)), "mistake")
	_, err := q.SoftDeleteNotificationEvent(ctx, sitedb.SoftDeleteNotificationEventParams{ID: "mistake", UpdatedByIdentityName: &by})
	require.NoError(t, err)

	// The server deletes rows: the events-only ones at once, while the bolus and its treatment wait for the day to settle
	cfg.RetentionDeleteRows = true
	truncate()
	_, err = q.GetActivity(ctx, sitedb.GetActivityParams{ID: "walk", IncludeDeleted: true})
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: "site", IncludeDeleted: true})
	require.ErrorIs(t, err, sql.ErrNoRows, "a treatment of no event")
	_, err = q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: "bolus-group"})
	require.NoError(t, err, "a treatment of a dose waits")
	require.ElementsMatch(t, []any{"walk", "memo", "overnight", "site", "bolus"}, eventIDs(day(old)), "the deleted note is not brought back")

	// A late row for each day packages them again: the old day keeps the events whose rows are gone, and the next day
	// keeps the temp target in force at its start, though its row is gone
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "late snack", Carbs: new(10.0), OccurredAt: old.Add(22 * time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	must(q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "late lunch", Carbs: new(50.0), OccurredAt: next.Add(12 * time.Hour).UnixMilli(), CreatedByIdentityName: by}))
	truncate()
	require.ElementsMatch(t, []any{"walk", "memo", "overnight", "site", "bolus", "late snack"}, eventIDs(day(old)))
	inForce := []any{}
	for _, event := range day(next).InForce {
		inForce = append(inForce, event["id"])
	}
	require.Contains(t, inForce, "overnight")
	require.ElementsMatch(t, []any{"breakfast", "late lunch"}, eventIDs(day(next)))
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

// A day with more events than a cold day holds is packaged with the most important of them: what a person ate and
// gave is kept before a device's notifications, as many of which are kept as fit. The day says how many it dropped,
// and all its rows are deleted.
func TestADayWithTooManyEventsKeepsTheMostImportant(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	settle := db.PackagedSettle
	db.PackagedSettle = 0
	t.Cleanup(func() { db.PackagedSettle = settle })
	ctx, q := t.Context(), site.SiteDB.Queries()
	old := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -40)
	// 600 notifications through the day, each with words that do not compress to nothing, and a meal and a bolus
	// late in it, after most of them
	for i := range 600 {
		_, err := q.CreateNotificationEvent(ctx, sitedb.CreateNotificationEventParams{ID: fmt.Sprint("note", i), Title: fmt.Sprintf("Alert %x", i*7919),
			Content: new(fmt.Sprintf("%x %x %x", i*104729, i*1299709, i*15485863)), NotificationType: "announcement",
			OccurredAt: old.Add(time.Duration(i) * time.Minute).UnixMilli(), CreatedByIdentityName: "loop"})
		require.NoError(t, err)
	}
	late := old.Add(20 * time.Hour).UnixMilli()
	_, err := q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: "meal", Carbs: new(40.0), OccurredAt: late, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	_, err = q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "bolus", Dose: 4, OccurredAt: late + 1000, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	for range 2 {
		_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
	}
	row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "events", Date: old.Format(time.DateOnly)})
	require.NoError(t, err)
	require.LessOrEqual(t, len(row.Body), 16384)
	_, day, err := retention.ReadEvents(row.Body)
	require.NoError(t, err)
	// The meal and the bolus are kept, in the day's order, after the notifications that fit, the oldest of them
	require.Greater(t, day.Dropped, 0)
	require.Equal(t, 602, len(day.Events)+day.Dropped)
	require.Greater(t, len(day.Events), 100, "as many notifications as fit are kept")
	require.Equal(t, []any{"meal", "bolus"}, []any{day.Events[len(day.Events)-2]["id"], day.Events[len(day.Events)-1]["id"]})
	require.Equal(t, "note0", day.Events[0]["id"])
	// Its rows are deleted, those that did not fit too
	notes, err := q.AllNotificationEvents(ctx, sitedb.AllNotificationEventsParams{Limit: 1000})
	require.NoError(t, err)
	require.Empty(t, notes)
	food, err := q.AllFoodIntakes(ctx, sitedb.AllFoodIntakesParams{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, food)
}
