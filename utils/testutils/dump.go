package testutils

import (
	"context"
	"os"
	"path/filepath"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
)

// DumpOnFailure registers a cleanup hook that saves the in-memory DB
// to disk ONLY if the test fails.
func DumpOnFailure(t interface {
	Helper()
	Failed() bool
	Logf(format string, args ...any)
	Name() string
	Cleanup(func())
}, controlDB *db.ControlDB, filePrefix string) {
	t.Helper()

	dumpFolder := filepath.Join(utils.FindProjectRoot(), ".test/dumps/", filePrefix+"_"+t.Name())
	// delete folder if it exists
	_ = os.RemoveAll(dumpFolder)

	t.Cleanup(func() {

		// Only run if the test actually failed
		if !t.Failed() {
			return
		}

		// Ensure output directory exists (e.g., ./debug_dumps)

		utils.MkDirs(dumpFolder)
		backupPath := filepath.Join(dumpFolder, "control.sqlite")
		err := utils.SqliteBackup(context.TODO(), controlDB.SQLiteDB, backupPath)
		if err != nil {
			t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved: %v", err)
			return
		}

		for _, site := range controlDB.AllSites() {
			if site.Subdomain == "" {
				continue
			}
			_, backupPath := utils.SiteDBPaths(site.Subdomain)
			backupPath = filepath.Join(dumpFolder, site.Subdomain+".sqlite")
			err := utils.SqliteBackup(context.TODO(), site.SQLiteDB, backupPath)
			if err != nil {
				t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved for site %s: %v", site.Subdomain, err)
				return
			}
		}
		t.Logf("🛑 Test failed! In-memory DB snapshot saved to: %s", backupPath)
	})
}
