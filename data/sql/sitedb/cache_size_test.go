package sitedb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// A site DB's connection has a 64 KiB page cache, as thousands of sites each keep connections, while the server's own
// DBs keep SQLite's 2 MB
func TestSiteDBCacheSize(t *testing.T) {
	serverConfig := &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()}
	site, err := sitedb.NewDB(t.Context(), "abcdefgh", serverConfig)
	require.NoError(t, err)
	t.Cleanup(func() { site.Close() })
	var cacheSize int64
	require.NoError(t, site.SQLiteDB.QueryRowContext(t.Context(), "PRAGMA cache_size").Scan(&cacheSize))
	require.Equal(t, int64(-64), cacheSize)

	control, err := controldb.NewDB(t.Context(), serverConfig)
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	require.NoError(t, control.SQLiteDB.QueryRowContext(t.Context(), "PRAGMA cache_size").Scan(&cacheSize))
	require.Equal(t, int64(-2000), cacheSize)
}
