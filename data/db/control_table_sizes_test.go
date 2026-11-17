package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// The control DB is measured by table as a site's is, largest first, by the resource check, and kept in memory, so the
// usage page reads only that, not its whole DB again
func TestControlTableSizes(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	_, tables := controlDB.ControlDBSize()
	require.Empty(t, tables, "none before the first measure")

	records := func() map[string]int64 {
		require.NoError(t, controlDB.RecordControlTableSizes(t.Context()))
		_, tables := controlDB.ControlDBSize()
		require.NotEmpty(t, tables)
		records := map[string]int64{}
		for i, table := range tables {
			records[table.Name] = table.Records
			require.Positive(t, table.Bytes)
			if i > 0 {
				require.LessOrEqual(t, table.Bytes, tables[i-1].Bytes)
			}
		}
		return records
	}
	before := records()
	require.Contains(t, before, "sites")
	_, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	require.Equal(t, before["sites"]+1, records()["sites"])
}
