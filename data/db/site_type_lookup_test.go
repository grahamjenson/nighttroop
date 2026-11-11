package db_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkTypeLookups finds a type's newest treatments among a year of them: a Nightscout properties request looks up
// the newest of five change types, some of which a site may never have, and a delete by type takes the newest 1000 of a
// type.
// Run with go test -run XXX -bench TypeLookups ./data/db
func BenchmarkTypeLookups(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	ctx := b.Context()
	const hour, fiveMinutes = int64(60 * 60 * 1000), int64(5 * 60 * 1000)
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(b, err)
	insert := site.SiteDB.QueriesWithTx(tx)
	// Mostly calibrations, with a site change every few days, as an xDrip+ and pump user has
	changeTypes := []string{"Calibration", "Calibration", "Calibration", "Calibration", "Site Change", "Sensor Change"}
	for i := range int64(2000) {
		_, err := insert.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: fmt.Sprint("c", i),
			GroupType: changeTypes[i%int64(len(changeTypes))], OccurredAt: i * 4 * hour, CreatedByIdentityName: "test"})
		require.NoError(b, err)
	}
	// A temp basal every five minutes, with a bolus among every ten treatments
	groupTypes := []string{"Temp Basal", "Temp Basal", "Temp Basal", "Temp Basal", "Temp Basal", "Temp Basal", "Temp Basal", "Temp Basal", "Correction Bolus", "Meal Bolus"}
	for i := range int64(100_000) {
		_, err := insert.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: fmt.Sprint("g", i),
			GroupType: groupTypes[i%int64(len(groupTypes))], OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
	}
	require.NoError(b, tx.Commit())
	q := site.SiteDB.Queries()

	for _, groupType := range []string{"Site Change", "Pump Battery Change"} {
		b.Run("newest "+groupType, func(b *testing.B) {
			for b.Loop() {
				groups, err := q.GetEventGroupOfTypeWithoutJoins(b.Context(), sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: groupType, Limit: 1})
				require.NoError(b, err)
				if groupType == "Site Change" {
					require.Len(b, groups, 1)
				}
			}
		})
	}
	b.Run("newest 1000 temp basals", func(b *testing.B) {
		for b.Loop() {
			groups, err := q.GetEventGroupOfTypeWithoutJoins(b.Context(), sitedb.GetEventGroupOfTypeWithoutJoinsParams{GroupType: "Temp Basal", Limit: 1000})
			require.NoError(b, err)
			require.Len(b, groups, 1000)
			require.Equal(b, "g99997", groups[0].ID)
		}
	})
}
