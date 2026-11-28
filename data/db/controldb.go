package db

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/thin"
)

type ControlDB struct {
	*controldb.DB

	sites openSites
	// createSiteLock serialises CreateSite and HydrateSite so two calls never pick, open or delete the same subdomain's DB
	createSiteLock sync.Mutex

	Jobs *JobsDB
	Logs *LogsDB
	// Live is what the server keeps only while it runs, like each identity's alarms
	Live *livedb.DB

	setup setup
	// unavailable are the live and disabled sites whose DB is not open, as it was missing or did not open at start
	unavailable unavailableSites
	// measured are the server's own DBs' tables as last measured, by DB name, to *measuredTables, kept in memory as those
	// DBs have no row of their own to keep them in
	measured sync.Map
	// replica replicates the control DB once Replicate starts it. Nil, as without S3 or in tests, replicates nothing.
	replica *replica.Replica
	// diskFull is whether the DBs' disk had less than DiskRefuseFreeMB free at the last check of the server's resources
	diskFull atomic.Bool
}

// DiskFull is whether the DBs' disk was nearly full at the last check, so sites are not created and their writes that
// add data are refused, as a full disk would stop every site
func (db *ControlDB) DiskFull() bool { return db.diskFull.Load() }

// SetDiskFull keeps whether the DBs' disk is nearly full, and is whether that changed
func (db *ControlDB) SetDiskFull(full bool) bool { return db.diskFull.Swap(full) != full }

// Replicate replicates the control DB to S3 with control, as it is written
func (db *ControlDB) Replicate(ctx context.Context, control *replica.Replica) error {
	if err := control.Start(ctx, db.Path); err != nil {
		return err
	}
	db.replica = control
	return nil
}

func NewControlDB(ctx context.Context, serverConfig *config.ServerConfig) (*ControlDB, error) {

	rawDB, err := controldb.NewDB(ctx, serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot open control db: %w", err)
	}

	jobs, err := NewJobsDB(ctx, serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot open jobs db: %w", err)
	}

	logs, err := NewLogsDB(ctx, serverConfig)
	if err != nil {
		return nil, err
	}

	live, err := livedb.NewDB(ctx, serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot open live db: %w", err)
	}

	// Create a new control db
	controlDB := &ControlDB{
		DB:    rawDB,
		Jobs:  jobs,
		Logs:  logs,
		Live:  live,
		sites: openSites{bySubdomain: map[string]*SiteDB{}},
	}

	// Without a control admin the server starts in setup, see control_setup.go
	if err := controlDB.startSetupIfNoAdmin(ctx); err != nil {
		return nil, fmt.Errorf("cannot check for the control admin: %w", err)
	}

	if err := controlDB.LoadRateLimits(ctx); err != nil {
		return nil, fmt.Errorf("cannot load the rate limits: %w", err)
	}

	// This will go over all non-deleted sites and add them to the control db sites
	if err := controlDB.fillSites(ctx); err != nil {
		return nil, fmt.Errorf("cannot fill sites: %w", err)
	}

	return controlDB, nil
}

// fillSites opens every live and disabled site's DB. Opening one is mostly SQLite reading its schema, which with
// thousands of sites is most of the server's start, so a site is opened on each core at once.
func (db *ControlDB) fillSites(ctx context.Context) error {
	res, err := db.Queries.GetLiveAndDisabledSites(ctx)
	if err != nil {
		return fmt.Errorf("cannot query sites: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())
	for _, model := range res {
		// A site without its DB, or whose DB does not open, is unavailable, so the start goes on without it, and is
		// restored from its backup by RestoreUnavailableSites
		if !utils.FileExists(sitedb.GetPath(model.ID, db.ServerConfig)) {
			slog.ErrorContext(aslog.WithSite(ctx, model.ID), "Site DB is missing, so the site is unavailable", "site_id", model.ID)
			db.MakeSiteUnavailable(model.ID, UnavailableMissing)
			continue
		}
		g.Go(func() error {
			siteDB, err := NewSiteDB(ctx, model.ID, db.ServerConfig)
			if err != nil {
				slog.ErrorContext(aslog.WithSite(ctx, model.ID), "Site DB does not open, so the site is unavailable", "site_id", model.ID, "error", err)
				db.MakeSiteUnavailableFor(model.ID, err)
				return nil
			}
			db.sites.add(siteDB)
			return nil
		})
	}
	return g.Wait()
}

// restoreParallelism is how many unavailable sites are restored at once, as a new server restores every site
const restoreParallelism = 16

// RestoreUnavailableSites restores each unavailable site from its latest backup, downloaded with download, a few at
// once, and opens it. Each is a site_db_missing event, and one not restored is logged as an error, so the control admin
// sees it. One not restored stays unavailable while the server runs without it.
func (db *ControlDB) RestoreUnavailableSites(ctx context.Context, download s3utils.DownloadFunc) {
	g := errgroup.Group{}
	g.SetLimit(restoreParallelism)
	for siteID := range db.UnavailableSites() {
		g.Go(func() error {
			// A server stopping before every site is restored leaves the rest to its next start
			if ctx.Err() != nil {
				return nil
			}
			// One whose DB is being made again now, or was already, is left to that
			if !db.unavailable.claim(siteID) {
				return nil
			}
			defer db.unavailable.release(siteID)
			siteCtx := aslog.WithSite(ctx, siteID)
			// Without S3 there is nothing to download, so it does not wait on a download that cannot work
			reason := ""
			if !db.ServerConfig.Status.S3Working {
				reason = "s3_not_working"
			} else if found, err := db.restoreSite(siteCtx, siteID, download); err != nil {
				slog.ErrorContext(siteCtx, "Cannot restore an unavailable site's DB from its backup", "site_id", siteID, "error", err)
				reason = "restore_failed"
				// A backup that cannot be migrated, as one from long ago, or whose schema is not the migrations', says so
				if why := unavailableReason(err); why != UnavailableCannotOpen {
					reason = why
					db.MakeSiteUnavailableFor(siteID, err)
				}
			} else if !found {
				reason = "not_in_backup"
			}
			if reason == "" {
				db.unavailable.remove(siteID)
				slog.InfoContext(siteCtx, "Unavailable site's DB restored from its backup", "site_id", siteID)
			} else {
				slog.ErrorContext(siteCtx, "Site is unavailable, as its DB could not be restored", "site_id", siteID, "reason", reason)
			}
			db.Logs.SiteDBMissingEvent(siteCtx, siteID, reason == "", reason)
			return nil
		})
	}
	_ = g.Wait()
}

// restoreSite downloads the site's latest backup beside its folder, puts its DB in place and opens it, and is whether
// it had a backup. A DB already there, which did not open, is moved aside first, and put back if the restored one does
// not open either, so a DB a later release can open, as one whose migration failed, is never lost.
func (db *ControlDB) restoreSite(ctx context.Context, siteID string, download s3utils.DownloadFunc) (bool, error) {
	path := sitedb.GetPath(siteID, db.ServerConfig)
	folder := filepath.Dir(path)
	if err := utils.MkDirs(folder); err != nil {
		return false, err
	}
	// Named like a restore's staging folder, so one left by a start killed part way is cleaned up like those
	dir, err := os.MkdirTemp(filepath.Dir(folder), "."+filepath.Base(folder)+"-restore-")
	if err != nil {
		return false, fmt.Errorf("cannot create staging dir: %w", err)
	}
	defer os.RemoveAll(dir)
	// Downloaded to a path in dir that is not there yet, so the download's own staging folder, made beside it, is in
	// dir too and removed with it
	downloaded := filepath.Join(dir, "site")
	if _, found, err := download(ctx, sitedb.GetS3SitePath(siteID, db.ServerConfig), downloaded); err != nil || !found {
		return false, err
	}
	return true, db.replaceSiteDB(ctx, siteID, filepath.Join(downloaded, filepath.Base(path)))
}

// ErrDBDoesNotOpen is a DB put in place of a site's that does not open or migrate, so the one there is put back
var ErrDBDoesNotOpen = errors.New("the DB does not open or migrate")

// replaceSiteDB puts the DB at staged, on the same disk, in place of the site's, and opens it. A DB there that did not
// open is moved aside first, and put back if the one at staged does not open either.
func (db *ControlDB) replaceSiteDB(ctx context.Context, siteID string, staged string) error {
	path := sitedb.GetPath(siteID, db.ServerConfig)
	aside := fmt.Sprintf("%s.unopened-%d", path, time.Now().UnixMilli())
	if err := moveDB(path, aside); err != nil {
		return fmt.Errorf("cannot move the DB that did not open aside: %w", err)
	}
	// A rename on the same disk, so the site's DB is there whole or not at all
	if err := os.Rename(staged, path); err != nil {
		return errors.Join(fmt.Errorf("there is no DB to put in place: %w", err), moveDB(aside, path))
	}
	siteDB, err := NewSiteDB(ctx, siteID, db.ServerConfig)
	if err != nil {
		return errors.Join(fmt.Errorf("%w: %w", ErrDBDoesNotOpen, err), utils.DeleteFiles(dbFiles(path)), moveDB(aside, path))
	}
	db.sites.add(siteDB)
	return nil
}

///
// Query Methods
///

type SiteCreateMaterial struct {
	Subdomain     string
	BackupKeyHash string
	Password      string
}

type SiteCreateResult struct {
	SiteDB          *SiteDB
	AdminPassword   string
	TotpSecret      string
	PwdNSToken      string
	ReadOnlyNSToken string
	BackupPassword  string
}

// ErrSiteIDTaken is a site ID picked for a new site that a site, in any state, already has
var ErrSiteIDTaken = fmt.Errorf("that site ID is taken")

func (db *ControlDB) CreateSite(ctx context.Context) (*SiteCreateResult, error) {
	return db.createSite(ctx, "")
}

// CreateSiteWithID creates a site at a site ID picked for it, such as a vanity one for a special user. It is 8
// lowercase letters like every site's, and one already taken is refused.
func (db *ControlDB) CreateSiteWithID(ctx context.Context, subdomain string) (*SiteCreateResult, error) {
	if !utils.SyntacticallyValidateSiteID(subdomain) {
		return nil, fmt.Errorf("site ID %q is not %d lowercase letters", subdomain, utils.SITE_ID_LENGTH)
	}
	return db.createSite(ctx, subdomain)
}

// createSite creates a site at subdomain, or at a random one when it is ""
func (db *ControlDB) createSite(ctx context.Context, subdomain string) (*SiteCreateResult, error) {
	db.createSiteLock.Lock()
	defer db.createSiteLock.Unlock()

	if subdomain != "" {
		if db.SubdomainUsed(ctx, subdomain) {
			return nil, ErrSiteIDTaken
		}
	} else {
		subdomain = utils.NextSiteSubdomain(db.ServerConfig.IsDevOrTest())
		if db.SubdomainUsed(ctx, subdomain) {
			// Try Again with an actual random ID
			subdomain = utils.NextSiteSubdomain(false)
			if db.SubdomainUsed(ctx, subdomain) {
				return nil, fmt.Errorf("site already exists")
			}
		}
	}

	backupPassword, backupKeyHash, err := security.SecureCredentialsWithPrefix(subdomain+":", db.ServerConfig.GetBcryptCost())
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	siteDB, err := NewSiteDB(ctx, subdomain, db.ServerConfig)
	if err != nil {
		// Nothing of the new site was kept, so its half made DB file goes too
		_ = utils.DeleteFiles([]string{sitedb.GetPath(subdomain, db.ServerConfig), sitedb.GetBackupPath(subdomain, db.ServerConfig)})
		return nil, fmt.Errorf("Site DB Creation Error: %w", err)
	}

	// The open sites are held until the site's DB is added, so its row is never found without its DB open
	db.sites.mu.Lock()
	defer db.sites.mu.Unlock()

	adminIdentity, pwdIdentity, readOnlyIdentity, err := db.createSiteRows(ctx, subdomain, backupKeyHash)
	if err != nil {
		// Nothing of the site was kept, so its new DB file goes too
		siteDB.TryCloseAndDelete()
		return nil, err
	}

	db.sites.bySubdomain[subdomain] = siteDB
	db.Logs.SiteCreatedEvent(ctx, subdomain)
	// Its tables are measured now, a quick read of so new a DB, so its pages show them before the resource check runs
	if err := db.RecordSiteTableSizes(ctx, []*SiteDB{siteDB}); err != nil {
		slog.ErrorContext(ctx, "Cannot measure a new site's tables", "site_id", subdomain, "error", err)
	}

	return &SiteCreateResult{
		SiteDB:          siteDB,
		AdminPassword:   adminIdentity.Password,
		TotpSecret:      *adminIdentity.TotpSecret,
		PwdNSToken:      *pwdIdentity.NSToken,
		ReadOnlyNSToken: *readOnlyIdentity.NSToken,
		BackupPassword:  backupPassword,
	}, nil
}

////
// Delete USER methods
////

// DeleteSites marks the sites deleted, then closes and removes each one's DB files. A file that cannot be removed is
// only logged, so a failure part way leaves the sites marked deleted with some files left behind.
func (db *ControlDB) DeleteSites(ctx context.Context, subdomains []string) error {
	// Set the deleted at timestammp in the control DB
	if len(subdomains) == 0 {
		return nil
	}
	bytes, _ := json.Marshal(subdomains)
	deletedSubdomains, err := db.Queries.DeleteSites(ctx, controldb.DeleteSitesParams{
		JsonEach: string(bytes),
	})

	if err != nil {
		return fmt.Errorf("cannot delete site: %w", err)
	}

	slog.Info("Deleting sites", "requested", subdomains, "deleted", deletedSubdomains)

	// Its identities, with their credentials and their devices' push tokens, and its push credentials go with
	// it; an archived site keeps them, as it can be brought back
	for _, subdomain := range deletedSubdomains {
		if err := db.Queries.DeleteSitePushCredentials(ctx, controldb.DeleteSitePushCredentialsParams{SiteID: &subdomain}); err != nil {
			return fmt.Errorf("cannot delete site %s's push credentials: %w", subdomain, err)
		}
		if err := db.Queries.DeleteAllSiteIdentities(ctx, controldb.DeleteAllSiteIdentitiesParams{SiteID: subdomain}); err != nil {
			return fmt.Errorf("cannot delete site %s's identities: %w", subdomain, err)
		}
	}

	// Close the DB and remove from the map. An unavailable site has no open DB: it is not restored any more, and
	// the DB files it has go too
	for _, subdomain := range deletedSubdomains {
		files := append(dbFiles(sitedb.GetPath(subdomain, db.ServerConfig)), sitedb.GetBackupPath(subdomain, db.ServerConfig))
		if site := db.sites.remove(subdomain); site != nil {
			site.Close()
		}
		db.unavailable.remove(subdomain)

		err := utils.DeleteFiles(files)
		if err != nil {
			slog.Error("Failed to delete site db files", "subdomain", subdomain, "error", err)
		} else {
			slog.Info("Deleted site db files", "subdomain", subdomain)
		}
	}

	return nil
}

// ArchiveSite backs up a site's DB and uploads it, with upload, to its archive path, then marks the site archived and
// deletes its DB, so it is no longer served or backed up. The copy is kept as its backup here, which HydrateSite
// restores from first, so a server without S3 archives with a nil upload, keeping only that copy.
func (db *ControlDB) ArchiveSite(ctx context.Context, subdomain string, upload s3utils.UploadFunc) error {
	site := db.sites.get(subdomain)
	if site == nil {
		return fmt.Errorf("site %s is not open", subdomain)
	}
	// Backed up to a folder of its own beside the site's backup file, on the same disk, as the backup job can be
	// rewriting that file at the same time
	if err := utils.MkParentDirs(site.BackupPath); err != nil {
		return fmt.Errorf("cannot make a folder to archive site %s: %w", subdomain, err)
	}
	dir, err := os.MkdirTemp(filepath.Dir(site.BackupPath), ".archive-*")
	if err != nil {
		return fmt.Errorf("cannot make a folder to archive site %s: %w", subdomain, err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, filepath.Base(site.BackupPath))
	if err := backup.CompressedBackup(ctx, site.SQLiteDB, path); err != nil {
		return fmt.Errorf("cannot back up site %s: %w", subdomain, err)
	}
	if upload != nil {
		if err := upload(ctx, sitedb.GetS3ArchivePath(subdomain, db.ServerConfig), s3utils.SiteRetention, []string{path}); err != nil {
			return fmt.Errorf("cannot upload the archive of site %s: %w", subdomain, err)
		}
	}
	bytes, _ := json.Marshal([]string{subdomain})
	if _, err := db.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: string(bytes)}); err != nil {
		return fmt.Errorf("cannot archive site %s: %w", subdomain, err)
	}

	db.sites.remove(subdomain)
	site.Close()
	if err := os.Rename(path, site.BackupPath); err != nil {
		slog.ErrorContext(ctx, "Failed to keep the archive of a site here", "subdomain", subdomain, "error", err)
	}
	if err := utils.DeleteFiles(dbFiles(site.Path)); err != nil {
		slog.ErrorContext(ctx, "Failed to delete archived site db files", "subdomain", subdomain, "error", err)
	}
	return nil
}

// HydrateSite brings an archived site back: it restores the site's DB from its backup here, kept when it was archived,
// or else downloads, with download, its archive from S3, opens it, migrating it, and makes the site live again. A nil
// download, as without S3, restores only from the backup here. A DB that does not open or migrate makes it live but
// unavailable, see MakeSiteUnavailableFor.
func (db *ControlDB) HydrateSite(ctx context.Context, subdomain string, download s3utils.DownloadFunc) error {
	db.createSiteLock.Lock()
	defer db.createSiteLock.Unlock()
	if db.sites.get(subdomain) != nil {
		return fmt.Errorf("site %s is already open", subdomain)
	}
	path, backupPath := sitedb.GetPath(subdomain, db.ServerConfig), sitedb.GetBackupPath(subdomain, db.ServerConfig)
	switch {
	case utils.FileExists(backupPath):
		if err := backup.Decompress(backupPath, path); err != nil {
			return fmt.Errorf("cannot restore site %s from its backup here: %w", subdomain, err)
		}
	case download == nil:
		return fmt.Errorf("site %s has no archive here, and S3 is off", subdomain)
	default:
		if _, found, err := download(ctx, sitedb.GetS3ArchivePath(subdomain, db.ServerConfig), filepath.Dir(path)); err != nil {
			return fmt.Errorf("cannot download the archive of site %s: %w", subdomain, err)
		} else if !found || !utils.FileExists(path) {
			return fmt.Errorf("site %s has no archive", subdomain)
		}
	}
	// A DB restored that does not open or migrate, as one archived long ago, still makes the site live again, but
	// unavailable, with its DB kept here, so its page offers the fixes: download, fix and upload it, restore a backup or
	// make a new empty DB
	siteDB, openErr := NewSiteDB(ctx, subdomain, db.ServerConfig)
	result, err := db.Queries.HydrateSite(ctx, controldb.HydrateSiteParams{ID: subdomain})
	if err == nil {
		if rows, _ := result.RowsAffected(); rows == 0 {
			err = fmt.Errorf("site %s is not archived", subdomain)
		}
	}
	if err != nil {
		if siteDB != nil {
			siteDB.Close()
		}
		// The archive is still here or in S3, so the copy restored from it goes
		_ = utils.DeleteFiles(dbFiles(path))
		return err
	}
	if openErr != nil {
		slog.ErrorContext(ctx, "Restored site's DB does not open, so the site is unavailable", "site_id", subdomain, "error", openErr)
		db.MakeSiteUnavailableFor(subdomain, openErr)
		return nil
	}

	db.sites.add(siteDB)
	return nil
}

// TruncateSite runs the job's truncate of one site now (see TruncateSites), looking at all of it and not only at
// what was written since its last backup: it packages its days, truncates it, and records its new size, which it
// returns. It does nothing the job does not do.
func (db *ControlDB) TruncateSite(ctx context.Context, subdomain string) (*int64, error) {
	site := db.sites.get(subdomain)
	if site == nil {
		return nil, fmt.Errorf("site %s is not open", subdomain)
	}
	if err := db.TruncateSites(ctx, []*SiteDB{site}, map[string]int64{subdomain: 0}); err != nil {
		return nil, err
	}
	sizeKB := site.EstimatedSizeKB()
	if err := db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: sizeKB, ID: subdomain}); err != nil {
		return sizeKB, err
	}
	// Its tables are measured again now, for the page that shows its new size
	return sizeKB, db.RecordSiteTableSizes(ctx, []*SiteDB{site})
}

// TruncateSites truncates sites, as the backup job does before copying each, without shrinking their DBs, as SQLite
// uses the space again for what comes next. It looks back truncateLooksBack, or to when a site was last truncated in
// truncatedAt when that is longer ago, 0 for all of it. It truncates harder a site far over its DB size limit. It goes
// on past a site it cannot truncate, returning each one's error.
func (db *ControlDB) TruncateSites(ctx context.Context, sites []*SiteDB, truncatedAt map[string]int64) error {
	limits, err := db.Queries.GetSiteDBSizeLimits(ctx)
	errs := []error{err}
	limitsKB := map[string]int64{}
	for _, limit := range limits {
		limitsKB[limit.ID] = *limit.MaxDbSizeKbLimit
	}
	for _, site := range sites {
		siteCtx := aslog.WithSite(ctx, site.Subdomain)
		since := time.Now().Add(-truncateLooksBack).UnixMilli()
		if at, ok := truncatedAt[site.Subdomain]; ok {
			since = min(since, at)
		}
		err := db.packageSite(siteCtx, site, since, time.Now())
		if err == nil {
			err = truncateSite(siteCtx, site, since)
		}
		if limitKB, ok := limitsKB[site.Subdomain]; ok && err == nil {
			err = db.hardTruncateSite(siteCtx, site, limitKB)
		}
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// hardTruncateSizeFactor is how many times its DB size limit a site's DB must be for the job to truncate it harder, as
// a site can fill its DB on purpose. An honest Loop user grows about 4.4 MB a month after the job's truncate, so takes
// about 7 months to pass it at the default limit of 20 MB.
const hardTruncateSizeFactor = 1.5

const (
	// hardTruncateJSONKeptFor is how long a site over hardTruncateSizeFactor keeps JSON that matters only while
	// recent, hardTruncateReportsKeptFor its device reports, and hardTruncateRecordsKeptFor its readings and treatments
	hardTruncateJSONKeptFor    = 24 * time.Hour
	hardTruncateReportsKeptFor = 7 * 24 * time.Hour
	hardTruncateRecordsKeptFor = 90 * 24 * time.Hour
	// hardTruncatedBy is who the "Site truncated" event of a hard truncate says truncated the site
	hardTruncatedBy = "size limit"
)

// vacuumFreePart is the part of a site's DB that must be free pages for the job to shrink it. A shrink writes the whole
// DB again, and a site that stays over its limit changes a few records at each backup, so less than this is left for
// the next writes to fill.
const vacuumFreePart = 0.1

// vacuumIfMostlyFree shrinks site's DB if vacuumFreePart of it or more is free pages
func vacuumIfMostlyFree(ctx context.Context, site *SiteDB) error {
	free, err := site.FreePart(ctx)
	if err != nil || free < vacuumFreePart {
		return err
	}
	return site.Vacuum(ctx)
}

// hardTruncateSite truncates the site harder while its DB is over hardTruncateSizeFactor times its limit, limitKB.
// First it clears the JSON that matters only while recent, and deletes device reports over 7 days old and the loop
// settings no record refers to. If the site is still over, it deletes its readings and treatments over 90 days old,
// but the newest of each type. After each step that changed records it shrinks the DB if enough of it is free. If any
// changed, it records the new size and logs a "Site truncated" event.
func (db *ControlDB) hardTruncateSite(ctx context.Context, site *SiteDB, limitKB int64) error {
	// A DB is measured without its free pages, as a DB left unshrunk keeps them and the next writes fill them
	over := func(sizeKB *int64) (bool, error) {
		if sizeKB == nil {
			return false, nil
		}
		free, err := site.FreePart(ctx)
		return float64(*sizeKB)*(1-free) > hardTruncateSizeFactor*float64(limitKB), err
	}
	beforeKB := site.EstimatedSizeKB()
	if isOver, err := over(beforeKB); err != nil || !isOver {
		return err
	}
	started, q := time.Now(), site.Queries()
	var changed int64
	count := func(rows int64, err error) error { changed += rows; return err }
	dayAgo := sitedb.ClearOldDeviceReportJSONParams{Before: started.Add(-hardTruncateJSONKeptFor).UnixMilli()}
	weekAgo := sitedb.DeleteOldDeviceReportsParams{Before: started.Add(-hardTruncateReportsKeptFor).UnixMilli()}
	err := errors.Join(
		count(q.ClearOldDeviceReportJSON(ctx, dayAgo)),
		count(q.ClearOldBolusCalculatorJSON(ctx, sitedb.ClearOldBolusCalculatorJSONParams(dayAgo))),
		count(q.ClearOldInsulinDoseJSON(ctx, sitedb.ClearOldInsulinDoseJSONParams(dayAgo))),
		count(q.ClearOldProfileSwitchJSON(ctx, sitedb.ClearOldProfileSwitchJSONParams(dayAgo))),
		count(q.DeleteOldDeviceReports(ctx, weekAgo)),
		count(q.DeleteUnreferencedLoopSettings(ctx)),
	)
	if err == nil && changed > 0 {
		err = vacuumIfMostlyFree(ctx, site)
	}
	stillOver, overErr := over(site.EstimatedSizeKB())
	if err = errors.Join(err, overErr); err == nil && stillOver {
		// The groups go first, as the events they refer to are kept
		old := sitedb.DeleteOldEventGroupsParams{Before: started.Add(-hardTruncateRecordsKeptFor).UnixMilli()}
		step1 := changed
		err = errors.Join(
			count(q.DeleteOldEventGroups(ctx, old)),
			count(q.DeleteOldGlucoseReadings(ctx, sitedb.DeleteOldGlucoseReadingsParams(old))),
			count(q.DeleteOldInsulinDoses(ctx, sitedb.DeleteOldInsulinDosesParams(old))),
			count(q.DeleteOldFoodIntakes(ctx, sitedb.DeleteOldFoodIntakesParams(old))),
			count(q.DeleteOldTempBasals(ctx, sitedb.DeleteOldTempBasalsParams(old))),
			count(q.DeleteOldTempTargets(ctx, sitedb.DeleteOldTempTargetsParams(old))),
			count(q.DeleteOldProfileSwitches(ctx, sitedb.DeleteOldProfileSwitchesParams(old))),
			count(q.DeleteOldActivities(ctx, sitedb.DeleteOldActivitiesParams(old))),
			count(q.DeleteOldNotificationEvents(ctx, sitedb.DeleteOldNotificationEventsParams(old))),
		)
		if err == nil && changed > step1 {
			err = vacuumIfMostlyFree(ctx, site)
		}
	}
	if err != nil {
		return fmt.Errorf("cannot truncate site %s harder: %w", site.Subdomain, err)
	}
	if changed == 0 {
		return nil
	}
	afterKB := site.EstimatedSizeKB()
	db.Logs.SiteTruncatedEvent(ctx, site.Subdomain, hardTruncatedBy, started, beforeKB, afterKB)
	return db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: afterKB, ID: site.Subdomain})
}

// truncateLooksBack is how far back the job looks for device reports written, or loop statuses turned a day old, to
// clear, as those before were cleared by an earlier run; a day covers a server down for a while
const truncateLooksBack = 24 * time.Hour

// DeviceReportsKeptFor is how long a device report is kept at all
const DeviceReportsKeptFor = 30 * 24 * time.Hour

// ThinParams are ThinOldDeviceReports's, for the reports written since since, at now, both in ms, with utils/thin's
// ages and buckets, which a backfill keeps to as well
func ThinParams(since, now int64) sitedb.ThinOldDeviceReportsParams {
	return sitedb.ThinOldDeviceReportsParams{Since: since, Now: now, KeepAllMs: thin.KeepAll.Milliseconds(),
		WeekMs: thin.Week.Milliseconds(), FortnightMs: thin.Fortnight.Milliseconds(), SmallMs: thin.Small.Milliseconds(),
		MediumMs: thin.Medium.Milliseconds(), LargeMs: thin.Large.Milliseconds()}
}

// truncateSite clears the predictions from each device's loop statuses over a day old, keeps its pump as sent only on
// its newest report with one, thins its reports from before the last hour to fewer the older they are, and deletes
// those over 30 days old, looking at those written since since, 0 for all of them. Then it deletes the loop settings
// no record refers to any more.
func truncateSite(ctx context.Context, site *SiteDB, since int64) error {
	now := time.Now()
	if err := site.Queries().ClearOldLoopStatuses(ctx, sitedb.ClearOldLoopStatusesParams{Since: since, Now: now.UnixMilli()}); err != nil {
		return fmt.Errorf("cannot truncate site %s: %w", site.Subdomain, err)
	}
	if err := site.Queries().ClearOldPumps(ctx, sitedb.ClearOldPumpsParams{Since: since}); err != nil {
		return fmt.Errorf("cannot truncate site %s: %w", site.Subdomain, err)
	}
	if err := site.Queries().ThinOldDeviceReports(ctx, ThinParams(since, now.UnixMilli())); err != nil {
		return fmt.Errorf("cannot thin the device reports of site %s: %w", site.Subdomain, err)
	}
	before := now.Add(-DeviceReportsKeptFor).UnixMilli()
	if _, err := site.Queries().DeleteOldDeviceReports(ctx, sitedb.DeleteOldDeviceReportsParams{Before: before}); err != nil {
		return fmt.Errorf("cannot delete the old device reports of site %s: %w", site.Subdomain, err)
	}
	// The loop settings that the reports just thinned or deleted were the last to refer to
	if _, err := site.Queries().DeleteUnreferencedLoopSettings(ctx); err != nil {
		return fmt.Errorf("cannot delete the unreferenced loop settings of site %s: %w", site.Subdomain, err)
	}
	return nil
}

////
// END OF DELETE USER METHODS
////

// //
// Get USER Methods
// //

func (db *ControlDB) GetNSitesDBs() int {
	return db.sites.count()
}

func (db *ControlDB) AllSites() []*SiteDB {
	return db.sites.all()
}

func (db *ControlDB) SubdomainUsed(ctx context.Context, subdomain string) bool {
	exists, err := db.Queries.SiteExists(ctx, controldb.SiteExistsParams{
		ID: subdomain,
	})
	if err != nil {
		return false
	}
	return exists
}

// UpdateSiteAndIdentityDateFlags records when the site and identity were last seen and written to, each at most every
// 10 minutes from its own last record, and the site DB's size with its last write. It writes nothing when none is due,
// as each write waits on the control DB's one write lock.
func (db *ControlDB) UpdateSiteAndIdentityDateFlags(ctx context.Context, siteModel *controldb.Site, siteDB *SiteDB, identityModel *controldb.SiteIdentity, updateLastSeen bool, updateLastWrite bool) error {
	now := time.Now().UnixMilli()
	var siteErr, identityErr error
	if lastSeenAt, lastWriteAt := dueDateFlags(now, siteModel.LastSeenAt, siteModel.LastWriteAt, updateLastSeen, updateLastWrite); lastSeenAt != nil || lastWriteAt != nil {
		var estimatedDBSizeKB *int64
		if lastWriteAt != nil {
			estimatedDBSizeKB = siteDB.EstimatedSizeKB()
		}
		siteErr = db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{
			LastSeenAt:        lastSeenAt,
			LastWriteAt:       lastWriteAt,
			EstimatedDbSizeKb: estimatedDBSizeKB,
			ID:                siteModel.ID,
		})
	}
	if lastSeenAt, lastWriteAt := dueDateFlags(now, identityModel.LastSeenAt, identityModel.LastWriteAt, updateLastSeen, updateLastWrite); lastSeenAt != nil || lastWriteAt != nil {
		identityErr = db.Queries.UpdateSiteIdentityDateFlags(ctx, controldb.UpdateSiteIdentityDateFlagsParams{
			LastSeenAt:  lastSeenAt,
			LastWriteAt: lastWriteAt,
			ID:          identityModel.ID,
		})
	}
	return errors.Join(siteErr, identityErr)
}

// dueDateFlags is now, for each of the last seen and last write to record, when it was never recorded or was last
// recorded over 10 minutes ago, and nil otherwise
func dueDateFlags(now int64, lastSeenAt *int64, lastWriteAt *int64, updateLastSeen bool, updateLastWrite bool) (*int64, *int64) {
	due := func(update bool, last *int64) *int64 {
		if update && (last == nil || now > *last+(10*time.Minute).Milliseconds()) {
			return &now
		}
		return nil
	}
	return due(updateLastSeen, lastSeenAt), due(updateLastWrite, lastWriteAt)
}

// RecordSiteSizes measures each site's DB on disk and keeps the sizes that changed, in one transaction, so the sizes
// shown count what syncs write as well as what the site's own requests do. A DB in memory has no size to keep.
func (db *ControlDB) RecordSiteSizes(ctx context.Context, sites []*SiteDB) error {
	tx, err := db.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot record site sizes: %w", err)
	}
	defer tx.Rollback()
	q := db.Queries.WithTx(tx)
	for _, site := range sites {
		if sizeKB := site.EstimatedSizeKB(); sizeKB != nil {
			if err := q.UpdateSiteDBSize(ctx, controldb.UpdateSiteDBSizeParams{EstimatedDbSizeKb: sizeKB, ID: site.Subdomain}); err != nil {
				return fmt.Errorf("cannot record site %s size: %w", site.Subdomain, err)
			}
		}
	}
	return tx.Commit()
}

// RecordAbuseScores scores each site for abuse, from its DB's size now and its connectors' failures, and keeps the
// scores in one transaction, so the control pages show each site's level without measuring them
func (db *ControlDB) RecordAbuseScores(ctx context.Context, sites []*SiteDB) error {
	limits, err := db.Queries.GetSiteDBSizeLimits(ctx)
	failures, failuresErr := db.Queries.GetSiteConnectorFailures(ctx)
	if err := errors.Join(err, failuresErr); err != nil {
		return fmt.Errorf("cannot score sites for abuse: %w", err)
	}
	signals := map[string]*controldb.AbuseSignals{}
	for _, site := range sites {
		signals[site.Subdomain] = &controldb.AbuseSignals{SizeKB: site.EstimatedSizeKB()}
	}
	for _, limit := range limits {
		if site := signals[limit.ID]; site != nil {
			site.LimitKB = limit.MaxDbSizeKbLimit
		}
	}
	for _, failure := range failures {
		if site := signals[failure.SiteID]; site != nil {
			site.ConnectorFailures = failure.ConsecutiveFailedSyncs
		}
	}
	tx, err := db.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot record abuse scores: %w", err)
	}
	defer tx.Rollback()
	q := db.Queries.WithTx(tx)
	for id, site := range signals {
		if err := q.UpdateSiteAbuseScore(ctx, controldb.UpdateSiteAbuseScoreParams{AbuseScore: site.Score(), ID: id}); err != nil {
			return fmt.Errorf("cannot record site %s abuse score: %w", id, err)
		}
	}
	return tx.Commit()
}

// RecordSiteSize measures one site's DB on disk now and keeps its size, as RecordSiteSizes does every site's, so a page
// showing it shows now's size rather than the last resource check's. A site whose DB is not open here, or is in
// memory, keeps the size it has.
func (db *ControlDB) RecordSiteSize(ctx context.Context, subdomain string) error {
	site := db.sites.get(subdomain)
	if site == nil {
		return nil
	}
	return db.RecordSiteSizes(ctx, []*SiteDB{site})
}

// RecordSiteTableSizes measures each table of each site's DB from SQLite's dbstat and keeps them in its site, for its
// pages to show without reading the whole DB. A site is measured again only when its DB's size changed since, as dbstat
// reads every page of the DB; one in memory, whose size is not known, is measured each time.
func (db *ControlDB) RecordSiteTableSizes(ctx context.Context, sites []*SiteDB) error {
	var errs []error
	for _, site := range sites {
		sizeKB := site.EstimatedSizeKB()
		if sizeKB != nil && site.tablesMeasuredKB.Load() == *sizeKB {
			continue
		}
		if err := db.recordSiteTableSizes(ctx, site); err != nil {
			errs = append(errs, err)
			continue
		}
		if sizeKB != nil {
			site.tablesMeasuredKB.Store(*sizeKB)
		}
	}
	return errors.Join(errs...)
}

func (db *ControlDB) recordSiteTableSizes(ctx context.Context, site *SiteDB) error {
	sizes, err := site.TableSizes(ctx)
	if err != nil {
		return fmt.Errorf("cannot measure site %s tables: %w", site.Subdomain, err)
	}
	tables := types.TableSizes{}
	for _, size := range sizes {
		tables[size.Name] = types.TableSize{Bytes: size.Bytes, Records: size.Records}
	}
	return db.Queries.UpdateSiteTableSizes(ctx, controldb.UpdateSiteTableSizesParams{TableSizesJson: tables, ID: site.Subdomain})
}

// DBTable is a table of one of the server's own DBs: the bytes it and its indexes take, and its records
type DBTable struct {
	Name    string
	Bytes   int64
	Records int64
}

// ServerDB is one of the server's own DBs: its name, its size on disk in KB now, or nil for one in memory, and its
// tables, largest first, as last measured, or none before they have been
type ServerDB struct {
	Name   string
	SizeKB *int64
	Tables []DBTable
}

// measuredTables are a DB's tables, largest first, and its size on disk when they were measured
type measuredTables struct {
	sizeKB *int64
	tables []DBTable
}

// serverDB is one of the server's own DBs by name, with its file and how to measure its tables
type serverDB struct {
	name     string
	path     string
	inMemory bool
	measure  func(context.Context) ([]DBTable, error)
}

// serverDBs are the server's own DBs, the control DB first, then its logs and events, its jobs, and what it keeps
// only while it runs
func (db *ControlDB) serverDBs() []serverDB {
	return []serverDB{
		{"control", db.Path, db.InMemory, func(ctx context.Context) ([]DBTable, error) { return dbTables(db.TableSizes(ctx)) }},
		{"logs", db.Logs.Path, db.Logs.InMemory, func(ctx context.Context) ([]DBTable, error) { return dbTables(db.Logs.TableSizes(ctx)) }},
		{"jobs", db.Jobs.Path, db.Jobs.InMemory, func(ctx context.Context) ([]DBTable, error) { return dbTables(db.Jobs.TableSizes(ctx)) }},
		{"live", db.Live.Path, db.Live.InMemory, func(ctx context.Context) ([]DBTable, error) { return dbTables(db.Live.TableSizes(ctx)) }},
	}
}

// dbTables are a DB's measured tables as DBTables, as each DB's package has its own TableSize with the same fields
func dbTables[T ~struct {
	Name    string
	Bytes   int64
	Records int64
}](sizes []T, err error) ([]DBTable, error) {
	tables := make([]DBTable, len(sizes))
	for i, size := range sizes {
		tables[i] = DBTable(size)
	}
	return tables, err
}

// RecordServerTableSizes measures each table of each of the server's own DBs from SQLite's dbstat, as
// RecordSiteTableSizes does a site's, and only when the DB's size changed since, or each time for one in memory, as
// dbstat reads every page of the DB
func (db *ControlDB) RecordServerTableSizes(ctx context.Context) error {
	var errs []error
	for _, server := range db.serverDBs() {
		sizeKB := diskSizeKB(server.path, server.inMemory)
		if last, ok := db.measured.Load(server.name); ok && sizeKB != nil && last.(*measuredTables).sizeKB != nil && *last.(*measuredTables).sizeKB == *sizeKB {
			continue
		}
		tables, err := server.measure(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("cannot measure the %s DB's tables: %w", server.name, err))
			continue
		}
		db.measured.Store(server.name, &measuredTables{sizeKB: sizeKB, tables: tables})
	}
	return errors.Join(errs...)
}

// ServerDBs are the server's own DBs, each with its size now and its tables as last measured
func (db *ControlDB) ServerDBs() []ServerDB {
	dbs := []ServerDB{}
	for _, server := range db.serverDBs() {
		size := ServerDB{Name: server.name, SizeKB: diskSizeKB(server.path, server.inMemory)}
		if last, ok := db.measured.Load(server.name); ok {
			size.Tables = last.(*measuredTables).tables
		}
		dbs = append(dbs, size)
	}
	return dbs
}

// SiteTableSizes are the bytes each table of the site's DB takes, its indexes with it, and its records, largest first,
// as last measured, or none before the site's tables have been
func (db *ControlDB) SiteTableSizes(ctx context.Context, subdomain string) ([]sitedb.TableSize, error) {
	site, err := db.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: subdomain})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sizes := []sitedb.TableSize{}
	for name, size := range site.TableSizesJson {
		sizes = append(sizes, sitedb.TableSize{Name: name, Bytes: size.Bytes, Records: size.Records})
	}
	slices.SortFunc(sizes, func(a, b sitedb.TableSize) int {
		return cmp.Or(cmp.Compare(b.Bytes, a.Bytes), strings.Compare(a.Name, b.Name))
	})
	return sizes, nil
}

// ErrSiteNotFound is a site there is not, or whose DB is not open, as against one that cannot be looked up now
var ErrSiteNotFound = errors.New("site not found")

func (db *ControlDB) GetSite(ctx context.Context, subdomain string) (*SiteDB, *controldb.Site, error) {
	// Check if site exists in the db
	model, err := db.Queries.GetSite(ctx, controldb.GetSiteParams{
		ID: subdomain,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrSiteNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get site from db: %w", err)
	}
	if model.ID == "" {
		return nil, nil, ErrSiteNotFound
	}

	siteDB := db.sites.get(subdomain)
	if siteDB == nil && db.unavailable.has(subdomain) {
		return nil, nil, ErrSiteUnavailable
	}
	if siteDB == nil {
		return nil, nil, fmt.Errorf("Cannot get SiteDB from memory: %w in memory", ErrSiteNotFound)
	}
	return siteDB, model, nil
}

////
// End of Get SITE Methods
////

// Close sends the control DB's last writes to its replica, then closes the DBs
func (db *ControlDB) Close() error {
	if err := db.replica.Close(context.Background()); err != nil {
		slog.Error("Failed to send the control DB's last writes to its replica", "error", err)
	}
	for _, site := range db.sites.all() {
		err := site.Close()
		if err != nil {
			slog.Error("Failed to close site db", "subdomain", site.Subdomain, "error", err)
		}
	}

	err := db.Jobs.Close()
	if err != nil {
		slog.Error("Failed to close jobs db", "error", err)
	}

	if err := db.Logs.Close(); err != nil {
		slog.Error("Failed to close logs db", "error", err)
	}

	if err := db.Live.Close(); err != nil {
		slog.Error("Failed to close live db", "error", err)
	}

	return db.SQLiteDB.Close()
}

// ////
// Identities
// ////

// createSiteRows writes the site and its default identities in one transaction, so a failure part way, such as the
// request being cancelled, leaves none of them
func (db *ControlDB) createSiteRows(ctx context.Context, subdomain string, backupKeyHash string) (*IdentityValues, *IdentityValues, *IdentityValues, error) {
	tx, err := db.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create site: %w", err)
	}
	defer tx.Rollback()
	q := db.Queries.WithTx(tx)

	if _, err := q.CreateSite(ctx, controldb.CreateSiteParams{ID: subdomain, BackupKeyHash: backupKeyHash}); err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create site: %w", err)
	}
	adminIdentity, pwdIdentity, readOnlyIdentity, err := createSiteIdentities(db, ctx, q, subdomain)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create admin identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create site: %w", err)
	}
	return adminIdentity, pwdIdentity, readOnlyIdentity, nil
}

// createSiteIdentities makes a new site's first identities, in its transaction; a test makes it fail partway
var createSiteIdentities = (*ControlDB).createSiteDefaultIdentities

func (db *ControlDB) createSiteDefaultIdentities(ctx context.Context, q *controldb.Queries, subdomain string) (*IdentityValues, *IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	// Create Default Read-Only User
	readOnlyIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Read-Only User:	 %w", err)
	}
	if readOnlyIdentity.NSToken == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Read-Only User: NSToken is nil")
	}

	return adminIdentity, pwdIdentity, readOnlyIdentity, nil
}

type IdentityValues struct {
	Password   string
	TotpSecret *string
	NSToken    *string
}

// CreateSiteIdentity writes with q, so a caller can pass the queries of its transaction
// CreateSiteIdentity makes an identity on the site, with the app it is for when it is a device added for one
func (db *ControlDB) CreateSiteIdentity(ctx context.Context, q *controldb.Queries, siteID string, name string, role consts.IdentityRole, client *string) (*IdentityValues, error) {

	security := db.ServerConfig.Security
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.IdentityMustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(siteID)
		if err != nil {
			return nil, fmt.Errorf("cannot generate user TOTP: %w", err)
		}

		totpSecretPtr = &totpSecret
		encTOTPSecret = encTOTPSecretRet
	}

	nsToken, digest, err := db.nightscoutToken(siteID, name, role, password)
	if err != nil {
		return nil, err
	}

	err = q.CreateSiteIdentity(ctx, controldb.CreateSiteIdentityParams{
		SiteID:        siteID,
		Name:          name,
		PasswordHash:  hashedPassword,
		TotpSecretEnc: encTOTPSecret,
		Role:          role,
		Client:        client,

		NightscoutTokenDigest: digest,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create user identity: %w", err)
	}

	return &IdentityValues{
		Password:   password,
		TotpSecret: totpSecretPtr,
		NSToken:    nsToken,
	}, nil
}

// ResetSiteIdentityPassword gives one of a site's identities a new random password, and a device the Nightscout token
// made from it, recorded as changed by the control or site identity named by. The old password stops working.
func (db *ControlDB) ResetSiteIdentityPassword(ctx context.Context, identity *controldb.SiteIdentity, by string) (*IdentityValues, error) {
	password, hashedPassword, err := db.ServerConfig.Security.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}
	nsToken, digest, err := db.nightscoutToken(identity.SiteID, identity.Name, identity.Role, password)
	if err == nil {
		err = db.deletePasskeys(ctx, identity)
	}
	if err != nil {
		return nil, err
	}
	if _, err := db.Queries.UpdateSiteIdentityPassword(ctx, controldb.UpdateSiteIdentityPasswordParams{PasswordHash: hashedPassword,
		NightscoutTokenDigest: digest, UpdatedByIdentityName: &by, SiteID: identity.SiteID, ID: identity.ID}); err != nil {
		return nil, err
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, identity.SiteID, identity.Name, "password", by)
	return &IdentityValues{Password: password, NSToken: nsToken}, nil
}

// nightscoutToken is the Nightscout token of an identity that can have one, made from its password, and the
// digest kept to find it by the token's SHA1, see security.NightscoutTokenDigest
func (db *ControlDB) nightscoutToken(siteID, name string, role consts.IdentityRole, password string) (*string, *string, error) {
	if !can.IdentityHaveNightscoutToken(role) {
		return nil, nil, nil
	}
	token, err := db.ServerConfig.Security.CreateNightscoutTokenForIdentity(siteID, name, string(role), password)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create nightscout token for identity: %w", err)
	}
	digest := security.NightscoutTokenDigest(security.NightscoutTokenSHA1(token))
	return &token, &digest, nil
}

// ResetSiteIdentityTOTP gives one of a site's identities a new TOTP secret, for its authenticator app, recorded as
// changed by the control or site identity named by. Codes from the old secret stop working.
func (db *ControlDB) ResetSiteIdentityTOTP(ctx context.Context, identity *controldb.SiteIdentity, by string) (*IdentityValues, error) {
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate(identity.SiteID)
	if err != nil {
		return nil, fmt.Errorf("cannot generate user TOTP: %w", err)
	}
	if err := db.deletePasskeys(ctx, identity); err != nil {
		return nil, err
	}
	if _, err := db.Queries.UpdateSiteIdentityTOTP(ctx, controldb.UpdateSiteIdentityTOTPParams{TotpSecretEnc: encTOTPSecret,
		UpdatedByIdentityName: &by, SiteID: identity.SiteID, ID: identity.ID}); err != nil {
		return nil, err
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, identity.SiteID, identity.Name, "totp", by)
	return &IdentityValues{TotpSecret: &totpSecret}, nil
}

// deletePasskeys removes an identity's passkeys before its password or TOTP is reset, as for a lost or taken device,
// so it then signs in only with its new credentials; removed first, a reset that then fails leaves it to add them again
func (db *ControlDB) deletePasskeys(ctx context.Context, identity *controldb.SiteIdentity) error {
	return db.Queries.DeleteSiteIdentityPasskeys(ctx, controldb.DeleteSiteIdentityPasskeysParams{SiteID: identity.SiteID, IdentityID: identity.ID})
}

// CreateControlIdentity makes a control admin, with a new password and TOTP secret shown this once, recorded as made by
// the admin by, or by no one for the first
func (db *ControlDB) CreateControlIdentity(ctx context.Context, name string, role consts.ControlRole, by *controldb.ControlIdentity) (*IdentityValues, error) {
	password, hashedPassword, err := db.ServerConfig.Security.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("cannot generate control password: %w", err)
	}
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate("")
	if err != nil {
		return nil, fmt.Errorf("cannot generate control TOTP: %w", err)
	}
	params := controldb.CreateControlIdentityParams{Name: name, PasswordHash: hashedPassword, TotpSecretEnc: encTOTPSecret, Role: role}
	if by != nil {
		params.CreatedByIdentityID, params.CreatedByIdentityName = by.ID, by.Name
	}
	if err := db.Queries.CreateControlIdentity(ctx, params); err != nil {
		return nil, fmt.Errorf("cannot create control identity: %w", err)
	}
	return &IdentityValues{Password: password, TotpSecret: &totpSecret}, nil
}

// ResetControlIdentityPassword gives another control admin a new random password, recorded as changed by the admin by.
// Its passkeys are removed first, as for a lost device, so it signs in only with the new password.
func (db *ControlDB) ResetControlIdentityPassword(ctx context.Context, identity, by *controldb.ControlIdentity) (*IdentityValues, error) {
	password, hashedPassword, err := db.ServerConfig.Security.GeneratePassword()
	if err == nil {
		err = db.Queries.DeleteControlIdentityPasskeys(ctx, controldb.DeleteControlIdentityPasskeysParams{ControlIdentityID: identity.ID})
	}
	if err == nil {
		err = db.Queries.UpdateControlIdentityPassword(ctx, controldb.UpdateControlIdentityPasswordParams{PasswordHash: hashedPassword,
			UpdatedByIdentityID: &by.ID, UpdatedByIdentityName: &by.Name, ID: identity.ID})
	}
	if err != nil {
		return nil, fmt.Errorf("cannot reset the control admin's password: %w", err)
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, "", identity.Name, "password", by.Name)
	return &IdentityValues{Password: password}, nil
}

// ResetControlIdentityTOTP gives another control admin a new TOTP secret, recorded as changed by the admin by. Its
// passkeys are removed first, as for a lost device, so it signs in only with the new secret.
func (db *ControlDB) ResetControlIdentityTOTP(ctx context.Context, identity, by *controldb.ControlIdentity) (*IdentityValues, error) {
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate("")
	if err == nil {
		err = db.Queries.DeleteControlIdentityPasskeys(ctx, controldb.DeleteControlIdentityPasskeysParams{ControlIdentityID: identity.ID})
	}
	if err == nil {
		err = db.Queries.UpdateControlIdentityTOTP(ctx, controldb.UpdateControlIdentityTOTPParams{TotpSecretEnc: encTOTPSecret,
			UpdatedByIdentityID: &by.ID, UpdatedByIdentityName: &by.Name, ID: identity.ID})
	}
	if err != nil {
		return nil, fmt.Errorf("cannot reset the control admin's TOTP: %w", err)
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, "", identity.Name, "totp", by.Name)
	return &IdentityValues{TotpSecret: &totpSecret}, nil
}

// ErrLastControlAdmin refuses to delete control's last admin, or make it read only, as then none could manage control
var ErrLastControlAdmin = errors.New("the last admin cannot be deleted or made read only")

// SetControlIdentityRole gives a control admin a role, recorded as changed by the admin by, unless it is the last admin
func (db *ControlDB) SetControlIdentityRole(ctx context.Context, identity *controldb.ControlIdentity, role consts.ControlRole, by *controldb.ControlIdentity) error {
	return lastAdminKept(db.Queries.UpdateControlIdentityRole(ctx, controldb.UpdateControlIdentityRoleParams{Role: role,
		UpdatedByIdentityID: &by.ID, UpdatedByIdentityName: &by.Name, ID: identity.ID}))
}

// DeleteControlIdentity deletes a control admin and its passkeys, unless it is the last admin
func (db *ControlDB) DeleteControlIdentity(ctx context.Context, identity *controldb.ControlIdentity) error {
	err := lastAdminKept(db.Queries.DeleteControlIdentity(ctx, controldb.DeleteControlIdentityParams{ID: identity.ID}))
	if err == nil {
		err = db.Queries.DeleteControlIdentityPasskeys(ctx, controldb.DeleteControlIdentityPasskeysParams{ControlIdentityID: identity.ID})
	}
	return err
}

// lastAdminKept is a write to a control admin that changed it, or else ErrLastControlAdmin, as its query keeps the last
// admin as it is
func lastAdminKept(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err == nil && changed == 0 {
		err = ErrLastControlAdmin
	}
	return err
}

////
// Find Types of Sites (e.g. unused)
////

////
// Utils
////

// GetSiteDBsFromModels is the open DB of each site, leaving out, and logging, any site not open, so one site without its
// DB does not stop the rest
func (db *ControlDB) GetSiteDBsFromModels(models []*controldb.Site) []*SiteDB {
	sites := []*SiteDB{}
	for _, model := range models {
		site := db.sites.get(model.ID)
		// An unavailable site is left out quietly, as its restore logs why
		if site == nil && !db.unavailable.has(model.ID) {
			slog.Error("Site is not open, so it is left out", "site_id", model.ID)
		}
		if site == nil {
			continue
		}
		sites = append(sites, site)
	}

	return sites
}
