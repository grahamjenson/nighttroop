package db_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// addReadings adds count glucose readings to the site, every step milliseconds back from newest, with IDs from 0 newest
func addReadings(t testing.TB, siteDB *db.SiteDB, count int64, newest int64, step int64) {
	ctx := t.Context()
	tx, err := siteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	insert := siteDB.Queries.WithTx(tx)
	for i := range count {
		_, err := insert.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint(i), Mgdl: 100, ReadingType: "sgv", OccurredAt: newest - i*step, CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
}

// TestSearchDateFilters checks each date filter, and each pair of them, finds the readings at the times it should
func TestSearchDateFilters(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	// Readings 9 to 0 at times 1 to 10 seconds, so reading 6 is at 4 seconds
	addReadings(t, site.SiteDB, 10, 10_000, 1_000)
	at := func(seconds int64) *int64 { v := seconds * 1_000; return &v }

	tests := []struct {
		name   string
		params sitedb.SearchGlucoseReadingsParams
		want   []string
	}{
		{"no filter", sitedb.SearchGlucoseReadingsParams{}, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{"from", sitedb.SearchGlucoseReadingsParams{DateGTE: at(4)}, []string{"0", "1", "2", "3", "4", "5", "6"}},
		{"after", sitedb.SearchGlucoseReadingsParams{DateGT: at(4)}, []string{"0", "1", "2", "3", "4", "5"}},
		{"up to", sitedb.SearchGlucoseReadingsParams{DateLTE: at(4)}, []string{"6", "7", "8", "9"}},
		{"before", sitedb.SearchGlucoseReadingsParams{DateLT: at(4)}, []string{"7", "8", "9"}},
		{"at", sitedb.SearchGlucoseReadingsParams{DateEQ: at(4)}, []string{"6"}},
		{"at no reading", sitedb.SearchGlucoseReadingsParams{DateEQ: new(int64(4_500))}, nil},
		{"from and before", sitedb.SearchGlucoseReadingsParams{DateGTE: at(3), DateLT: at(6)}, []string{"5", "6", "7"}},
		{"after and up to", sitedb.SearchGlucoseReadingsParams{DateGT: at(3), DateLTE: at(6)}, []string{"4", "5", "6"}},
		{"the later of from and after", sitedb.SearchGlucoseReadingsParams{DateGTE: at(5), DateGT: at(7)}, []string{"0", "1", "2"}},
		{"the earlier of up to and before", sitedb.SearchGlucoseReadingsParams{DateLTE: at(5), DateLT: at(3)}, []string{"8", "9"}},
		{"at and after it", sitedb.SearchGlucoseReadingsParams{DateEQ: at(4), DateGT: at(4)}, nil},
		{"at and from it", sitedb.SearchGlucoseReadingsParams{DateEQ: at(4), DateGTE: at(4)}, []string{"6"}},
		{"from after up to", sitedb.SearchGlucoseReadingsParams{DateGTE: at(6), DateLTE: at(3)}, nil},
		{"after a negative time", sitedb.SearchGlucoseReadingsParams{DateGT: at(-5)}, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{"a page", sitedb.SearchGlucoseReadingsParams{DateGTE: at(2), Skip: 2, Limit: 3}, []string{"2", "3", "4"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.params.Limit == 0 {
				tt.params.Limit = 100
			}
			rows, err := site.SiteDB.Queries.SearchGlucoseReadings(t.Context(), tt.params)
			require.NoError(t, err)
			var ids []string
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			require.Equal(t, tt.want, ids)
		})
	}
}

// BenchmarkSearchDateRange reads a site's glucose readings in a time range, among a year of readings every 5 minutes.
// Run with go test -run XXX -bench SearchDateRange ./data/db
func BenchmarkSearchDateRange(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	now := time.Now().UnixMilli()
	const fiveMinutes, day = int64(5 * 60 * 1000), int64(24 * 60 * 60 * 1000)
	addReadings(b, site.SiteDB, 365*288, now, fiveMinutes)

	tests := []struct {
		name   string
		params sitedb.SearchGlucoseReadingsParams
		want   int
	}{
		{"since the last poll", sitedb.SearchGlucoseReadingsParams{DateGT: new(now - fiveMinutes - 1), Limit: 1000}, 2},
		{"the last day", sitedb.SearchGlucoseReadingsParams{DateGTE: new(now - day + 1), DateLTE: &now, Limit: 10000}, 288},
		{"a day 90 days ago", sitedb.SearchGlucoseReadingsParams{DateGTE: new(now - 91*day + 1), DateLT: new(now - 90*day + 1), Limit: 10000}, 288},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				rows, err := site.SiteDB.Queries.SearchGlucoseReadings(b.Context(), tt.params)
				require.NoError(b, err)
				require.Len(b, rows, tt.want)
			}
		})
	}
}
