package sitedb

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
