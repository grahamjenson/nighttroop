package db_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkEventsToGroup finds the events to pick for a group, as the site's create and edit forms offer them, among
// a year of readings every five minutes and a year of temp basals, each in its own group, as a Loop user has.
// Run with go test -run XXX -bench EventsToGroup ./data/db
func BenchmarkEventsToGroup(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	ctx := b.Context()
	const hour, fiveMinutes = int64(60 * 60 * 1000), int64(5 * 60 * 1000)
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(b, err)
	insert := site.SiteDB.QueriesWithTx(tx)
	for i := range int64(100_000) {
		_, err := insert.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint("r", i), Mgdl: 100,
			ReadingType: "sgv", OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
		_, err = insert.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: fmt.Sprint("t", i), DurationMinutes: 30,
			Rate: 1, OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
		_, err = insert.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: fmt.Sprint("g", i), GroupType: "Temp Basal",
			TempBasalID: new(fmt.Sprint("t", i)), OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
	}
	require.NoError(b, tx.Commit())
	q := site.SiteDB.Queries()

	// The two hours around a time half way through the year, for a new group and for editing one in them
	at := 50_000 * fiveMinutes
	for name, groupID := range map[string]string{"new group": "", "editing a group": "g50000"} {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				rows, err := q.EventsToGroup(b.Context(), sitedb.EventsToGroupParams{GroupID: groupID, From: at - hour, To: at + hour})
				require.NoError(b, err)
				// Every reading in the window, and only the edited group's own temp basal
				want := 25
				if groupID != "" {
					want++
				}
				require.Len(b, rows, want)
			}
		})
	}
}
