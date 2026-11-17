package db

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Why a site is unavailable: its DB file was missing at start, or did not open, or a restore could not bring it back,
// as S3 is not working, its backup is not there, or the download failed
const (
	UnavailableMissing       = "missing"
	UnavailableCannotOpen    = "cannot_open"
	UnavailableS3NotWorking  = "s3_not_working"
	UnavailableNotInBackup   = "not_in_backup"
	UnavailableRestoreFailed = "restore_failed"
)

// ErrSiteUnavailable is a live site not served now, as its DB is not here or does not open
var ErrSiteUnavailable = errors.New("site is temporarily unavailable")

// ErrSiteRestoring is a site being restored already, which is not tried twice at once
var ErrSiteRestoring = errors.New("site is being restored")

// RestoreBackoff is how long an unavailable site waits before each try to restore it again: about an hour in all
var RestoreBackoff = []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute, 32 * time.Minute}

// restoreParallelism is how many unavailable sites are restored at once, as a new server restores every site
const restoreParallelism = 16

// UnavailableSite is a site not served, as its DB is not here or does not open, for Reason since Since. TriesLeft is
// how many more times it is restored, the next at NextTry, nil when none is to come.
type UnavailableSite struct {
	ID        string
	Reason    string
	Since     time.Time
	TriesLeft int
	NextTry   *time.Time
}

type unavailableSite struct {
	UnavailableSite
	// wasLive is whether it was live, so its requests are told it is unavailable; a disabled site's, that it is not there
	wasLive bool
	// cancel stops its tries in the background, and restoring is held while it is tried, so it is never tried twice
	cancel    context.CancelFunc
	restoring sync.Mutex
}

// unavailableSites are the sites not served, by ID
type unavailableSites struct {
	mu     sync.Mutex
	bySite map[string]*unavailableSite
}

func (u *unavailableSites) get(siteID string) *unavailableSite {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.bySite[siteID]
}

// UnavailableSites are the sites not served now, the longest unavailable first
func (db *ControlDB) UnavailableSites() []UnavailableSite {
	db.unavailable.mu.Lock()
	defer db.unavailable.mu.Unlock()
	sites := []UnavailableSite{}
	for _, site := range db.unavailable.bySite {
		sites = append(sites, site.UnavailableSite)
	}
	slices.SortFunc(sites, func(a, b UnavailableSite) int { return cmp.Or(a.Since.Compare(b.Since), strings.Compare(a.ID, b.ID)) })
	return sites
}

// MakeSiteUnavailable stops serving a live or disabled site for reason, closing its DB if it is open, until a restore
// brings it back
func (db *ControlDB) MakeSiteUnavailable(ctx context.Context, siteID, reason string) error {
	row, err := db.Queries.MarkSiteUnavailable(ctx, controldb.MarkSiteUnavailableParams{ID: siteID, Reason: &reason})
	if err != nil {
		return fmt.Errorf("cannot make site %s unavailable: %w", siteID, err)
	}
	if site := db.sites.remove(siteID); site != nil {
		site.Close()
	}
	db.unavailable.mu.Lock()
	defer db.unavailable.mu.Unlock()
	site := db.unavailable.bySite[siteID]
	if site == nil {
		site = &unavailableSite{}
		db.unavailable.bySite[siteID] = site
	}
	site.ID, site.Reason, site.Since, site.wasLive = siteID, reason, time.UnixMilli(*row.UnavailableAt), row.DisabledAt == nil
	return nil
}

// makeSiteAvailable serves the site's DB again, live or disabled as it was, and stops its tries
func (db *ControlDB) makeSiteAvailable(ctx context.Context, siteDB *SiteDB) error {
	if _, err := db.Queries.MarkSiteAvailable(ctx, controldb.MarkSiteAvailableParams{ID: siteDB.Subdomain}); err != nil {
		siteDB.Close()
		return fmt.Errorf("cannot make site %s available: %w", siteDB.Subdomain, err)
	}
	db.sites.add(siteDB)
	db.unavailable.mu.Lock()
	defer db.unavailable.mu.Unlock()
	if site := db.unavailable.bySite[siteDB.Subdomain]; site != nil && site.cancel != nil {
		site.cancel()
	}
	delete(db.unavailable.bySite, siteDB.Subdomain)
	return nil
}

// RestoreUnavailableSites tries once to restore each unavailable site from its latest backup, downloaded with
// download, a few at once, then tries again in the background, so the start goes on without a site it cannot restore
func (db *ControlDB) RestoreUnavailableSites(ctx context.Context, download s3utils.DownloadFunc) {
	g := errgroup.Group{}
	g.SetLimit(restoreParallelism)
	for _, site := range db.UnavailableSites() {
		g.Go(func() error {
			if restored, _ := db.tryRestore(ctx, site.ID, download, 0); !restored {
				db.retryRestore(site.ID, download)
			}
			return nil
		})
	}
	_ = g.Wait()
}

// RetryRestore tries to restore an unavailable site now, for an admin, and when it cannot, tries again in the
// background as after the start. It is ErrSiteNotFound for a site that is not unavailable, and ErrSiteRestoring while
// it is being tried.
func (db *ControlDB) RetryRestore(ctx context.Context, siteID string, download s3utils.DownloadFunc) (bool, error) {
	if db.unavailable.get(siteID) == nil {
		return false, ErrSiteNotFound
	}
	restored, err := db.tryRestore(ctx, siteID, download, 0)
	if !restored && err == nil {
		db.retryRestore(siteID, download)
	}
	return restored, err
}

// retryRestore tries to restore the unavailable site after each wait of RestoreBackoff, until it is restored, and pages
// the admin if it never is. It stops the site's tries before.
func (db *ControlDB) retryRestore(siteID string, download s3utils.DownloadFunc) {
	ctx, cancel := context.WithCancel(aslog.WithSite(aslog.NewJob(db.retries, "RestoreUnavailableSite"), siteID))
	db.unavailable.mu.Lock()
	site := db.unavailable.bySite[siteID]
	if site == nil {
		db.unavailable.mu.Unlock()
		cancel()
		return
	}
	if site.cancel != nil {
		site.cancel()
	}
	site.cancel = cancel
	db.unavailable.mu.Unlock()
	go func() {
		defer cancel()
		for try, wait := range db.RestoreBackoff {
			db.nextTry(site, len(db.RestoreBackoff)-try, new(time.Now().Add(wait)))
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
			}
			if restored, _ := db.tryRestore(ctx, siteID, download, try+1); restored {
				return
			}
		}
		if reason := db.nextTry(site, 0, nil); ctx.Err() == nil {
			db.Page(ctx, "A site cannot be restored, so it is unavailable", "site_id", siteID, "reason", reason)
		}
	}()
}

// nextTry records the site's tries left and when the next is, and is why it is unavailable
func (db *ControlDB) nextTry(site *unavailableSite, left int, at *time.Time) string {
	db.unavailable.mu.Lock()
	defer db.unavailable.mu.Unlock()
	site.TriesLeft, site.NextTry = left, at
	return site.Reason
}

// tryRestore tries once to restore the unavailable site from its latest backup, downloaded with download, and opens it.
// A failure, its try'th, is logged with its error, so Sentry has it, and recorded as an event, and leaves it unavailable
// for why.
func (db *ControlDB) tryRestore(ctx context.Context, siteID string, download s3utils.DownloadFunc, try int) (bool, error) {
	site := db.unavailable.get(siteID)
	if site == nil {
		return true, nil
	}
	if !site.restoring.TryLock() {
		return false, ErrSiteRestoring
	}
	defer site.restoring.Unlock()
	ctx = aslog.WithSite(ctx, siteID)
	reason, err := UnavailableS3NotWorking, error(nil)
	if db.ServerConfig.Status.S3Working {
		reason, err = db.restoreSite(ctx, siteID, download)
	}
	if reason == "" {
		slog.InfoContext(ctx, "Unavailable site restored from its backup", "site_id", siteID, "try", try)
		db.Logs.SiteAvailableEvent(ctx, siteID, try)
		return true, nil
	}
	slog.ErrorContext(ctx, "Cannot restore an unavailable site from its backup", "site_id", siteID, "reason", reason, "try", try, "error", err)
	db.Logs.SiteUnavailableEvent(ctx, siteID, reason, try)
	if err := db.MakeSiteUnavailable(ctx, siteID, reason); err != nil {
		slog.ErrorContext(ctx, "Cannot record why a site is unavailable", "error", err)
	}
	return false, nil
}

// restoreSite downloads the site's latest backup beside its folder, moves its DB into place and opens it, or is why it
// could not. A restored DB that does not open is moved aside.
func (db *ControlDB) restoreSite(ctx context.Context, siteID string, download s3utils.DownloadFunc) (string, error) {
	path := sitedb.GetPath(siteID, db.ServerConfig)
	folder := filepath.Dir(path)
	if err := utils.MkDirs(folder); err != nil {
		return UnavailableRestoreFailed, err
	}
	// Named like a restore's staging folder, so one left by a start killed part way is cleaned up like those
	dir, err := os.MkdirTemp(filepath.Dir(folder), "."+filepath.Base(folder)+"-restore-")
	if err != nil {
		return UnavailableRestoreFailed, fmt.Errorf("cannot create staging dir: %w", err)
	}
	defer os.RemoveAll(dir)
	// Downloaded to a path in dir that is not there yet, so the download's own staging folder, made beside it, is in
	// dir too and removed with it
	downloaded := filepath.Join(dir, "site")
	if _, found, err := download(ctx, sitedb.GetS3SitePath(siteID, db.ServerConfig), downloaded); err != nil {
		return UnavailableRestoreFailed, err
	} else if !found {
		return UnavailableNotInBackup, nil
	}
	// A rename on the same disk, so the site's DB is there whole or not at all
	if err := os.Rename(filepath.Join(downloaded, filepath.Base(path)), path); err != nil {
		return UnavailableRestoreFailed, fmt.Errorf("the backup has no DB: %w", err)
	}
	siteDB, err := NewSiteDB(ctx, siteID, db.ServerConfig)
	if err != nil {
		return UnavailableCannotOpen, errors.Join(err, moveAside(path))
	}
	if err := db.makeSiteAvailable(ctx, siteDB); err != nil {
		return UnavailableRestoreFailed, err
	}
	return "", nil
}

// moveAside renames a site's DB file that does not open, with its write-ahead log, to <file>.broken-<time in ms>, so it
// is kept to look into and a restore can take its place
func moveAside(path string) error {
	suffix := fmt.Sprintf(".broken-%d", time.Now().UnixMilli())
	for _, file := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.Rename(file, file+suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cannot move a broken site DB aside: %w", err)
		}
	}
	return nil
}
