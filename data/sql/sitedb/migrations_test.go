package sitedb

import (
	"context"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// A fresh site DB runs the index migration, so the All*/Search* queries that sort
// and filter on occurred_at, updated_at and the type columns walk an index instead
// of scanning and sorting the whole table. sqlc has no sqlite_master, so there is
// no prepared query to list the indexes; the applied migration version stands in,
// since golang-migrate fails NewDB when a CREATE INDEX in it fails.
func TestFreshSiteDBAppliesIndexMigration(t *testing.T) {
	db, err := NewDB(context.TODO(), "indexes", &config.ServerConfig{Mode: consts.Test})
	require.NoError(t, err)
	defer db.Close()

	driver, err := sqlite3.WithInstance(db.SQLiteDB, &sqlite3.Config{})
	require.NoError(t, err)
	source, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)
	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	require.NoError(t, err)

	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.False(t, dirty)
	require.EqualValues(t, 2, version)

	// The down migration removes everything the up migration added
	require.NoError(t, m.Steps(-1))
	version, _, err = m.Version()
	require.NoError(t, err)
	require.EqualValues(t, 1, version)
	require.NoError(t, m.Steps(1))
}
