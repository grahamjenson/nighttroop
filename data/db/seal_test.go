package db_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's ended days are sealed: each day's CGM glucose becomes one row of cold_days, a value every 5 minutes. Its
// rows stay until the server deletes rows, and then only those older than its hot days go, and only those their cold
// day has. A day whose rows are gone still takes a late reading, and loses a deleted one.
func TestSealingASitesGlucose(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	t.Cleanup(func() { cfg.RetentionDeleteRows, cfg.RetentionHotDays = false, retention.DefaultHotDays })
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

	// sealed is a day's glucose by its 5 minute step, without the steps with none, or nil when the day is not sealed
	sealed := func(day time.Time) map[int]float64 {
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
	seal := func() {
		_, err := controlDB.TruncateSite(ctx, siteID)
		require.NoError(t, err)
	}

	// Sealing writes each ended day, in whole mg/dL, and deletes no row
	before := rows()
	seal()
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 48: 140, 60: 150}, sealed(old))
	require.Equal(t, map[int]float64{287: 180}, sealed(recent))
	require.Nil(t, sealed(today), "today has not ended")
	require.Equal(t, before, rows())

	// With rows deleted, the CGM readings older than the hot days go: not a hot day's, a meter's, one with a note, or
	// one a treatment refers to. The deleted reading, which was read as deleted, goes too.
	cfg.RetentionDeleteRows = true
	seal()
	require.ElementsMatch(t, []string{"noted", "treated", "meter", "recent", "today"}, rows())
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 48: 140, 60: 150}, sealed(old), "the day is as it was sealed")

	// A reading written late for that day, as a backfill does, is added to it, and the day keeps what it had. A
	// reading of it then deleted clears its step.
	reading("late", old.Add(7*time.Hour), 170, nil)
	_, err = q.SoftDeleteGlucoseReading(ctx, sitedb.SoftDeleteGlucoseReadingParams{ID: "noted"})
	require.NoError(t, err)
	seal()
	require.Equal(t, map[int]float64{0: 100, 12: 114, 24: 120, 60: 150, 84: 170}, sealed(old))
	require.NotContains(t, rows(), "late")

	// Fewer hot days delete the readings of 2 days ago too, and their day is as it was
	cfg.RetentionHotDays = 1
	seal()
	require.NotContains(t, rows(), "recent")
	require.Contains(t, rows(), "today")
	require.Equal(t, map[int]float64{287: 180}, sealed(recent))
}

// A run seals the days written since the run before, but all the days of a site never sealed
func TestSealingOnlyTheDaysWritten(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	write := func(id string, daysAgo int) {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: today.AddDate(0, 0, -daysAgo).UnixMilli(), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	isSealed := func(daysAgo int) bool {
		_, err := q.GetColdDay(ctx, sitedb.GetColdDayParams{Kind: "glucose", Date: today.AddDate(0, 0, -daysAgo).Format(time.DateOnly)})
		return err == nil
	}
	// A run that looks at nothing written before an hour from now still seals all of a site never sealed
	later := time.Now().Add(time.Hour).UnixMilli()
	write("a", 400)
	write("b", 3)
	require.NoError(t, controlDB.SealSite(ctx, site.SiteDB, later, time.Now()))
	require.True(t, isSealed(400) && isSealed(3))

	// Once sealed, a run seals only the days written since: not one written before, but it with a run that looks back
	write("c", 5)
	require.NoError(t, controlDB.SealSite(ctx, site.SiteDB, later, time.Now()))
	require.False(t, isSealed(5))
	require.NoError(t, controlDB.SealSite(ctx, site.SiteDB, time.Now().Add(-time.Hour).UnixMilli(), time.Now()))
	require.True(t, isSealed(5))
}
