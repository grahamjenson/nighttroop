package testutils

import (
	"context"
	"os"
	"path/filepath"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/backup"
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
		err := backup.SqliteBackup(context.TODO(), controlDB.SQLiteDB, backupPath)
		if err != nil {
			t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved: %v", err)
			return
		}

		backupPath = filepath.Join(dumpFolder, "jobs.sqlite")
		err = backup.SqliteBackup(context.TODO(), controlDB.Jobs.SQLiteDB, backupPath)
		if err != nil {
			t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved for jobs DB: %v", err)
			return
		}

		for _, site := range controlDB.AllSites() {
			if site.Subdomain == "" {
				continue
			}

			backupPath = filepath.Join(dumpFolder, site.Subdomain+".sqlite")
			err := backup.SqliteBackup(context.TODO(), site.SQLiteDB, backupPath)
			if err != nil {
				t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved for site %s: %v", site.Subdomain, err)
				return
			}
		}
		t.Logf("🛑 Test failed! In-memory DB snapshot saved to: %s", backupPath)
	})
}

func DumpJobsOnFailure(t interface {
	Helper()
	Failed() bool
	Logf(format string, args ...any)
	Name() string
	Cleanup(func())
}, jobs *db.JobsDB, filePrefix string) {
	t.Helper()

	dumpFolder := filepath.Join(utils.FindProjectRoot(), ".test/dumps/", filePrefix+"_"+t.Name())
	// delete folder if it exists
	_ = os.RemoveAll(dumpFolder)

	t.Cleanup(func() {

		// Only run if the test actually failed
		if !t.Failed() {
			return
		}

		utils.MkDirs(dumpFolder)
		backupPath := filepath.Join(dumpFolder, "jobs.sqlite")
		err := backup.SqliteBackup(context.TODO(), jobs.SQLiteDB, backupPath)
		if err != nil {
			t.Logf("🛑 Test failed! In-memory DB snapshot could not be saved for jobs DB: %v", err)
			return
		}
		t.Logf("🛑 Test failed! In-memory DB snapshot saved to: %s", backupPath)
	})
}
