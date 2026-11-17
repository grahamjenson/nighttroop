package db_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkSearchTotals counts a site's glucose readings and a device's statuses, as each page of a site's search
// does, among a year of each every 5 minutes, the statuses with a loop's status a day old (its predictions cleared).
// Run with go test -run XXX -bench SearchTotals ./data/db
func BenchmarkSearchTotals(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	now := time.Now().UnixMilli()
	const count, fiveMinutes = 365 * 288, int64(5 * 60 * 1000)
	addReadings(b, site.SiteDB, count, now, fiveMinutes)

	ctx := b.Context()
	status := `{"iob":{"iob":1.5},"suggested":{"reason":"` + strings.Repeat("COB 0, Dev 5, BGI -1, ISF 50; ", 40) + `"}}`
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(b, err)
	for i := range int64(count) {
		_, err := site.SiteDB.QueriesWithTx(tx).CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: fmt.Sprint(i), LoopStatusJson: &status,
			LoopName: new("openaps"), OccurredAt: now - i*fiveMinutes, CreatedByDeviceName: new("phone"), CreatedByIdentityName: "test"})
		require.NoError(b, err)
	}
	require.NoError(b, tx.Commit())

	b.Run("glucose readings", func(b *testing.B) {
		for b.Loop() {
			totals, err := site.SiteDB.Queries().GlucoseReadingTotals(b.Context())
			require.NoError(b, err)
			require.Equal(b, int64(count), totals.Total)
		}
	})
	b.Run("a device's statuses", func(b *testing.B) {
		for b.Loop() {
			totals, err := site.SiteDB.Queries().DeviceReportTotals(b.Context(), sitedb.DeviceReportTotalsParams{Device: new("phone")})
			require.NoError(b, err)
			require.Equal(b, int64(count), totals.Total)
		}
	})
}
