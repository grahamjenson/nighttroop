package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

type ControlDB struct {
	*controldb.DB

	// previously modified
	sites map[string]*SiteDB

	sitesRWLock sync.RWMutex
	// createSiteLock serialises CreateSite and HydrateSite so two calls never pick, open or delete the same subdomain's DB
	createSiteLock sync.Mutex

	Jobs *JobsDB
	Logs *LogsDB
	// Live is what the server keeps only while it runs, like each identity's alarms
	Live *livedb.DB

	setup setup
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
		sites: make(map[string]*SiteDB),
	}

	// Without a control admin the server starts in setup, see control_setup.go
	if err := controlDB.startSetupIfNoAdmin(ctx); err != nil {
		return nil, fmt.Errorf("cannot check for the control admin: %w", err)
	}

	// This will go over all non-deleted sites and add them to the control db sites
	_, err = controlDB.fillSites()
	if err != nil {
		return nil, fmt.Errorf("cannot fill sites: %w", err)
	}

	return controlDB, nil
}

// This will fetch all the sites in
func (db *ControlDB) fillSites() ([]string, error) {
	sites := []string{}
	res, err := db.Queries.GetLiveAndDisabledSites(context.TODO())

	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	for _, model := range res {
		sites = append(sites, model.ID)

		// Site DB has to be available
		path := sitedb.GetPath(model.ID, db.ServerConfig)

		if !utils.FileExists(path) {
			return nil, fmt.Errorf("The Site DB %s Must exist and doesnt", model.ID)
		}

		siteDB, err := NewSiteDB(context.TODO(), model.ID, db.ServerConfig)
		if err != nil {
			return sites, fmt.Errorf("SiteDB failed to be added to the ControlDB list, %w", err)
		}
		db.sites[model.ID] = siteDB
	}

	return sites, nil
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

func (db *ControlDB) CreateSite() (*SiteCreateResult, error) {
	return db.createSite("")
}

// CreateSiteWithID creates a site at a site ID picked for it, such as a vanity one for a special user. It is 8
// lowercase letters like every site's, and one already taken is refused.
func (db *ControlDB) CreateSiteWithID(subdomain string) (*SiteCreateResult, error) {
	if !utils.SyntacticallyValidateSiteID(subdomain) {
		return nil, fmt.Errorf("site ID %q is not %d lowercase letters", subdomain, utils.SITE_ID_LENGTH)
	}
	return db.createSite(subdomain)
}

// createSite creates a site at subdomain, or at a random one when it is ""
func (db *ControlDB) createSite(subdomain string) (*SiteCreateResult, error) {
	db.createSiteLock.Lock()
	defer db.createSiteLock.Unlock()

	if subdomain != "" {
		if db.SubdomainUsed(subdomain) {
			return nil, ErrSiteIDTaken
		}
	} else {
		subdomain = utils.NextSiteSubdomain(db.ServerConfig.IsDevOrTest())
		if db.SubdomainUsed(subdomain) {
			// Try Again with an actual random ID
			subdomain = utils.NextSiteSubdomain(false)
			if db.SubdomainUsed(subdomain) {
				return nil, fmt.Errorf("site already exists")
			}
		}
	}

	backupPassword, backupKeyHash, err := security.SecureCredentialsWithPrefix(subdomain+":", db.ServerConfig.GetBcryptCost())
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	siteDB, err := NewSiteDB(context.TODO(), subdomain, db.ServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Site DB Creation Error: %w", err)
	}

	// Lock and write site to the db and add to the in memory map
	db.sitesRWLock.Lock()
	defer db.sitesRWLock.Unlock()

	_, err = db.Queries.CreateSite(context.TODO(), controldb.CreateSiteParams{
		ID:            subdomain,
		BackupKeyHash: backupKeyHash,
	})

	if err != nil {
		// Clean up the created DB file
		// If anything happens later the archiving process can clean it up

		siteDB.TryCloseAndDelete()

		return nil, fmt.Errorf("cannot create site: %w", err)
	}

	////
	// Setup the new SiteDB
	////

	adminIdentity, pwdIdentity, readOnlyIdentity, err := db.createSiteDefaultIdentities(subdomain)

	if err != nil {
		siteDB.TryCloseAndDelete()
		return nil, fmt.Errorf("cannot create admin identity: %w", err)
	}

	db.sites[subdomain] = siteDB
	db.Logs.SiteCreatedEvent(context.TODO(), subdomain)

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

// TODO think through implications of errors at different places in this method
func (db *ControlDB) DeleteSites(subdomains []string) error {
	// Set the deleted at timestammp in the control DB
	if len(subdomains) == 0 {
		return nil
	}
	bytes, _ := json.Marshal(subdomains)
	deletedSubdomains, err := db.Queries.DeleteSites(context.TODO(), controldb.DeleteSitesParams{
		JsonEach: string(bytes),
	})

	if err != nil {
		return fmt.Errorf("cannot delete site: %w", err)
	}

	slog.Info("Deleting sites", "requested", subdomains, "deleted", deletedSubdomains)

	// Close the DB and remove from the map
	for _, subdomain := range deletedSubdomains {
		db.sitesRWLock.Lock()
		site, ok := db.sites[subdomain]
		if !ok {
			db.sitesRWLock.Unlock()
			return fmt.Errorf("site not found")
		}
		delete(db.sites, subdomain)
		db.sitesRWLock.Unlock()

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
	db.sitesRWLock.RLock()
	site, ok := db.sites[subdomain]
	db.sitesRWLock.RUnlock()
	if !ok {
		return fmt.Errorf("site %s is not open", subdomain)
	}
	if err := site.Backup(ctx); err != nil {
		return fmt.Errorf("cannot back up site %s: %w", subdomain, err)
	}
	if err := upload(ctx, sitedb.GetS3ArchivePath(subdomain, db.ServerConfig), s3utils.SiteRetention, []string{site.BackupPath}); err != nil {
		return fmt.Errorf("cannot upload the archive of site %s: %w", subdomain, err)
	}
	bytes, _ := json.Marshal([]string{subdomain})
	if _, err := db.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: string(bytes)}); err != nil {
		return fmt.Errorf("cannot archive site %s: %w", subdomain, err)
	}

	db.sitesRWLock.Lock()
	delete(db.sites, subdomain)
	db.sitesRWLock.Unlock()
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
	db.sitesRWLock.RLock()
	_, open := db.sites[subdomain]
	db.sitesRWLock.RUnlock()
	if open {
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

	db.sitesRWLock.Lock()
	db.sites[subdomain] = siteDB
	db.sitesRWLock.Unlock()
	return nil
}

// TruncateSite clears the raw JSON a site's DB keeps of what each row older than TruncateRecordsAfterDays was sent,
// keeping all its data, then shrinks the DB and records its new size, which it returns
func (db *ControlDB) TruncateSite(ctx context.Context, subdomain string) (*int64, error) {
	db.sitesRWLock.RLock()
	site, ok := db.sites[subdomain]
	db.sitesRWLock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("site %s is not open", subdomain)
	}
	q, before := site.Queries, time.Now().AddDate(0, 0, -db.ServerConfig.TruncateRecordsAfterDays).UnixMilli()
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
	l := 0
	db.sitesRWLock.RLock()
	l = len(db.sites)
	db.sitesRWLock.RUnlock()
	return l
}

func (db *ControlDB) AllSites() []*SiteDB {
	db.sitesRWLock.RLock()
	defer db.sitesRWLock.RUnlock()

	sites := []*SiteDB{}
	for _, site := range db.sites {
		sites = append(sites, site)
	}
	return sites
}

func (db *ControlDB) SubdomainUsed(subdomain string) bool {
	exists, err := db.Queries.SiteExists(context.TODO(), controldb.SiteExistsParams{
		ID: subdomain,
	})
	if err != nil {
		return false
	}
	return exists
}

// UpdateSiteAndIdentityDateFlags records when the site and identity were last seen and written to, at most every
// 10 minutes, and the site DB's size with its last write
func (db *ControlDB) UpdateSiteAndIdentityDateFlags(siteModel *controldb.Site, siteDB *SiteDB, identityModel *controldb.SiteIdentity, updateLastSeen bool, updateLastWrite bool) error {
	subdomain := siteModel.ID

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := siteModel.LastWriteAt == nil || time.Now().After(siteModel.LastWriteAtDate().Add(10*time.Minute))
	lastSeenMoreThan10MinutesAgo := siteModel.LastSeenAt == nil || time.Now().After(siteModel.LastSeenAtDate().Add(10*time.Minute))

	now := time.Now().UnixMilli()
	var lastSeenAt *int64
	var lastWriteAt *int64
	var estimatedDBSizeKB *int64
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		lastSeenAt = &now
	}
	if updateLastWrite && lastWriteMoreThan10MinutesAgo {
		lastWriteAt = &now
		estimatedDBSizeKB = siteDB.EstimatedSizeKB()
	}
	db.Queries.UpdateSiteDateFlags(context.TODO(), controldb.UpdateSiteDateFlagsParams{
		LastSeenAt:        lastSeenAt,
		LastWriteAt:       lastWriteAt,
		EstimatedDbSizeKb: estimatedDBSizeKB,
		ID:                subdomain,
	})

	db.Queries.UpdateSiteIdentityDateFlags(context.TODO(), controldb.UpdateSiteIdentityDateFlagsParams{
		LastSeenAt:  lastSeenAt,
		LastWriteAt: lastWriteAt,
		ID:          identityModel.ID,
	})

	return nil
}

func (db *ControlDB) GetSite(subdomain string) (*SiteDB, *controldb.Site, error) {
	// Check if site exists in the db
	model, err := db.Queries.GetSite(context.TODO(), controldb.GetSiteParams{
		ID: subdomain,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get site from db: %w", err)
	}
	if model.ID == "" {
		return nil, nil, fmt.Errorf("site not found")
	}

	siteDB, err := db.getSiteDBFromMap(subdomain)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get SiteDB from memory: %w", err)
	}
	return siteDB, model, nil
}

func (db *ControlDB) getSiteDBFromMap(subdomain string) (*SiteDB, error) {
	db.sitesRWLock.RLock()
	defer db.sitesRWLock.RUnlock()

	siteDB := db.sites[subdomain]
	if siteDB == nil {
		return nil, fmt.Errorf("site not found in memory")
	}

	return siteDB, nil
}

////
// End of Get SITE Methods
////

func (db *ControlDB) Close() error {
	// Keep it locked so no new sites can be created
	for _, site := range db.sites {
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

func (db *ControlDB) createSiteDefaultIdentities(subdomain string) (*IdentityValues, *IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateSiteIdentity(context.TODO(), db.Queries, subdomain, consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateSiteIdentity(context.TODO(), db.Queries, subdomain, consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	// Create Default Read-Only User
	readOnlyIdentity, err := db.CreateSiteIdentity(context.TODO(), db.Queries, subdomain, consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead)
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

	err = q.CreateSiteIdentity(ctx, controldb.CreateSiteIdentityParams{
		SiteID:        siteID,
		Name:          name,
		PasswordHash:  hashedPassword,
		TotpSecretEnc: encTOTPSecret,
		Role:          role,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create user identity: %w", err)
	}
	var nsToken *string
	if can.HaveNightscoutToken(role) {
		token, err := security.CreateNightscoutTokenForIdentity(siteID, name, string(role), password)
		if err != nil {
			return nil, fmt.Errorf("cannot create nightscout token for identity: %w", err)
		}
		nsToken = &token
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
	if _, err := db.Queries.UpdateSiteIdentityPassword(ctx, controldb.UpdateSiteIdentityPasswordParams{PasswordHash: hashedPassword,
		UpdatedByIdentityName: &by, SiteID: identity.SiteID, ID: identity.ID}); err != nil {
		return nil, err
	}
	db.Logs.IdentityCredentialsChangedEvent(ctx, identity.SiteID, identity.Name, "password", by)
	values := &IdentityValues{Password: password}
	if can.HaveNightscoutToken(identity.Role) {
		token, err := db.ServerConfig.Security.CreateNightscoutTokenForIdentity(identity.SiteID, identity.Name, string(identity.Role), password)
		if err != nil {
			return nil, fmt.Errorf("cannot create nightscout token for identity: %w", err)
		}
		values.NSToken = &token
	}
	return values, nil
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

func (db *ControlDB) CreateControlIdentity(name string, role consts.ControlRole) (*IdentityValues, error) {
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

	err = db.Queries.CreateControlIdentity(context.TODO(), controldb.CreateControlIdentityParams{
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

func (db *ControlDB) GetSiteDBsFromModels(models []*controldb.Site) ([]*SiteDB, error) {
	sites := []*SiteDB{}
	db.sitesRWLock.RLock()
	defer db.sitesRWLock.RUnlock()

	for _, model := range models {
		site := db.sites[model.ID]
		if site == nil {
			return sites, fmt.Errorf("site not found in memory")
		}
		sites = append(sites, site)
	}

	return sites, nil
}
