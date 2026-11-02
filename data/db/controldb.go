package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"

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
	// missingSites are the live sites whose DB files were missing at start, for RestoreMissingSites
	missingSites []string
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
		// A site without its DB, such as one restored before its batch's backup, is left out rather than stop the start
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

// RestoreMissingSites restores each live site whose DB file was missing at start from its batch's latest backup,
// downloaded once per batch with download, and opens it. Each is recorded as a site_db_missing event, so the control
// admin sees a site the backup could not bring back, which is still left out.
func (db *ControlDB) RestoreMissingSites(ctx context.Context, download s3utils.DownloadFunc) {
	byBatch := map[string][]string{}
	for _, siteID := range db.missingSites {
		byBatch[siteID[:1]] = append(byBatch[siteID[:1]], siteID)
	}
	db.missingSites = nil
	for batch, siteIDs := range byBatch {
		// Without S3 there is nothing to download, so the start does not wait on a download that cannot work
		var restored map[string]bool
		reason := "not_in_backup"
		if !db.ServerConfig.Status.S3Working {
			reason = "s3_not_working"
		} else if got, err := db.restoreFromBatch(ctx, batch, siteIDs, download); err != nil {
			slog.ErrorContext(ctx, "Cannot restore missing site DBs from their batch's backup", "batch", batch, "error", err)
			restored, reason = got, "restore_failed"
		} else {
			restored = got
		}
		for _, siteID := range siteIDs {
			if restored[siteID] {
				slog.InfoContext(ctx, "Missing site DB restored from its batch's backup", "site_id", siteID)
				db.Logs.SiteDBMissingEvent(aslog.WithSite(ctx, siteID), siteID, true, "")
			} else {
				db.Logs.SiteDBMissingEvent(aslog.WithSite(ctx, siteID), siteID, false, reason)
			}
		}
	}
}

// restoreFromBatch downloads batch's latest backup beside its folder, moves into place each of siteIDs' DBs it has,
// and opens them, and is the sites it restored
func (db *ControlDB) restoreFromBatch(ctx context.Context, batch string, siteIDs []string, download s3utils.DownloadFunc) (map[string]bool, error) {
	folder := sitedb.GetBatchFolder(batch, db.ServerConfig)
	if err := utils.MkDirs(folder); err != nil {
		return nil, err
	}
	// Named like a restore's staging folder, so one left by a start killed part way is cleaned up like those
	dir, err := os.MkdirTemp(filepath.Dir(folder), "."+batch+"-restore-")
	if err != nil {
		return nil, fmt.Errorf("cannot create staging dir: %w", err)
	}
	defer os.RemoveAll(dir)
	if _, found, err := download(ctx, sitedb.GetS3BatchPath(batch, db.ServerConfig), dir); err != nil || !found {
		return nil, err
	}
	restored := map[string]bool{}
	for _, siteID := range siteIDs {
		path := sitedb.GetPath(siteID, db.ServerConfig)
		// A rename on the same disk, so the site's DB is there whole or not at all
		if err := os.Rename(filepath.Join(dir, filepath.Base(path)), path); err != nil {
			continue
		}
		siteDB, err := NewSiteDB(ctx, siteID, db.ServerConfig)
		if err != nil {
			return restored, fmt.Errorf("cannot open restored site %s: %w", siteID, err)
		}
		db.sites.add(siteDB)
		restored[siteID] = true
	}
	return restored, nil
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

// DeleteSites marks the sites deleted, then closes and removes each one's DB files. A site missing from the map stops
// the rest being closed, and a file that cannot be removed is only logged, so a failure part way leaves the sites
// marked deleted with some files left behind.
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

	// Close the DB and remove from the map
	for _, subdomain := range deletedSubdomains {
		site := db.sites.remove(subdomain)
		if site == nil {
			return fmt.Errorf("site not found")
		}
		site.Close()

		err := utils.DeleteFiles([]string{site.Path, site.BackupPath})
		if err != nil {
			slog.Error("Failed to delete site db files", "subdomain", subdomain, "error", err)
		} else {
			slog.Info("Deleted site db files", "subdomain", subdomain)
		}
	}

	return nil
}

// ArchiveSite backs up a site's DB and uploads it, with upload, to its archive path, then marks the site archived and
// deletes its local files, so it is no longer served or backed up but can be restored from the archive
func (db *ControlDB) ArchiveSite(ctx context.Context, subdomain string, upload s3utils.UploadFunc) error {
	site := db.sites.get(subdomain)
	if site == nil {
		return fmt.Errorf("site %s is not open", subdomain)
	}
	// Backed up to a folder of its own beside the site's backup file, on the same disk, as a batch backup can be
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
	if err := backup.SqliteBackup(ctx, site.SQLiteDB, path); err != nil {
		return fmt.Errorf("cannot back up site %s: %w", subdomain, err)
	}
	if err := upload(ctx, sitedb.GetS3ArchivePath(subdomain, db.ServerConfig), s3utils.SiteRetention, []string{path}); err != nil {
		return fmt.Errorf("cannot upload the archive of site %s: %w", subdomain, err)
	}
	bytes, _ := json.Marshal([]string{subdomain})
	if _, err := db.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: string(bytes)}); err != nil {
		return fmt.Errorf("cannot archive site %s: %w", subdomain, err)
	}

	db.sites.remove(subdomain)
	site.Close()
	if err := utils.DeleteFiles([]string{site.Path, site.BackupPath}); err != nil {
		slog.ErrorContext(ctx, "Failed to delete archived site db files", "subdomain", subdomain, "error", err)
	}
	return nil
}

// HydrateSite brings an archived site back: it downloads, with download, the site's archive into its DB's folder, opens
// it and makes the site live again
func (db *ControlDB) HydrateSite(ctx context.Context, subdomain string, download s3utils.DownloadFunc) error {
	db.createSiteLock.Lock()
	defer db.createSiteLock.Unlock()
	if db.sites.get(subdomain) != nil {
		return fmt.Errorf("site %s is already open", subdomain)
	}
	path := sitedb.GetPath(subdomain, db.ServerConfig)
	if _, found, err := download(ctx, sitedb.GetS3ArchivePath(subdomain, db.ServerConfig), filepath.Dir(path)); err != nil {
		return fmt.Errorf("cannot download the archive of site %s: %w", subdomain, err)
	} else if !found || !utils.FileExists(path) {
		return fmt.Errorf("site %s has no archive", subdomain)
	}
	siteDB, err := NewSiteDB(ctx, subdomain, db.ServerConfig)
	if err != nil {
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
		_ = utils.DeleteFiles([]string{path})
		return err
	}

	db.sites.add(siteDB)
	return nil
}

// TruncateSite clears the raw JSON a site's DB keeps of what each row older than TruncateRecordsAfterDays was sent,
// keeping all its data, then shrinks the DB and records its new size, which it returns
func (db *ControlDB) TruncateSite(ctx context.Context, subdomain string) (*int64, error) {
	site := db.sites.get(subdomain)
	if site == nil {
		return nil, fmt.Errorf("site %s is not open", subdomain)
	}
	q, before := site.Queries(), time.Now().AddDate(0, 0, -db.ServerConfig.TruncateRecordsAfterDays).UnixMilli()
	for _, clear := range []func(context.Context, int64) error{
		clearBefore(q.ClearActivitiesRawJSON),
		clearBefore(q.ClearComponentEventsRawJSON),
		clearBefore(q.ClearDeviceReportsRawJSON),
		clearBefore(q.ClearEventGroupsRawJSON),
		clearBefore(q.ClearFoodIntakesRawJSON),
		clearBefore(q.ClearFoodPresetsRawJSON),
		clearBefore(q.ClearGlucoseReadingsRawJSON),
		clearBefore(q.ClearInsulinDosesRawJSON),
		clearBefore(q.ClearNotificationEventsRawJSON),
		clearBefore(q.ClearProfilePresetsRawJSON),
		clearBefore(q.ClearProfileSwitchesRawJSON),
		clearBefore(q.ClearTempBasalsRawJSON),
		clearBefore(q.ClearTempTargetsRawJSON),
	} {
		if err := clear(ctx, before); err != nil {
			return nil, fmt.Errorf("cannot clear raw json of site %s: %w", subdomain, err)
		}
	}
	if err := site.Vacuum(ctx); err != nil {
		return nil, fmt.Errorf("cannot vacuum site %s: %w", subdomain, err)
	}
	sizeKB := site.EstimatedSizeKB()
	return sizeKB, db.Queries.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{EstimatedDbSizeKb: sizeKB, ID: subdomain})
}

// clearBefore calls a query clearing raw JSON with its params, which sqlc makes a type of their own for each query
func clearBefore[P ~struct{ Before int64 }](clear func(context.Context, P) error) func(context.Context, int64) error {
	return func(ctx context.Context, before int64) error {
		return clear(ctx, P(struct{ Before int64 }{before}))
	}
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

func (db *ControlDB) GetSite(ctx context.Context, subdomain string) (*SiteDB, *controldb.Site, error) {
	// Check if site exists in the db
	model, err := db.Queries.GetSite(ctx, controldb.GetSiteParams{
		ID: subdomain,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get site from db: %w", err)
	}
	if model.ID == "" {
		return nil, nil, fmt.Errorf("site not found")
	}

	siteDB := db.sites.get(subdomain)
	if siteDB == nil {
		return nil, nil, fmt.Errorf("Cannot get SiteDB from memory: site not found in memory")
	}
	return siteDB, model, nil
}

////
// End of Get SITE Methods
////

func (db *ControlDB) Close() error {
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
	adminIdentity, pwdIdentity, readOnlyIdentity, err := db.createSiteDefaultIdentities(ctx, q, subdomain)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create admin identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, fmt.Errorf("cannot create site: %w", err)
	}
	return adminIdentity, pwdIdentity, readOnlyIdentity, nil
}

func (db *ControlDB) createSiteDefaultIdentities(ctx context.Context, q *controldb.Queries, subdomain string) (*IdentityValues, *IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	// Create Default Read-Only User
	readOnlyIdentity, err := db.CreateSiteIdentity(ctx, q, subdomain, consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead)
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
func (db *ControlDB) CreateSiteIdentity(ctx context.Context, q *controldb.Queries, siteID string, name string, role consts.IdentityRole) (*IdentityValues, error) {

	security := db.ServerConfig.Security
	hostname := db.ServerConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.MustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", name, siteID, hostname))
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
	if !can.HaveNightscoutToken(role) {
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
	totpSecret, encTOTPSecret, err := db.ServerConfig.Security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", identity.Name, identity.SiteID, db.ServerConfig.HostName))
	if err != nil {
		return nil, fmt.Errorf("cannot generate user TOTP: %w", err)
	}
	if _, err := db.Queries.UpdateSiteIdentityTOTP(ctx, controldb.UpdateSiteIdentityTOTPParams{TotpSecretEnc: encTOTPSecret,
		UpdatedByIdentityName: &by, SiteID: identity.SiteID, ID: identity.ID}); err != nil {
		return nil, err
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, identity.SiteID, identity.Name, "totp", by)
	return &IdentityValues{TotpSecret: &totpSecret}, nil
}

func (db *ControlDB) CreateControlIdentity(ctx context.Context, name string, role consts.ControlRole) (*IdentityValues, error) {
	security := db.ServerConfig.Security
	hostname := db.ServerConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate control password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""

	totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", name, consts.CONTROL_IDENTITY_SUBDOMAIN, hostname))
	if err != nil {
		return nil, fmt.Errorf("cannot generate control TOTP: %w", err)
	}

	totpSecretPtr = &totpSecret
	encTOTPSecret = encTOTPSecretRet

	err = db.Queries.CreateControlIdentity(ctx, controldb.CreateControlIdentityParams{
		Name:          name,
		PasswordHash:  hashedPassword,
		TotpSecretEnc: encTOTPSecret,
		Role:          role,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create control identity: %w", err)
	}

	return &IdentityValues{
		Password:   password,
		TotpSecret: totpSecretPtr,
	}, nil
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
