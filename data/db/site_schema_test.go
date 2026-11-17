package db_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/migrateup"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site DB at the latest migration whose schema is not the one the migrations make on a new DB, as one made before a
// migration was edited, makes the site unavailable as its schema differs. So does a backup like it, and a backup with
// the migrations' schema restores it.
func TestSiteWhoseSchemaDiffersIsUnavailable(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir, cfg.Status.S3Working = consts.Development, t.TempDir(), true
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	site, err := controlDB.CreateSite(ctx)
	require.NoError(t, err)
	siteID, path := site.SiteDB.Subdomain, sitedb.GetPath(site.SiteDB.Subdomain, &cfg)
	require.NoError(t, controlDB.Close())
	good, err := os.ReadFile(path)
	require.NoError(t, err)

	// The site's DB made again by its first migration as it was before an edit, with a table the migrations do not make
	edited, err := fs.ReadFile(sitedb.Migrations, "migrations/000001_init_schema.up.sql")
	require.NoError(t, err)
	before := fstest.MapFS{
		"migrations/000001_init_schema.up.sql":   {Data: append(edited, "\nCREATE TABLE dropped (id INTEGER PRIMARY KEY);\n"...)},
		"migrations/000001_init_schema.down.sql": {Data: []byte("")},
	}
	require.NoError(t, utils.DeleteFiles([]string{path, path + "-wal", path + "-shm"}))
	raw, err := sitedb.OpenSqliteDB(path, 0)
	require.NoError(t, err)
	require.NoError(t, migrateup.Up("sitedb", raw, before))
	require.NoError(t, raw.Close())
	bad, err := os.ReadFile(path)
	require.NoError(t, err)

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err, "the start goes on without the site")
	t.Cleanup(func() { controlDB.Close() })
	require.Equal(t, map[string]string{siteID: db.UnavailableSchemaDiffers}, controlDB.UnavailableSites())
	_, _, err = controlDB.GetSite(ctx, siteID)
	require.ErrorIs(t, err, db.ErrSiteUnavailable)

	restore := func(backup []byte) {
		controlDB.RestoreUnavailableSites(ctx, func(_ context.Context, _ string, outpath string) (int64, bool, error) {
			require.NoError(t, os.MkdirAll(outpath, 0o755))
			return int64(len(backup)), true, os.WriteFile(filepath.Join(outpath, filepath.Base(path)), backup, 0o600)
		})
	}
	// A backup with the same schema stays unavailable, and its event says why
	restore(bad)
	require.Equal(t, map[string]string{siteID: db.UnavailableSchemaDiffers}, controlDB.UnavailableSites())
	found, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_db_missing"), Limit: 10})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.JSONEq(t, `{"restored":false,"reason":"schema_differs"}`, *found[0].Attributes)

	// One with the migrations' schema restores it
	restore(good)
	require.Empty(t, controlDB.UnavailableSites())
	_, _, err = controlDB.GetSite(ctx, siteID)
	require.NoError(t, err)
}
