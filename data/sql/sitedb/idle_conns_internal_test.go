package sitedb

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// readAtOnce reads n times at once, each in a transaction held open until all have read, so each needs its own
// connection
func readAtOnce(t *testing.T, db *sql.DB, n int) {
	txs := []*sql.Tx{}
	for range n {
		tx, err := db.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true})
		require.NoError(t, err)
		_, err = New(tx).SearchGlucoseReadings(t.Context(), SearchGlucoseReadingsParams{Limit: 1})
		require.NoError(t, err)
		txs = append(txs, tx)
	}
	for _, tx := range txs {
		require.NoError(t, tx.Rollback())
	}
}

// A busy site's DB keeps up to maxIdleConns idle connections, so its requests at once do not each open a new one, and
// closes them once unused, as each holds its own parsed schema, most of an open site's memory
func TestIdleConnectionsClose(t *testing.T) {
	defaultIdleTime := connMaxIdleTime
	connMaxIdleTime = 500 * time.Millisecond
	t.Cleanup(func() { connMaxIdleTime = defaultIdleTime })

	db, err := OpenSqliteDB(filepath.Join(t.TempDir(), "site.sqlite"), 0)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, MigrateUp(db))

	// More reads at once than are kept need a connection each, and maxIdleConns of them are kept after
	require.Equal(t, 4, maxIdleConns)
	readAtOnce(t, db, maxIdleConns+2)
	require.Equal(t, maxIdleConns, db.Stats().Idle)

	// Then they close, once unused
	require.Eventually(t, func() bool { return db.Stats().OpenConnections == 0 }, 3*time.Second, 10*time.Millisecond)
}

// A site's DB is opened at start with no connection left open, as every site's is opened then, but still keeps up to
// maxIdleConns once used, while one in memory keeps its one connection, which holds its data
func TestNewDBLeavesNoConnectionOpen(t *testing.T) {
	file, err := NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { file.Close() })
	require.Equal(t, 0, file.SQLiteDB.Stats().OpenConnections)
	_, err = file.Queries().AllActivities(t.Context(), AllActivitiesParams{Limit: 1})
	require.NoError(t, err)
	readAtOnce(t, file.SQLiteDB, maxIdleConns+2)
	require.Equal(t, maxIdleConns, file.SQLiteDB.Stats().Idle)

	memory, err := NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: consts.Test, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { memory.Close() })
	require.Equal(t, 1, memory.SQLiteDB.Stats().OpenConnections)
	_, err = memory.Queries().AllActivities(t.Context(), AllActivitiesParams{Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 1, memory.SQLiteDB.Stats().OpenConnections)
}
