package db_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's DB that cannot be opened at start, as when a migration of it keeps failing, makes the site unavailable while
// the server starts without it. It is kept, also when its backup does not open either, so once it is fixed the next
// start opens it with its data.
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

	// At the first migration, whose CREATEs fail when run again, while the later ones run again on it
	setDirty := func(dirty bool) {
		raw, err := sql.Open("sqlite", path)
		require.NoError(t, err)
		defer raw.Close()
		_, err = raw.Exec("UPDATE schema_migrations SET version = 1, dirty = ?", dirty)
		require.NoError(t, err)
	}

	// Dirty at a migration it already has, so forced back and migrated again, the migration fails, as one that keeps
	// failing does
	setDirty(true)
	controlDB, err = db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err, "the start goes on without the site")
	require.Equal(t, map[string]string{site.SiteDB.Subdomain: db.UnavailableCannotOpen}, controlDB.UnavailableSites())
	_, _, err = controlDB.GetSite(t.Context(), site.SiteDB.Subdomain)
	require.ErrorIs(t, err, db.ErrSiteUnavailable)

	// Its backup, a copy of the same DB, does not open either, so the DB that was there is put back
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	cfg.Status.S3Working = true
	controlDB.RestoreUnavailableSites(t.Context(), func(_ context.Context, _ string, outpath string) (int64, bool, error) {
		require.NoError(t, os.MkdirAll(outpath, 0o755))
		return int64(len(before)), true, os.WriteFile(filepath.Join(outpath, filepath.Base(path)), before, 0o600)
	})
	require.Equal(t, map[string]string{site.SiteDB.Subdomain: db.UnavailableCannotOpen}, controlDB.UnavailableSites())
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after, "the site's DB is kept")
	require.True(t, utils.FileExists(backupPath), "and its backup")
	require.NoError(t, controlDB.Close())

	// Fixed, it opens again
	setDirty(false)
	controlDB, err = db.NewControlDB(t.Context(), &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	reopened, _, err := controlDB.GetSite(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.NotNil(t, reopened)
}
