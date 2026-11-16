package sitedb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// addReadings adds readings from to to in a transaction, each with a note of noteLength
func addReadings(t *testing.T, db *DB, from, to, noteLength int) {
	tx, err := db.SQLiteDB.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	note := strings.Repeat("n", noteLength)
	for i := from; i < to; i++ {
		_, err = db.QueriesWithTx(tx).CreateGlucoseReading(t.Context(), CreateGlucoseReadingParams{ID: fmt.Sprint(i), Mgdl: 100,
			ReadingType: "sgv", OccurredAt: int64(i + 1), Notes: &note, CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
}

// A site's DB runs its queries, in a transaction or not, on the connection it has, each prepared there on its first
// run. Preparing a query first run in a transaction on another connection kept a second one open, with its own schema
// and copy of each statement, which doubled a busy site's memory.
func TestSiteQueriesUseOneConnection(t *testing.T) {
	for _, mode := range []consts.ServerMode{consts.Test, consts.Development} { // in memory, then in a file
		t.Run(string(mode), func(t *testing.T) {
			ctx := t.Context()
			db, err := NewDB(ctx, "abcdefgh", &config.ServerConfig{Mode: mode, DataDir: t.TempDir()})
			require.NoError(t, err)
			t.Cleanup(func() { db.Close() })

			for i := range 3 {
				addReadings(t, db, i, i+1, 0)
				readings, err := db.Queries().SearchGlucoseReadings(ctx, SearchGlucoseReadingsParams{Limit: 10})
				require.NoError(t, err)
				require.Len(t, readings, i+1)
			}
			require.Equal(t, 1, db.SQLiteDB.Stats().OpenConnections)

			// Many first runs at once each succeed
			var wg sync.WaitGroup
			for range 10 {
				wg.Go(func() {
					_, err := db.Queries().AllTempTargets(ctx, AllTempTargetsParams{Limit: 1})
					assert.NoError(t, err)
				})
			}
			wg.Wait()
		})
	}
}

// residentKB is the process's resident memory in KB, which counts SQLite's, or 0 when the OS cannot say
func residentKB() int {
	status, _ := os.ReadFile("/proc/self/status")
	for line := range strings.Lines(string(status)) {
		if fields := strings.Fields(line); len(fields) > 1 && fields[0] == "VmRSS:" {
			kb, _ := strconv.Atoi(fields[1])
			return kb
		}
	}
	return 0
}

// A connection finalizes its statements when it closes, as one closed with any left open is kept by SQLite, with its
// schema and statements, about 1MB, for the life of the server: so connections opened and closed as a site is busy, then
// idle for connMaxIdleTime, do not add up
func TestClosedConnectionsFreeTheirStatements(t *testing.T) {
	ctx := t.Context()
	db, err := NewDB(ctx, "abcdefgh", &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.SQLiteDB.SetMaxIdleConns(0) // each connection closes once used
	write := func(from, to int) {
		for i := from; i < to; i++ {
			addReadings(t, db, i, i+1, 0)
			_, err := db.Queries().SearchEventGroups(ctx, SearchEventGroupsParams{Limit: 1})
			require.NoError(t, err)
		}
	}
	write(0, 20)
	before := residentKB()
	write(20, 320)
	if before == 0 {
		t.Skip("the OS does not say the process's memory")
	}
	require.Less(t, residentKB()-before, 100*1024, "300 connections, each kept with its statements, would keep about 300MB")
}
