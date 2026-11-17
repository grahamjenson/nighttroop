package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/testutils"
)

// The server's own DBs are measured by table as a site's is, largest first, by the resource check, and kept in memory,
// so the usage page reads only that, not each whole DB again
func TestServerTableSizes(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	names := []string{}
	for _, serverDB := range controlDB.ServerDBs() {
		names = append(names, serverDB.Name)
		require.Empty(t, serverDB.Tables, "none before the first measure: %s", serverDB.Name)
	}
	require.Equal(t, []string{"control", "logs", "jobs", "live"}, names)

	records := func() map[string]map[string]int64 {
		require.NoError(t, controlDB.RecordServerTableSizes(t.Context()))
		records := map[string]map[string]int64{}
		for _, serverDB := range controlDB.ServerDBs() {
			records[serverDB.Name] = tableRecords(t, serverDB)
		}
		return records
	}
	before := records()
	require.Contains(t, before["control"], "sites")
	require.Contains(t, before["logs"], "events")
	require.Contains(t, before["jobs"], "jobs")
	require.Contains(t, before["live"], "resource_readings")
	_, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	require.Equal(t, before["control"]["sites"]+1, records()["control"]["sites"])
}

// tableRecords are a DB's tables' records by name, checking each has bytes, largest first
func tableRecords(t *testing.T, serverDB db.ServerDB) map[string]int64 {
	require.NotEmpty(t, serverDB.Tables, serverDB.Name)
	records := map[string]int64{}
	for i, table := range serverDB.Tables {
		records[table.Name] = table.Records
		require.Positive(t, table.Bytes)
		if i > 0 {
			require.LessOrEqual(t, table.Bytes, serverDB.Tables[i-1].Bytes)
		}
	}
	return records
}
