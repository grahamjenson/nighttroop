package db_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's DB is measured by table, each table's indexes counted with it, largest first, with its records, which its
// indexes' entries are not counted in, and a site whose DB is not open here has no tables to measure
func TestSiteTableSizes(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	for i := range 3 {
		_, err := site.SiteDB.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint("reading", i), Mgdl: 120, ReadingType: "sgv", OccurredAt: int64(i + 1), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	tables, err := controlDB.SiteTableSizes(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.NotEmpty(t, tables)
	names, records := map[string]bool{}, map[string]int64{}
	for i, table := range tables {
		names[table.Name], records[table.Name] = true, table.Records
		require.Positive(t, table.Bytes)
		require.False(t, strings.Contains(table.Name, "autoindex") || strings.HasSuffix(table.Name, "_idx"), "an index is counted with its table: %s", table.Name)
		if i > 0 {
			require.LessOrEqual(t, table.Bytes, tables[i-1].Bytes)
		}
	}
	require.True(t, names["glucose_readings"])
	require.Equal(t, int64(3), records["glucose_readings"])
	require.Zero(t, records["device_reports"])

	tables, err = controlDB.SiteTableSizes(t.Context(), "no-such-site")
	require.NoError(t, err)
	require.Empty(t, tables)
}
