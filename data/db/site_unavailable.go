package db

import (
	"context"
	"errors"
	"maps"
	"os"
	"slices"
	"sync"
	"time"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Why a site is unavailable
const (
	// UnavailableMissing is a site whose DB file was not there at start
	UnavailableMissing = "missing"
	// UnavailableCannotOpen is a site whose DB file was there at start but did not open, as when its migration fails
	UnavailableCannotOpen = "cannot_open"
)

// ErrSiteUnavailable is a live or disabled site whose DB is not open, as it was missing or did not open at start, so its
// requests are refused for now rather than answered as a site there is not
var ErrSiteUnavailable = errors.New("site unavailable")

// unavailableSites are the sites whose DB is not open, by site ID with why, until a restore from its backup opens it
type unavailableSites struct {
	mu      sync.RWMutex
	reasons map[string]string
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
func (db *ControlDB) KeepRestoringUnavailableSites(ctx context.Context, download s3utils.DownloadFunc, retries []time.Duration, page page.Func) {
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
		page(ctx, "Sites are unavailable, as their DBs could not be restored", "site_ids", unavailable, "tries", len(retries)+1)
	}
}
