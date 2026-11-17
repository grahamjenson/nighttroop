package db_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's DB that cannot be opened at start, as when a migration of it keeps failing, is kept, so once it is fixed the
// next start opens it with its data
func TestSiteThatCannotOpenIsKept(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	controlDB, err := db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	path, backupPath := sitedb.GetPath(site.SiteDB.Subdomain, &cfg), sitedb.GetBackupPath(site.SiteDB.Subdomain, &cfg)
	require.NoError(t, site.SiteDB.Backup(t.Context()))
	require.NoError(t, controlDB.Close())

	setDirty := func(dirty bool) {
		raw, err := sql.Open("sqlite", path)
		require.NoError(t, err)
		defer raw.Close()
		_, err = raw.Exec("UPDATE schema_migrations SET dirty = ?", dirty)
		require.NoError(t, err)
	}

	// Dirty at a migration it already has, so forced back and migrated again, the migration fails, as one that keeps
	// failing does
	setDirty(true)
	_, err = db.NewControlDB(t.Context(), &cfg)
	require.ErrorContains(t, err, "sitedb migration 1 failed: SQL logic error: table glucose_readings already exists")
	require.True(t, utils.FileExists(path), "the site's DB is kept")
	require.True(t, utils.FileExists(backupPath), "and its backup")

	// Fixed, it opens again
	setDirty(false)
	controlDB, err = db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	reopened, _, err := controlDB.GetSite(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.NotNil(t, reopened)
}
