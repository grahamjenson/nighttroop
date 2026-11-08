package db_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's DB is measured by table, each table's indexes counted with it, largest first, and a site whose DB is not
// open here has no tables to measure
func TestSiteTableSizes(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	tables, err := controlDB.SiteTableSizes(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.NotEmpty(t, tables)
	names := map[string]bool{}
	for i, table := range tables {
		names[table.Name] = true
		require.Positive(t, table.Bytes)
		require.False(t, strings.Contains(table.Name, "autoindex") || strings.HasSuffix(table.Name, "_idx"), "an index is counted with its table: %s", table.Name)
		if i > 0 {
			require.LessOrEqual(t, table.Bytes, tables[i-1].Bytes)
		}
	}
	require.True(t, names["glucose_readings"])

	tables, err = controlDB.SiteTableSizes(t.Context(), "no-such-site")
	require.NoError(t, err)
	require.Empty(t, tables)
}
