package db_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkReadsByKey reads what a site's prepared queries read most often by a key: a treatment by its ID, as API v3
// history and each treatment uploaded do, and a device's newest status, as the alarm checker does every minute. Each run
// binds a new key, which must not make SQLite prepare the query again.
// Run with go test -run XXX -bench ReadsByKey ./data/db
func BenchmarkReadsByKey(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	ctx := b.Context()
	const fiveMinutes = int64(5 * 60 * 1000)
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(b, err)
	insert := site.SiteDB.QueriesWithTx(tx)
	device := "openaps://phone"
	// A day of a looper's boluses, each a treatment with its insulin dose, and its statuses
	for i := range int64(288) {
		dose, err := insert.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: fmt.Sprint("d", i), Dose: 0.5,
			OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
		_, err = insert.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: fmt.Sprint("g", i), GroupType: "Correction Bolus",
			InsulinDoseID: &dose.ID, OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
		_, err = insert.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprint("s", i), CreatedByDeviceName: &device,
			OccurredAt: i * fiveMinutes, CreatedByIdentityName: "test"})
		require.NoError(b, err)
	}
	require.NoError(b, tx.Commit())
	q := site.SiteDB.Queries()

	b.Run("treatment by ID", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			row, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: fmt.Sprint("g", i%288)})
			require.NoError(b, err)
			require.NotNil(b, row.InsulinDosesVw.ID)
		}
	})
	b.Run("device's newest status", func(b *testing.B) {
		devices := []*string{&device, new("xDrip-DexcomG6")}
		i := 0
		for b.Loop() {
			i++
			_, err := q.DeviceReportsForDevice(ctx, sitedb.DeviceReportsForDeviceParams{Device: devices[i%2], Limit: 1})
			require.NoError(b, err)
		}
	})
}
