package main

import (
	"crypto/sha256"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// runCheck is what check writes, and whether it says the start is stopped
func runCheck(t *testing.T, cfg *config.ServerConfig) (string, bool) {
	t.Helper()
	var out strings.Builder
	stopped, err := check(t.Context(), cfg, &out)
	require.NoError(t, err)
	return out.String(), stopped
}

// newSite is a site in the control DB with a DB that opens, as a server makes it
func newSite(t *testing.T, cfg *config.ServerConfig, control *controldb.DB, id string) *sitedb.DB {
	t.Helper()
	_, err := control.Queries.CreateSite(t.Context(), controldb.CreateSiteParams{ID: id, BackupKeyHash: "hash"})
	require.NoError(t, err)
	db, err := sitedb.NewDB(t.Context(), id, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// dbFiles are the hashes of a DB's file and its write-ahead log
func dbFiles(t *testing.T, path string) [2][32]byte {
	t.Helper()
	db, err := os.ReadFile(path)
	require.NoError(t, err)
	wal, _ := os.ReadFile(path + "-wal")
	return [2][32]byte{sha256.Sum256(db), sha256.Sum256(wal)}
}

func TestCheck(t *testing.T) {
	ctx := t.Context()
	cfg := &config.ServerConfig{Mode: consts.Development, DataDir: t.TempDir()}
	out, stopped := runCheck(t, cfg)
	require.False(t, stopped)
	require.Contains(t, out, "There is no control DB")

	control, err := controldb.NewDB(ctx, cfg)
	require.NoError(t, err)
	defer control.Close()
	newSite(t, cfg, control, "abcdefgh").Close()
	out, stopped = runCheck(t, cfg)
	require.False(t, stopped, out)
	require.Contains(t, out, "Every one of the 1 live and disabled sites opens and migrates")

	// A live site whose DB is missing, with a backup in the data folder
	newSite(t, cfg, control, "bcdefghi").Close()
	require.NoError(t, os.Remove(sitedb.GetPath("bcdefghi", cfg)))
	backupPath := sitedb.GetBackupPath("bcdefghi", cfg)
	require.NoError(t, os.MkdirAll(strings.TrimSuffix(backupPath, "/bcdefghi.sqlite"), 0o700))
	require.NoError(t, os.WriteFile(backupPath, []byte("a backup"), 0o600))
	made := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(backupPath, made, made))
	// A disabled site whose DB is not a DB
	newSite(t, cfg, control, "cdefghij").Close()
	_, err = control.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: "cdefghij"})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(sitedb.GetPath("cdefghij", cfg), []byte("not a database"), 0o600))
	// A live site a failed migration left dirty at its first, which fails again, as its tables are there
	dirty := newSite(t, cfg, control, "defghijk")
	driver, err := sqlite.WithInstance(dirty.SQLiteDB, &sqlite.Config{})
	require.NoError(t, err)
	require.NoError(t, driver.SetVersion(1, true))
	require.NoError(t, dirty.Close())
	before := dbFiles(t, dirty.Path)

	out, stopped = runCheck(t, cfg)
	require.True(t, stopped, out)
	require.NotContains(t, out, "abcdefgh", "a site that opens does not stop the start")
	require.Contains(t, out, "site bcdefghi (live): its DB is missing, so the start stops unless its batch's backup in S3 has it\n"+
		"  backup in the data folder: "+backupPath+", made 2026-11-06T10:00:00Z\n")
	require.Regexp(t, `site cdefghij \(disabled\): cannot open: .*file is not a database.*\n`+
		`  backup in the data folder: none\n`, out)
	require.Regexp(t, `site defghijk \(live\): sitedb migration 1 failed: .*already exists.*\n`, out)

	require.Equal(t, before, dbFiles(t, dirty.Path), "the check migrated a copy, so the site's DB is not changed")

	// A control DB that cannot be opened stops the start before any site
	require.NoError(t, control.Close())
	require.NoError(t, os.WriteFile(controldb.GetPath(cfg), []byte("not a database"), 0o600))
	require.NoError(t, utils.DeleteFile(controldb.GetPath(cfg)+"-wal"))
	out, stopped = runCheck(t, cfg)
	require.True(t, stopped, out)
	require.Regexp(t, `^control DB: cannot open: .*file is not a database`, out)
	require.NotContains(t, out, "site ")
}
