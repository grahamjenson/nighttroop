package db_test

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's ended days are packaged: each day's CGM glucose becomes one row of cold_days, a value every 5 minutes. Its
// rows stay until the server deletes rows, and then only those older than its hot days go, and only those their cold
// day has. A day whose rows are gone still takes a late reading, and loses a deleted one.
func TestPackagingASitesGlucose(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.RetentionDeleteRows, cfg.RetentionHotDays = true, retention.DefaultHotDays })
	// The server deletes rows unless it is told not to: this starts as one that does not
	cfg.RetentionDeleteRows = false
	ctx, q, siteID := t.Context(), site.SiteDB.Queries(), site.SiteDB.Subdomain
	today := time.Now().UTC().Truncate(24 * time.Hour)
	old, recent := today.AddDate(0, 0, -40), today.AddDate(0, 0, -2)
	reading := func(id string, at time.Time, mgdl float64, change func(*sitedb.CreateGlucoseReadingParams)) {
		params := sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: mgdl, ReadingType: "sgv", OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone"}
		if change != nil {
			change(&params)
		}
		_, err := q.CreateGlucoseReading(ctx, params)
		require.NoError(t, err)
	}
	// 40 days ago: a reading at 00:00, two in the 5 minutes from 01:00 (the newest is kept), one at 02:00 that a sync
	// also read (the phone's is kept), one at 03:00 that was deleted, one with a note, one a treatment refers to, and
	// a meter's, which is not CGM glucose
	reading("midnight", old, 100.4, nil)
	reading("first", old.Add(time.Hour), 110, nil)
	reading("newest", old.Add(time.Hour+4*time.Minute), 114, nil)
	reading("phone", old.Add(2*time.Hour), 120, nil)
	reading("sync", old.Add(2*time.Hour+time.Second), 121, func(p *sitedb.CreateGlucoseReadingParams) { p.CreatedByIdentityName = "dexcom_connector" })
	reading("deleted", old.Add(3*time.Hour), 130, nil)
	_, err := q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: "deleted"})
	require.NoError(t, err)
	reading("noted", old.Add(4*time.Hour), 140, func(p *sitedb.CreateGlucoseReadingParams) { p.Notes = new("felt low") })
	reading("treated", old.Add(5*time.Hour), 150, nil)
	_, err = q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "check", GroupType: "BG Check", GlucoseReadingID: new("treated"), OccurredAt: old.Add(5 * time.Hour).UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	reading("meter", old.Add(6*time.Hour), 160, func(p *sitedb.CreateGlucoseReadingParams) { p.ReadingType = "mbg" })
	// 2 days ago, the last 5 minutes of the day, and today, which has not ended
	reading("recent", recent.Add(24*time.Hour-time.Second), 180, nil)
	reading("today", today.Add(time.Minute), 190, nil)

	// packaged is a day's glucose by its 5 minute step, without the steps with none, or nil when the day is not packaged
	packaged := func(day time.Time) map[int]float64 {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "glucose", Date: day.Format(time.DateOnly)})
		if err != nil {
			return nil
		}
		require.Equal(t, []byte{1, 1}, row.Body[:2], "it says its version and that it is a series")
		values, err := retention.UnpackGrid(row.Body)
		require.NoError(t, err)
		require.Len(t, values, 288)
		steps := map[int]float64{}
		for step, value := range values {
			if !math.IsNaN(value) {
				steps[step] = value
			}
		}
		return steps
	}
	rows := func() []string {
		all, err := q.AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 100})
		require.NoError(t, err)
		ids := []string{}
		for _, row := range all {
			ids = append(ids, row.ID)
		}
		return ids
	}
	packageDays := func() {
		_, err := controlDB.TruncateSite(ctx, siteID)
		require.NoError(t, err)
	}

	// Packaging writes each ended day, in whole mg/dL, and deletes no row
	before := rows()
	packageDays()
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 48: 140, 60: 150}, packaged(old))
	require.Equal(t, map[int]float64{287: 180}, packaged(recent))
	require.Nil(t, packaged(today), "today has not ended")
	require.Equal(t, before, rows())

	// With rows deleted, the CGM readings older than the hot days go: not a hot day's, a meter's, one with a note, or
	// one a treatment refers to. The deleted reading, which was read as deleted, goes too.
	cfg.RetentionDeleteRows = true
	packageDays()
	require.ElementsMatch(t, []string{"noted", "treated", "meter", "recent", "today"}, rows())
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 48: 140, 60: 150}, packaged(old), "the day is as it was packaged")

	// A reading written late for that day, as a backfill does, is added to it, and the day keeps what it had. A
	// reading of it then deleted clears its step.
	reading("late", old.Add(7*time.Hour), 170, nil)
	_, err = q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: "noted"})
	require.NoError(t, err)
	packageDays()
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 60: 150, 84: 170}, packaged(old))
	require.NotContains(t, rows(), "late")

	// Fewer hot days delete the readings of 2 days ago too, and their day is as it was
	cfg.RetentionHotDays = 1
	packageDays()
	require.NotContains(t, rows(), "recent")
	require.Contains(t, rows(), "today")
	require.Equal(t, map[int]float64{287: 180}, packaged(recent))
}

// A run packages the days written since the run before, but all the days of a site never packaged
func TestPackagingOnlyTheDaysWritten(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	write := func(id string, daysAgo int) {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: today.AddDate(0, 0, -daysAgo).UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	isPackaged := func(daysAgo int) bool {
		_, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "glucose", Date: today.AddDate(0, 0, -daysAgo).Format(time.DateOnly)})
		return err == nil
	}
	// A run that looks at nothing written before an hour from now still packages all of a site never packaged
	later := time.Now().Add(time.Hour).UnixMilli()
	write("a", 400)
	write("b", 3)
	require.NoError(t, controlDB.PackageSite(ctx, site.SiteDB, later, time.Now()))
	require.True(t, isPackaged(400) && isPackaged(3))

	// Once packaged, a run packages only the days written since: not one written before, but it with a run that looks back
	write("c", 5)
	require.NoError(t, controlDB.PackageSite(ctx, site.SiteDB, later, time.Now()))
	require.False(t, isPackaged(5))
	require.NoError(t, controlDB.PackageSite(ctx, site.SiteDB, time.Now().Add(-time.Hour).UnixMilli(), time.Now()))
	require.True(t, isPackaged(5))
}

// The readings of packaged days are read back the newest first, across as many days as are asked for, within the times
// asked, and only those the caller keeps
func TestColdGlucoseReadings(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	first := today.AddDate(0, 0, -600)
	// 500 days with 3 readings each, at 00:00, 12:00 and 23:55, each day's 1 mg/dL more than the day before's
	for dayOf := range 500 {
		for i, at := range []time.Duration{0, 12 * time.Hour, 24*time.Hour - 5*time.Minute} {
			_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint(dayOf, "-", i), Mgdl: float64(100 + dayOf%200), ReadingType: "sgv",
				OccurredAt: first.AddDate(0, 0, dayOf).Add(at).UnixMilli(), CreatedByIdentityName: "phone"})
			require.NoError(t, err)
		}
	}
	require.NoError(t, controlDB.PackageSite(ctx, site.SiteDB, 0, time.Now()))
	every := func(int64, float64) bool { return true }
	read := func(from, to time.Time, limit int, keep func(int64, float64) bool) []*sitedb.GlucoseReading {
		readings, err := db.ColdGlucoseReadings(ctx, q, from.UnixMilli(), to.UnixMilli(), limit, false, keep)
		require.NoError(t, err)
		return readings
	}

	// All of them, however many days that is, the newest first, each at the start of its step
	all := read(first, today, 5000, every)
	require.Len(t, all, 1500)
	// The oldest first is the same readings the other way round, and a few of them are the oldest
	oldest, err := db.ColdGlucoseReadings(ctx, q, first.UnixMilli(), today.UnixMilli(), 5000, true, every)
	require.NoError(t, err)
	require.Len(t, oldest, 1500)
	require.Equal(t, all[len(all)-1].ID, oldest[0].ID)
	require.Equal(t, all[0].ID, oldest[len(oldest)-1].ID)
	few, err := db.ColdGlucoseReadings(ctx, q, first.Add(time.Minute).UnixMilli(), today.UnixMilli(), 7, true, every)
	require.NoError(t, err)
	require.Len(t, few, 7)
	require.Equal(t, oldest[1].ID, few[0].ID)
	require.Equal(t, oldest[7].ID, few[6].ID)
	require.Equal(t, first.AddDate(0, 0, 499).Add(24*time.Hour-5*time.Minute).UnixMilli(), all[0].OccurredAt)
	require.Equal(t, first.UnixMilli(), all[1499].OccurredAt)
	require.True(t, slices.IsSortedFunc(all, func(a, b *sitedb.GlucoseReading) int { return int(b.OccurredAt - a.OccurredAt) }))
	require.Equal(t, db.ColdReadingID(all[0].OccurredAt), all[0].ID)
	require.Equal(t, "sgv", all[0].ReadingType)

	// The newest few of many days read only a few days, and are the same readings
	require.Equal(t, all[:7], read(first, today, 7, every))
	// From a time, up to and not including another, and only those kept
	day := first.AddDate(0, 0, 10)
	require.Len(t, read(day, day.Add(12*time.Hour), 100, every), 1)
	require.Len(t, read(day, day.Add(12*time.Hour+time.Millisecond), 100, every), 2)
	require.Len(t, read(day.Add(time.Millisecond), day.AddDate(0, 0, 2), 100, every), 5)
	require.Len(t, read(first, today, 5000, func(_ int64, mgdl float64) bool { return mgdl == 150 }), 9)
	require.Empty(t, read(today, first, 100, every))
	require.Empty(t, read(first.AddDate(-1, 0, 0), first, 100, every))
}

// What a loop's phone reports is packaged as series: its IOB and COB every 5 minutes, and the pump's reservoir and the
// batteries every hour, each a step's newest report. Today is packaged too, and a day keeps what it has once its
// reports are thinned.
func TestPackagingDeviceReports(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	day := today.AddDate(0, 0, -3)
	report := func(id string, at time.Time, device string, change func(*sitedb.CreateDeviceReportParams)) {
		params := sitedb.CreateDeviceReportParams{ID: id, OccurredAt: at.UnixMilli(), CreatedByIdentityName: "phone", CreatedByDeviceName: &device}
		change(&params)
		_, err := q.CreateDeviceReport(ctx, params)
		require.NoError(t, err)
	}
	loop := func(iob, basal, cob, reservoir, pump, phone float64) func(*sitedb.CreateDeviceReportParams) {
		return func(p *sitedb.CreateDeviceReportParams) {
			p.LoopName, p.LoopIob, p.LoopIobBasal, p.LoopCob = new("loop"), &iob, &basal, &cob
			p.PumpReservoir, p.PumpBatteryPercent, p.DeviceBatteryPercent = &reservoir, &pump, &phone
		}
	}
	// 3 days ago the loop's phone reports every 5 minutes for 2 hours, its IOB going up by 0.011 U. A carer's phone
	// reports its own battery, which is not the loop's phone's.
	for i := range 24 {
		report(fmt.Sprint("loop", i), day.Add(time.Duration(i)*5*time.Minute+time.Second), "loop://iPhone",
			loop(1+float64(i)*0.011, 0.25, 20-float64(i), 150-float64(i), 80, 60-float64(i)))
	}
	report("carer", day.Add(30*time.Minute), "carer", func(p *sitedb.CreateDeviceReportParams) { p.DeviceBatteryPercent = new(5.0) })
	// The pump says it is suspended at 00:20 and 00:25, and not at 00:30; the loop's other reports do not say
	for i, state := range []string{"suspended", "suspended", "normal"} {
		report(fmt.Sprint("pump", i), day.Add(time.Duration(4+i)*5*time.Minute+2*time.Second), "loop://iPhone", func(p *sitedb.CreateDeviceReportParams) { p.PumpState = &state })
	}
	// Today, a report in the last minute
	now := time.Now().UTC()
	report("now", now.Add(-time.Minute), "loop://iPhone", loop(3.333, 0.5, 12.34, 99.96, 70, 40))

	series := func(kind string, of time.Time) map[int]float64 {
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: kind, Date: of.Format(time.DateOnly)})
		if err != nil {
			return nil
		}
		values, err := retention.UnpackGrid(row.Body)
		require.NoError(t, err)
		steps := map[int]float64{}
		for step, value := range values {
			if !math.IsNaN(value) {
				steps[step] = value
			}
		}
		return steps
	}
	_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)

	// IOB to 0.01 U and COB to 0.1 g, every 5 minutes
	iob := series("iob", day)
	require.Len(t, iob, 24)
	require.Equal(t, 1.0, iob[0])
	require.Equal(t, 1.25, iob[23])
	require.Equal(t, map[int]float64{0: 20, 23: -3}, map[int]float64{0: series("cob", day)[0], 23: series("cob", day)[23]})
	require.Len(t, series("iob_basal", day), 24)
	// The reservoir and the batteries each hour, the hour's newest: of the loop's phone, not the carer's
	require.Equal(t, map[int]float64{0: 139, 1: 127}, series("pump_reservoir", day))
	require.Equal(t, map[int]float64{0: 80, 1: 80}, series("pump_battery", day))
	require.Equal(t, map[int]float64{0: 49, 1: 37}, series("phone_battery", day))
	// Whether the pump was suspended, every 5 minutes, where it said
	require.Equal(t, map[int]float64{4: 1, 5: 1, 6: 0}, series("pump_suspended", day))
	// Today is packaged as far as it has got
	nowStep := int(now.Add(-time.Minute).Sub(now.Add(-time.Minute).Truncate(24*time.Hour)) / (5 * time.Minute))
	require.Equal(t, map[int]float64{nowStep: 3.33}, series("iob", now.Add(-time.Minute)))
	require.Equal(t, 100.0, series("pump_reservoir", now.Add(-time.Minute))[now.Add(-time.Minute).Hour()])
	// A day with no device report has no device series, and its glucose is not theirs
	require.Nil(t, series("iob", day.AddDate(0, 0, -1)))
	require.Nil(t, series("glucose", day))

	// Truncating thinned the old day's reports to fewer, after it was packaged: the day keeps every value, and packaging
	// it again, when its rows are written, keeps them too
	rows, err := q.DeviceReportNumbersOfDay(ctx, sitedb.DeviceReportNumbersOfDayParams{From: day.UnixMilli(), To: day.Add(24 * time.Hour).UnixMilli()})
	require.NoError(t, err)
	require.Less(t, len(rows), 25)
	report("late", day.Add(3*time.Hour), "loop://iPhone", loop(2, 0.25, 5, 100, 75, 30))
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	iob = series("iob", day)
	require.Len(t, iob, 25)
	require.Equal(t, 1.25, iob[23])
	require.Equal(t, 2.0, iob[36])
	require.Equal(t, map[int]float64{0: 139, 1: 127, 3: 100}, series("pump_reservoir", day))
}

// A day packaged again unchanged, with none of its rows written since, keeps when it was packaged, so packaging a site
// again and again, as Truncate does, does not restart the wait before its rows are deleted. A row written since, though
// the day is the same, as a loop's dose that is not an event, a changed day and a day marked stale each say now.
func TestPackagingAnUnchangedDayKeepsWhenItWasPackaged(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	// A server that deletes rows would never package the old day again once it settled, nor one marked stale
	cfg.RetentionDeleteRows = false
	t.Cleanup(func() { cfg.RetentionDeleteRows = true })
	ctx, q := t.Context(), site.SiteDB.Queries()
	old := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -40)
	date := old.Format(time.DateOnly)
	dose := func(id string, automatic bool) {
		_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 4, Automatic: &automatic, OccurredAt: old.Add(5 * time.Hour).UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	packagedAt := func() int64 {
		_, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
		require.NoError(t, err)
		row, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: retention.EventsKind, Date: date})
		require.NoError(t, err)
		return row.UpdatedAt
	}
	// later is a time packaged after the last, once a millisecond has passed, so a write would show
	later := func(last int64) int64 {
		require.Eventually(t, func() bool { return time.Now().UnixMilli() > last }, time.Second, time.Millisecond)
		return packagedAt()
	}

	dose("bolus", false)
	first := packagedAt()
	require.Equal(t, first, later(first), "packaged again unchanged")

	dose("smb", true)
	written := later(first)
	require.Greater(t, written, first, "a row written since, though the day's events are the same")
	require.Equal(t, written, later(written))

	dose("correction", false)
	changed := later(written)
	require.Greater(t, changed, written, "a changed day")

	require.NoError(t, q.MarkColdDaysStale(ctx, sitedb.MarkColdDaysStaleParams{Date: date}))
	require.Greater(t, later(changed), changed, "a day marked to be packaged again")
}
