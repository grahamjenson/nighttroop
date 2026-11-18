package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/migrateup"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	pager "nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Why a site is unavailable
const (
	// UnavailableMissing is a site whose DB file was not there at start
	UnavailableMissing = "missing"
	// UnavailableCannotOpen is a site whose DB file was there at start but did not open
	UnavailableCannotOpen = "cannot_open"
	// UnavailableMigrationFailed is a site whose DB opened but could not be migrated to the latest migration, as one
	// restored from long ago whose migration fails
	UnavailableMigrationFailed = "migration_failed"
)

// ErrSiteUnavailable is a live or disabled site whose DB is not open, as it was missing or did not open at start, so its
// requests are refused for now rather than answered as a site there is not
var ErrSiteUnavailable = errors.New("site unavailable")

// unavailableSites are the sites whose DB is not open, by site ID with why, until a restore from its backup opens it
type unavailableSites struct {
	mu      sync.RWMutex
	reasons map[string]string
	// busy are the unavailable sites whose DB is being restored or made again now, so two never work on one at once
	busy map[string]bool
	// failedMigrations are the migrations that failed on the sites unavailable as UnavailableMigrationFailed
	failedMigrations map[string]uint
}

// claim marks the unavailable site busy, and is false if it is not unavailable or is already busy
func (u *unavailableSites) claim(siteID string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	if _, ok := u.reasons[siteID]; !ok || u.busy[siteID] {
		return false
	}
	if u.busy == nil {
		u.busy = map[string]bool{}
	}
	u.busy[siteID] = true
	return true
}

func (u *unavailableSites) release(siteID string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.busy, siteID)
}

func (u *unavailableSites) set(siteID, reason string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.reasons == nil {
		u.reasons = map[string]string{}
	}
	u.reasons[siteID] = reason
}

func (u *unavailableSites) remove(siteID string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.reasons, siteID)
	delete(u.failedMigrations, siteID)
}

func (u *unavailableSites) has(siteID string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	_, ok := u.reasons[siteID]
	return ok
}

// UnavailableSites are the unavailable sites, by site ID with why
func (db *ControlDB) UnavailableSites() map[string]string {
	db.unavailable.mu.RLock()
	defer db.unavailable.mu.RUnlock()
	sites := make(map[string]string, len(db.unavailable.reasons))
	for siteID, reason := range db.unavailable.reasons {
		sites[siteID] = reason
	}
	return sites
}

// MakeSiteUnavailable closes the site's DB, if it is open, and keeps it unavailable for reason, so its requests are
// refused for now and the jobs leave it out, until a restore from its backup opens it
func (db *ControlDB) MakeSiteUnavailable(siteID, reason string) {
	db.unavailable.set(siteID, reason)
	if site := db.sites.remove(siteID); site != nil {
		site.Close()
	}
}

// MakeSiteUnavailableFor makes the site unavailable for why its DB did not open: UnavailableMigrationFailed, keeping
// the migration that failed, when it could not be migrated, or UnavailableCannotOpen otherwise
func (db *ControlDB) MakeSiteUnavailableFor(siteID string, err error) {
	var failed *migrateup.FailedError
	if !errors.As(err, &failed) {
		db.MakeSiteUnavailable(siteID, UnavailableCannotOpen)
		return
	}
	db.unavailable.mu.Lock()
	if db.unavailable.failedMigrations == nil {
		db.unavailable.failedMigrations = map[string]uint{}
	}
	db.unavailable.failedMigrations[siteID] = failed.Migration
	db.unavailable.mu.Unlock()
	db.MakeSiteUnavailable(siteID, UnavailableMigrationFailed)
}

// SiteMigration is the migration the site's DB is at, or for one unavailable as its migration failed, the migration
// that failed, with failed true. It is false for a site whose DB is not open for another reason.
func (db *ControlDB) SiteMigration(siteID string) (version uint, failed bool, ok bool) {
	db.unavailable.mu.RLock()
	version, failed = db.unavailable.failedMigrations[siteID]
	db.unavailable.mu.RUnlock()
	site := db.sites.get(siteID)
	if failed || site == nil {
		return version, failed, failed
	}
	version, dirty, err := migrateup.Version(site.SQLiteDB, sitedb.Migrations)
	return version, dirty, err == nil
}

// dbFiles are a SQLite DB's files: the DB and its write-ahead log and shared memory
func dbFiles(path string) []string {
	return []string{path, path + "-wal", path + "-shm"}
}

// moveDB renames the DB at from, with its write-ahead log and shared memory, to to, leaving out any not there
func moveDB(from, to string) error {
	for i, file := range dbFiles(from) {
		if !utils.FileExists(file) {
			continue
		}
		if err := os.Rename(file, dbFiles(to)[i]); err != nil {
			return err
		}
	}
	return nil
}

// RestoreRetries are how long to wait before each new try at restoring the sites still unavailable: about an hour in
// all, longer each time, so a short S3 outage is waited out without hammering it
var RestoreRetries = []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute, 32 * time.Minute}

// KeepRestoringUnavailableSites restores the unavailable sites, then tries those still unavailable again after each of
// retries, and pages the admin with page if any are still unavailable after the last. Each failed try is logged as an
// error and a site_db_missing event. It stops when ctx ends, as the server stops.
func (db *ControlDB) KeepRestoringUnavailableSites(ctx context.Context, download s3utils.DownloadFunc, retries []time.Duration, page pager.Func) {
	db.RestoreUnavailableSites(ctx, download)
	for _, wait := range retries {
		if len(db.UnavailableSites()) == 0 {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		db.RestoreUnavailableSites(ctx, download)
	}
	if unavailable := slices.Sorted(maps.Keys(db.UnavailableSites())); len(unavailable) > 0 && ctx.Err() == nil {
		page(ctx, pager.Call, "Sites are unavailable, as their DBs could not be restored", "site_ids", unavailable, "tries", len(retries)+1)
	}
}

// ErrSiteNotUnavailable is a site whose DB is not unavailable, or is being restored now, so it is not made again
var ErrSiteNotUnavailable = errors.New("the site is not unavailable, or its DB is being restored")

// NewEmptySiteDB makes an unavailable site's DB again, empty, and opens it, so the site is served again without its
// data, as for one created just before a crash, which has no backup. Its identities, in the control DB, are kept, and
// a DB there that did not open is moved aside and kept. The site counts as written, so its new DB is backed up.
func (db *ControlDB) NewEmptySiteDB(ctx context.Context, siteID string) error {
	if !db.unavailable.claim(siteID) {
		return ErrSiteNotUnavailable
	}
	defer db.unavailable.release(siteID)
	path := sitedb.GetPath(siteID, db.ServerConfig)
	if err := utils.MkDirs(filepath.Dir(path)); err != nil {
		return err
	}
	if err := moveDB(path, fmt.Sprintf("%s.unopened-%d", path, time.Now().UnixMilli())); err != nil {
		return fmt.Errorf("cannot move the DB that did not open aside: %w", err)
	}
	siteDB, err := NewSiteDB(ctx, siteID, db.ServerConfig)
	if err != nil {
		return fmt.Errorf("cannot make a new DB: %w", err)
	}
	db.sites.add(siteDB)
	db.unavailable.remove(siteID)
	now := time.Now().UnixMilli()
	return db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastWriteAt: &now, ID: siteID})
}

// UploadSiteDB puts a DB a control admin uploaded, such as one they downloaded and fixed, in place of an unavailable
// site's DB, and opens it, migrating it, so the site is served again. As in a restore, a DB there that did not open is
// kept, and put back if the uploaded one is ErrDBDoesNotOpen. The site counts as written, so its new DB is backed up.
func (db *ControlDB) UploadSiteDB(ctx context.Context, siteID string, upload io.Reader) error {
	if !db.unavailable.claim(siteID) {
		return ErrSiteNotUnavailable
	}
	defer db.unavailable.release(siteID)
	folder := filepath.Dir(sitedb.GetPath(siteID, db.ServerConfig))
	if err := utils.MkDirs(folder); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(folder), "."+filepath.Base(folder)+"-restore-") // as a restore's, on its disk
	if err != nil {
		return fmt.Errorf("cannot create staging dir: %w", err)
	}
	defer os.RemoveAll(dir)
	staged := filepath.Join(dir, siteID+".sqlite")
	file, err := os.Create(staged)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, upload)
	if err = errors.Join(err, file.Close()); err != nil {
		return fmt.Errorf("cannot save the uploaded DB: %w", err)
	}
	if err := db.replaceSiteDB(ctx, siteID, staged); err != nil {
		return err
	}
	db.unavailable.remove(siteID)
	now := time.Now().UnixMilli()
	return db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastWriteAt: &now, ID: siteID})
}

// RestoreUnavailableSite restores the unavailable site from its backup, downloaded with download, such as an older
// version of it a control admin picked, and is whether it is served again. A DB there that did not open is kept, as in
// RestoreUnavailableSites. A site that is not unavailable, or is being restored now, is ErrSiteNotUnavailable.
func (db *ControlDB) RestoreUnavailableSite(ctx context.Context, siteID string, download s3utils.DownloadFunc) (bool, error) {
	if !db.unavailable.claim(siteID) {
		return false, ErrSiteNotUnavailable
	}
	defer db.unavailable.release(siteID)
	found, err := db.restoreSite(ctx, siteID, download)
	if err != nil || !found {
		return false, err
	}
	db.unavailable.remove(siteID)
	return true, nil
}
