package controldb

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// Every request uses the control DB, so its file keeps an idle connection for each of many requests at once, rather
// than closing all but database/sql's default of two and opening them again for the next requests
func TestIdleConnectionsKept(t *testing.T) {
	db, err := NewDB(t.Context(), &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.False(t, db.InMemory)

	require.Equal(t, 32, maxIdleConns)
	txs := []*sql.Tx{}
	for range maxIdleConns + 2 {
		tx, err := db.SQLiteDB.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true})
		require.NoError(t, err)
		_, err = db.Queries.WithTx(tx).GetServerSettings(t.Context())
		require.NoError(t, err)
		txs = append(txs, tx)
	}
	for _, tx := range txs {
		require.NoError(t, tx.Rollback())
	}
	require.Equal(t, maxIdleConns, db.SQLiteDB.Stats().Idle)
}
