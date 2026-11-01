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

// A site's DB keeps at most one idle connection and closes it once unused, as each holds its own parsed schema, most of
// an open site's memory, and a new one takes about a millisecond
func TestIdleConnectionsClose(t *testing.T) {
	defaultIdleTime := connMaxIdleTime
	connMaxIdleTime = 100 * time.Millisecond
	t.Cleanup(func() { connMaxIdleTime = defaultIdleTime })

	db, err := OpenSqliteDB(filepath.Join(t.TempDir(), "site.sqlite"), 0)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, MigrateUp(db))

	// Two reads at once need two connections, but only one is kept after
	tx, err := db.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	_, err = New(tx).SearchGlucoseReadings(t.Context(), SearchGlucoseReadingsParams{Limit: 1})
	require.NoError(t, err)
	_, err = New(db).SearchGlucoseReadings(t.Context(), SearchGlucoseReadingsParams{Limit: 1})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.Equal(t, 1, db.Stats().Idle)

	// Then that one closes, once unused
	require.Eventually(t, func() bool { return db.Stats().OpenConnections == 0 }, 2*time.Second, 10*time.Millisecond)
}

// A site's DB is opened at start with no connection left open, as every site's is opened then, while one in memory
// keeps its connection, which holds its data
func TestNewDBLeavesNoConnectionOpen(t *testing.T) {
	file, err := NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { file.Close() })
	require.Equal(t, 0, file.SQLiteDB.Stats().OpenConnections)
	_, err = file.Queries().AllActivities(t.Context(), AllActivitiesParams{Limit: 1})
	require.NoError(t, err)

	memory, err := NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: consts.Test, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { memory.Close() })
	require.Equal(t, 1, memory.SQLiteDB.Stats().OpenConnections)
}
