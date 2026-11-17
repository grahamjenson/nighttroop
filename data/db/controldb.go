package db

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/types"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
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
	// missingSites are the live sites whose DB files were missing at start, for Replicate to restore
	missingSites []string
	// replicas replicate the control DB and every open site's DB once Replicate starts them. Nil, as when S3 is off
	// or in tests, replicates nothing.
	replicas *replica.Replicas
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
		// A site without its DB is restored from its replica by Replicate, which stops the start if it cannot be
		if !utils.FileExists(sitedb.GetPath(model.ID, db.ServerConfig)) {
			slog.ErrorContext(ctx, "Site DB is missing, so the site is not opened", "site_id", model.ID)
			db.missingSites = append(db.missingSites, model.ID)
			continue
		}
		g.Go(func() error {
			siteDB, err := NewSiteDB(ctx, model.ID, db.ServerConfig)
			if err != nil {
				return fmt.Errorf("SiteDB failed to be added to the ControlDB list, %w", err)
			}
			db.sites.add(siteDB)
			return nil
		})
	}
	return g.Wait()
}

// ErrRestoredSiteCannotOpen is a site restored from its replica whose DB cannot be opened, as when it cannot be
// migrated, which stops the start as a site that is there and cannot be opened does
var ErrRestoredSiteCannotOpen = errors.New("a restored site's DB cannot be opened")

// ErrSitesMissing is live sites whose DB files are missing and were not restored from their replicas, which stops the
// start, as a site that cannot be opened does: a server without one of its sites serves nothing, so no data is written
// that a rollback would lose, and the sites' data is never served half there
var ErrSitesMissing = errors.New("live sites' DBs are missing and were not restored")

// restoreParallelism is how many missing sites are restored at once, as a new server restores every site
const restoreParallelism = 16

// Replicate restores each live site whose DB file was missing at start from its replica, and opens it, then
// replicates the control DB and every open site's DB with replicas. Each missing site is recorded as a site_db_missing
// event, so the control admin sees a site that could not be brought back. One that could not is ErrSitesMissing,
// naming the sites, and a restored site that cannot be opened is ErrRestoredSiteCannotOpen: either stops the start, so
// a release that cannot open every live site is rolled back rather than kept.
func (db *ControlDB) Replicate(ctx context.Context, replicas *replica.Replicas) error {
	var mu sync.Mutex
	var cannotOpen error
	missing := []string{}
	g := errgroup.Group{}
	g.SetLimit(restoreParallelism)
	for _, siteID := range db.missingSites {
		g.Go(func() error {
			reason, err := db.restoreSite(ctx, replicas, siteID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.ErrorContext(ctx, "Cannot restore a missing site DB from its replica", "site_id", siteID, "error", err)
				if errors.Is(err, ErrRestoredSiteCannotOpen) {
					cannotOpen = errors.Join(cannotOpen, err)
				}
			}
			if reason != "" {
				missing = append(missing, siteID)
			}
			db.Logs.SiteDBMissingEvent(aslog.WithSite(ctx, siteID), siteID, reason == "", reason)
			return nil
		})
	}
	_ = g.Wait()
	db.missingSites = nil
	if len(missing) > 0 {
		slices.Sort(missing)
		return errors.Join(cannotOpen, fmt.Errorf("%w: %s", ErrSitesMissing, strings.Join(missing, ", ")))
	}

	db.replicas = replicas
	sites := map[string]string{}
	for _, site := range db.sites.all() {
		sites[site.Path] = replica.SiteName(site.Subdomain)
		db.checkpointWithReplica(site)
	}
	return replicas.Start(ctx, db.Path, sites)
}

// restoreSite restores the site's DB from its replica and opens it, or is why it could not
func (db *ControlDB) restoreSite(ctx context.Context, replicas *replica.Replicas, siteID string) (string, error) {
	if replicas == nil {
		return "s3_not_working", nil
	}
	if found, err := replicas.Restore(ctx, replica.SiteName(siteID), sitedb.GetPath(siteID, db.ServerConfig)); err != nil {
		return "restore_failed", err
	} else if !found {
		return "not_in_backup", nil
	}
	siteDB, err := NewSiteDB(ctx, siteID, db.ServerConfig)
	if err != nil {
		return "restore_failed", fmt.Errorf("%w: %s: %w", ErrRestoredSiteCannotOpen, siteID, err)
	}
	db.sites.add(siteDB)
	slog.InfoContext(ctx, "Missing site DB restored from its replica", "site_id", siteID)
	return "", nil
}

// checkpointWithReplica has the site's Vacuum checkpoint with its replica, once it is replicated
func (db *ControlDB) checkpointWithReplica(site *SiteDB) {
	if db.replicas != nil {
		site.Checkpoint = func(ctx context.Context) error { return db.replicas.Checkpoint(ctx, site.Path) }
	}
}

// replicateNew replicates a site just made live, sending it and the control DB's row of it at once, so a restore has
// both. One not sent is logged, as Litestream sends it again.
func (db *ControlDB) replicateNew(ctx context.Context, site *SiteDB) {
	db.checkpointWithReplica(site)
	if err := db.replicas.Add(ctx, site.Path, replica.SiteName(site.Subdomain)); err != nil {
		slog.ErrorContext(ctx, "Cannot replicate the site", "site_id", site.Subdomain, "error", err)
	}
	db.syncControl(ctx)
}

// syncControl sends the control DB's writes now, such as a site's row once it is made, archived or deleted
func (db *ControlDB) syncControl(ctx context.Context) {
	if err := db.replicas.Sync(ctx, db.Path); err != nil {
		slog.ErrorContext(ctx, "Cannot send the control DB to its replica", "error", err)
	}
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
	result, err := db.createSite(ctx, "")
	if err == nil {
		db.replicateNew(ctx, result.SiteDB)
	}
	return result, err
}

// CreateSiteWithID creates a site at a site ID picked for it, such as a vanity one for a special user. It is 8
// lowercase letters like every site's, and one already taken is refused.
func (db *ControlDB) CreateSiteWithID(ctx context.Context, subdomain string) (*SiteCreateResult, error) {
	if !utils.SyntacticallyValidateSiteID(subdomain) {
		return nil, fmt.Errorf("site ID %q is not %d lowercase letters", subdomain, utils.SITE_ID_LENGTH)
	}
	result, err := db.createSite(ctx, subdomain)
	if err == nil {
		db.replicateNew(ctx, result.SiteDB)
	}
	return result, err
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
		_ = utils.DeleteFile(sitedb.GetPath(subdomain, db.ServerConfig))
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

// DeleteSites marks the sites deleted, then deletes each one's replica, and closes and removes its DB files. A site
// missing from the map stops the rest being closed, and a replica or file that cannot be removed is only logged, so a
// failure part way leaves the sites marked deleted with some of their data left behind.
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

	db.syncControl(ctx)
	// Close the DB and remove from the map
	for _, subdomain := range deletedSubdomains {
		site := db.sites.remove(subdomain)
		if site == nil {
			return fmt.Errorf("site not found")
		}
		if err := db.replicas.Remove(ctx, site.Path, replica.SiteName(subdomain), true); err != nil {
			slog.ErrorContext(ctx, "Failed to delete the site's replica", "subdomain", subdomain, "error", err)
		}
		site.Close()

		err := utils.DeleteFile(site.Path)
		if err != nil {
			slog.Error("Failed to delete site db files", "subdomain", subdomain, "error", err)
		} else {
			slog.Info("Deleted site db files", "subdomain", subdomain)
		}
	}

	return nil
}

// ErrNotReplicated is an archive or hydrate without S3, as a site's replica is its archive
var ErrNotReplicated = errors.New("sites are not replicated, as S3 is off")

// ArchiveSite sends a site's last writes to its replica, which is kept as its archive, then marks the site archived,
// stops replicating it and deletes its local files, so it is no longer served but can be restored from its replica.
// Files whose last writes could not be sent are kept.
func (db *ControlDB) ArchiveSite(ctx context.Context, subdomain string) error {
	site := db.sites.get(subdomain)
	if site == nil {
		return fmt.Errorf("site %s is not open", subdomain)
	}
	if db.replicas == nil {
		return ErrNotReplicated
	}
	if err := db.replicas.Sync(ctx, site.Path); err != nil {
		return fmt.Errorf("cannot send site %s to its replica: %w", subdomain, err)
	}
	bytes, _ := json.Marshal([]string{subdomain})
	if _, err := db.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: string(bytes)}); err != nil {
		return fmt.Errorf("cannot archive site %s: %w", subdomain, err)
	}
	db.syncControl(ctx)

	db.sites.remove(subdomain)
	// Its writes since the sync, as by a request that had it before it was removed, are sent as it stops
	err := db.replicas.Remove(ctx, site.Path, replica.SiteName(subdomain), false)
	site.Close()
	if err == nil {
		err = utils.DeleteFile(site.Path)
	}
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete archived site db files", "subdomain", subdomain, "error", err)
	}
	return nil
}

// HydrateSite brings an archived site back: it restores the site's DB from its replica, opens it, makes the site live
// again and replicates it again
func (db *ControlDB) HydrateSite(ctx context.Context, subdomain string) error {
	db.createSiteLock.Lock()
	defer db.createSiteLock.Unlock()
	if db.sites.get(subdomain) != nil {
		return fmt.Errorf("site %s is already open", subdomain)
	}
	if db.replicas == nil {
		return ErrNotReplicated
	}
	path := sitedb.GetPath(subdomain, db.ServerConfig)
	if found, err := db.replicas.Restore(ctx, replica.SiteName(subdomain), path); err != nil {
		return fmt.Errorf("cannot restore the archive of site %s: %w", subdomain, err)
	} else if !found {
		return fmt.Errorf("site %s has no archive", subdomain)
	}
	siteDB, err := NewSiteDB(ctx, subdomain, db.ServerConfig)
	if err != nil {
		// The archive is still in S3, so the copy restored from it goes
		_ = utils.DeleteFile(path)
		return err
	}
	result, err := db.Queries.HydrateSite(ctx, controldb.HydrateSiteParams{ID: subdomain})
	if err == nil {
		if rows, _ := result.RowsAffected(); rows == 0 {
			err = fmt.Errorf("site %s is not archived", subdomain)
		}
	}
	if err != nil {
		siteDB.Close()
		_ = utils.DeleteFile(path)
		return err
	}

	db.sites.add(siteDB)
	db.replicateNew(ctx, siteDB)
	return nil
}

// TruncateSite truncates a site as each backup does, but looking at all its device reports, so it clears
// those the backup no longer looks at, such as the newest of a device that then went a day without reporting. It deletes
// the loop settings no record refers to any more, then shrinks its DB, without its free pages, and records its new
// size, which it returns.
func (db *ControlDB) TruncateSite(ctx context.Context, subdomain string) (*int64, error) {
	site := db.sites.get(subdomain)
	if site == nil {
		return nil, fmt.Errorf("site %s is not open", subdomain)
	}
	if err := truncateSite(ctx, site, 0); err != nil {
		return nil, err
	}
	if _, err := site.Queries().DeleteUnreferencedLoopSettings(ctx); err != nil {
		return nil, fmt.Errorf("cannot delete the unreferenced loop settings of site %s: %w", subdomain, err)
	}
	if err := site.Vacuum(ctx); err != nil {
		return nil, fmt.Errorf("cannot vacuum site %s: %w", subdomain, err)
	}
	sizeKB := site.EstimatedSizeKB()
	if err := db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: sizeKB, ID: subdomain}); err != nil {
		return sizeKB, err
	}
	// Its tables are measured again now, as truncating changed them all, for the page that shows its new size
	return sizeKB, db.RecordSiteTableSizes(ctx, []*SiteDB{site})
}

// TruncateSites truncates sites, as each batch's backup does before copying them, without shrinking their DBs, as
// SQLite uses the space again for what comes next. It truncates harder a site far over its DB size limit. It goes on
// past a site it cannot truncate, returning each one's error.
func (db *ControlDB) TruncateSites(ctx context.Context, sites []*SiteDB) error {
	limits, err := db.Queries.GetSiteDBSizeLimits(ctx)
	errs := []error{err}
	limitsKB := map[string]int64{}
	for _, limit := range limits {
		limitsKB[limit.ID] = *limit.MaxDbSizeKbLimit
	}
	for _, site := range sites {
		siteCtx := aslog.WithSite(ctx, site.Subdomain)
		err := truncateSite(siteCtx, site, time.Now().Add(-truncateLooksBack).UnixMilli())
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

// hardTruncateSite truncates the site harder while its DB is over hardTruncateSizeFactor times its limit, limitKB.
// First it clears the JSON that matters only while recent, and deletes device reports over 7 days old and the loop
// settings no record refers to. If the site is still over, it deletes its readings and treatments over 90 days old,
// but the newest of each type. After each step that changed records it shrinks the DB, so its size shows them. If any
// changed, it records the new size and logs a "Site truncated" event.
func (db *ControlDB) hardTruncateSite(ctx context.Context, site *SiteDB, limitKB int64) error {
	over := func(sizeKB *int64) bool {
		return sizeKB != nil && float64(*sizeKB) > hardTruncateSizeFactor*float64(limitKB)
	}
	beforeKB := site.EstimatedSizeKB()
	if !over(beforeKB) {
		return nil
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
		err = site.Vacuum(ctx)
	}
	if sizeKB := site.EstimatedSizeKB(); err == nil && over(sizeKB) {
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
			err = site.Vacuum(ctx)
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

// deviceReportsKeptFor is how long a device report is kept at all
const deviceReportsKeptFor = 30 * 24 * time.Hour

// truncateSite clears the predictions from each device's loop statuses over a day old, keeps its pump as sent only on
// its newest report with one, thins its reports from before the last hour to fewer the older they are, and deletes
// those over 30 days old, looking at those written since since, 0 for all of them
func truncateSite(ctx context.Context, site *SiteDB, since int64) error {
	now := time.Now()
	if err := site.Queries().ClearOldLoopStatuses(ctx, sitedb.ClearOldLoopStatusesParams{Since: since, Now: now.UnixMilli()}); err != nil {
		return fmt.Errorf("cannot truncate site %s: %w", site.Subdomain, err)
	}
	if err := site.Queries().ClearOldPumps(ctx, sitedb.ClearOldPumpsParams{Since: since}); err != nil {
		return fmt.Errorf("cannot truncate site %s: %w", site.Subdomain, err)
	}
	if err := site.Queries().ThinOldDeviceReports(ctx, sitedb.ThinOldDeviceReportsParams{Since: since, Now: now.UnixMilli()}); err != nil {
		return fmt.Errorf("cannot thin the device reports of site %s: %w", site.Subdomain, err)
	}
	before := now.Add(-deviceReportsKeptFor).UnixMilli()
	if _, err := site.Queries().DeleteOldDeviceReports(ctx, sitedb.DeleteOldDeviceReportsParams{Before: before}); err != nil {
		return fmt.Errorf("cannot delete the old device reports of site %s: %w", site.Subdomain, err)
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
	if siteDB == nil {
		return nil, nil, fmt.Errorf("Cannot get SiteDB from memory: %w in memory", ErrSiteNotFound)
	}
	return siteDB, model, nil
}

////
// End of Get SITE Methods
////

// SiteReplicatedAt is when the open site's replica last had all its writes, or nil if not since start
func (db *ControlDB) SiteReplicatedAt(siteID string) *int64 {
	site := db.sites.get(siteID)
	if site == nil {
		return nil
	}
	return db.replicas.Replicated(site.Path)
}

// Close sends each replicated DB's last writes, then closes the DBs
func (db *ControlDB) Close() error {
	if err := db.replicas.Close(context.Background()); err != nil {
		slog.Error("Failed to send the DBs' last writes to their replicas", "error", err)
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
	hostname := db.ServerConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.IdentityMustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(siteID, fmt.Sprintf("%s@%s.%s", name, siteID, hostname))
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
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate(identity.SiteID, fmt.Sprintf("%s@%s.%s", identity.Name, identity.SiteID, db.ServerConfig.HostName))
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
	totpSecret, encTOTPSecret, err := db.controlTOTP(name)
	if err != nil {
		return nil, err
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

// controlTOTP is a new TOTP secret for a control admin's authenticator app, and the secret encrypted to keep
func (db *ControlDB) controlTOTP(name string) (string, string, error) {
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate("", fmt.Sprintf("%s@%s.%s", name, consts.CONTROL_IDENTITY_SUBDOMAIN, db.ServerConfig.HostName))
	if err != nil {
		return "", "", fmt.Errorf("cannot generate control TOTP: %w", err)
	}
	return totpSecret, encTOTPSecret, nil
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
	totpSecret, encTOTPSecret, err := db.controlTOTP(identity.Name)
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
// Batch Operations
////

////
// Utils
////

// GetSiteDBsFromModels is the open DB of each site, leaving out, and logging, any site not open, so one site without its
// DB does not stop the rest of its batch
func (db *ControlDB) GetSiteDBsFromModels(models []*controldb.Site) []*SiteDB {
	sites := []*SiteDB{}
	for _, model := range models {
		site := db.sites.get(model.ID)
		if site == nil {
			slog.Error("Site is not open, so it is left out of its batch", "site_id", model.ID)
			continue
		}
		sites = append(sites, site)
	}

	return sites
}
