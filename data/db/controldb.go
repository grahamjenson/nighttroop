package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/security"
)

type ControlDB struct {
	SQLiteDB     *sql.DB
	serverConfig config.ServerConfig

	// previously modified
	sites map[string]*SiteDB

	sitesRWLock sync.RWMutex

	Queries *control.Queries
}

func NewControlDB(serverConfig config.ServerConfig) (*ControlDB, error) {
	var err error

	sqldb, err := OpenSqliteDB(utils.CONTROL_DB_PATH, 2000, serverConfig.IsTest())
	if err != nil {
		return nil, fmt.Errorf("cannot open control db: %w", err)
	}
	// Create a new control db
	controlDB := &ControlDB{
		SQLiteDB:     sqldb,
		sites:        make(map[string]*SiteDB),
		serverConfig: serverConfig,
	}

	// Migrate the DB, run any unrun migrations
	err = control.MigrateUp("control", sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot migrate control db: %w", err)
	}

	// Prepare the statements
	controlDB.Queries, err = control.Prepare(context.TODO(), sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare control statements: %w", err)
	}

	// If no admin user exists create one with a random password and TOTP secret and print the ENCRYPTED credentials to the console
	err = controlDB.seedDefaultIdentities()
	if err != nil {
		return nil, fmt.Errorf("cannot create default identities: %w", err)
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
	res, err := db.Queries.GetNotDeletedSites(context.TODO())

	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	for _, model := range res {
		sites = append(sites, model.Subdomain)
		// Site DB has to be available
		siteDB, err := NewSiteDB(context.TODO(), model.Subdomain, true, db.serverConfig)
		if err != nil {
			return sites, fmt.Errorf("SiteDB failed to be added to the ControlDB list, %w", err)
		}
		db.sites[model.Subdomain] = siteDB
	}

	return sites, nil
}

func (db *ControlDB) seedDefaultIdentities() error {
	db.CreateIdentity("control", consts.DEXCOM_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	db.CreateIdentity("control", consts.LIBRE_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	db.CreateIdentity("control", consts.NIGHTSCOUT_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)

	idValues, err := db.CreateIdentity("control", consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil && errors.IsUniqueViolationError(err) {
		// Admin user already exists so we can ignore this error
		return nil
	} else if err != nil {
		return err
	}

	bundle := fmt.Sprintf("Admin Username: %s\nAdmin Password: %s\nAdmin TOTP Secret: %s\n", consts.FIRST_ADMIN_NAME, idValues.Password, *idValues.TotpSecret)
	encBundle, err := db.serverConfig.Security.Encrypt(bundle)
	if err != nil {
		return fmt.Errorf("cannot encrypt admin bundle: %w", err)
	}

	fmt.Println("######## ADMIN SERVER BUNDLE ########")
	fmt.Printf("ENCRYPTED_ADMIN_BUNDLE=%s\n", encBundle)
	fmt.Println("######## END OF ADMIN SERVER BUNDLE ########")
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

func (db *ControlDB) CreateSite() (*SiteCreateResult, error) {
	subdomain := utils.NextSiteSubdomain(db.serverConfig.IsDev() || db.serverConfig.IsTest())
	if db.SubdomainUsed(subdomain) {
		// Try Again with an actual random ID
		subdomain = utils.NextSiteSubdomain(false)
		if db.SubdomainUsed(subdomain) {
			return nil, fmt.Errorf("site already exists")
		}
	}

	backupPassword, backupKeyHash, err := security.SecureCredentialsWithPrefix(subdomain + ":")
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	siteDB, err := NewSiteDB(context.TODO(), subdomain, false, db.serverConfig)
	if err != nil {
		siteDB.TryCloseAndDelete()
		return nil, fmt.Errorf("Site DB Creation Error: %w", err)
	}

	// Lock and write site to the db and add to the in memory map
	db.sitesRWLock.Lock()
	defer db.sitesRWLock.Unlock()

	_, err = db.Queries.CreateSite(context.TODO(), control.CreateSiteParams{
		Subdomain:     subdomain,
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

	adminIdentity, pwdIdentity, readOnlyIdentity, err := db.createDefaultIdentities(subdomain)

	if err != nil {
		siteDB.TryCloseAndDelete()
		return nil, fmt.Errorf("cannot create admin identity: %w", err)
	}

	db.sites[subdomain] = siteDB

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
	deletedSubdomains, err := db.Queries.DeleteSites(context.TODO(), control.DeleteSitesParams{
		JsonEach: string(bytes),
	})

	if err != nil {
		return fmt.Errorf("cannot delete user: %w", err)
	}

	fmt.Println("TRYING TO DELETE", subdomains, "DELETED", deletedSubdomains)

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

		path, backupPath := utils.SiteDBPaths(subdomain)

		err := utils.DeleteFiles([]string{path, backupPath})
		if err != nil {
			slog.Error("Failed to delete user db files", "subdomain", subdomain, "error", err)
		} else {
			slog.Info("Deleted user db files", "subdomain", subdomain)
		}
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
	exists, err := db.Queries.SubdomainExists(context.TODO(), control.SubdomainExistsParams{
		Subdomain: subdomain,
	})
	if err != nil {
		return false
	}
	return exists
}

func (db *ControlDB) UpdateSiteDateFlags(siteModel *control.Site, updateLastSeen bool, updateLastWrite bool) error {
	subdomain := siteModel.Subdomain

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := siteModel.LastWriteAt == nil || time.Now().After(siteModel.LastWriteAtDate().Add(10*time.Minute))
	lastSeenMoreThan10MinutesAgo := siteModel.LastSeenAt == nil || time.Now().After(siteModel.LastSeenAtDate().Add(10*time.Minute))

	now := time.Now().UnixMilli()
	var lastSeenAt *int64
	var lastWriteAt *int64
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		lastSeenAt = &now
	}
	if updateLastWrite && lastWriteMoreThan10MinutesAgo {
		lastWriteAt = &now
	}
	db.Queries.UpdateSiteDateFlags(context.TODO(), control.UpdateSiteDateFlagsParams{
		LastSeenAt:  lastSeenAt,
		LastWriteAt: lastWriteAt,
		Subdomain:   subdomain,
	})

	return nil
}

func (db *ControlDB) GetSite(subdomain string) (*SiteDB, *control.Site, error) {
	// Check if site exists in the db
	model, err := db.Queries.GetSite(context.TODO(), control.GetSiteParams{
		Subdomain: subdomain,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get site from db: %w", err)
	}
	if model.Subdomain == "" {
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

func (db *ControlDB) HasSitesInBatch(char string) (bool, error) {
	res, err := db.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return false, fmt.Errorf("cannot query sites: %w", err)
	}
	return len(res) > 0, nil
}

func (db *ControlDB) GetSiteBatch(char string) ([]*SiteDB, error) {

	sites := []*SiteDB{}
	res, err := db.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return sites, fmt.Errorf("cannot query sites: %w", err)
	}

	return getSiteDBsFromModels(db, res)
}

func (db *ControlDB) GetNeverSeenSites(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*SiteDB, error) {
	res, err := db.Queries.GetNeverSeenSites(context.TODO(), control.GetNeverSeenSitesParams{
		LastSeenAt:  lastSeenMinutes.Minutes(),
		LastWriteAt: lastWriteMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return getSiteDBsFromModels(db, res)
}

func (db *ControlDB) GetInactiveSites(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*SiteDB, error) {
	res, err := db.Queries.GetInactiveSites(context.TODO(), control.GetInactiveSitesParams{
		LastSeenAt:  lastSeenMinutes.Minutes(),
		LastWriteAt: lastWriteMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query inactive sites: %w", err)
	}

	return getSiteDBsFromModels(db, res)

}

func getSiteDBsFromModels(db *ControlDB, models []*control.Site) ([]*SiteDB, error) {
	sites := []*SiteDB{}
	db.sitesRWLock.RLock()
	defer db.sitesRWLock.RUnlock()

	for _, model := range models {
		site := db.sites[model.Subdomain]
		if site == nil {
			return sites, fmt.Errorf("site not found in memory")
		}
		sites = append(sites, site)
	}

	return sites, nil
}

func (db *ControlDB) GetAllSiteSyncSources(ctx context.Context, subdomain string) (nightscout *control.SiteSyncSource, dexcom *control.SiteSyncSource, err error) {
	// TODO: This might become a problem we could make faster with a dirty flag or something
	sync_sources, err := db.Queries.GetAllSiteSyncSourcesForSite(ctx, control.GetAllSiteSyncSourcesForSiteParams{
		Subdomain: subdomain,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error getting sync sources: %w", err)
	}

	for _, ss := range sync_sources {
		if ss.Type == consts.SyncSourceTypeNightscout {
			nightscout = ss
		}
		if ss.Type == consts.SyncSourceTypeDexcom {
			dexcom = ss
		}
	}

	return nightscout, dexcom, nil

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

	return db.SQLiteDB.Close()
}

// ////
// Identities
// ////

func (db *ControlDB) createDefaultIdentities(subdomain string) (*IdentityValues, *IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateIdentity(subdomain, consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateIdentity(subdomain, consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	// Create Default Read-Only User
	readOnlyIdentity, err := db.CreateIdentity(subdomain, consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead)
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

func (db *ControlDB) CreateIdentity(subdomain string, name string, role consts.IdentityRole) (*IdentityValues, error) {

	security := db.serverConfig.Security
	hostname := db.serverConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.MustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", name, subdomain, hostname))
		if err != nil {
			return nil, fmt.Errorf("cannot generate user TOTP: %w", err)
		}

		totpSecretPtr = &totpSecret
		encTOTPSecret = encTOTPSecretRet
	}

	err = db.Queries.CreateIdentity(context.TODO(), control.CreateIdentityParams{
		Subdomain:  subdomain,
		Name:       name,
		Password:   hashedPassword,
		TotpSecret: encTOTPSecret,
		Role:       role,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create user identity: %w", err)
	}
	var nsToken *string
	if can.HaveNightscoutToken(role) {
		token, err := security.CreateNightscoutTokenForIdentity(subdomain, name, string(role), password)
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

func (db *ControlDB) UpdateIdentityDateFlags(identity *control.Identity, updateLastSeen bool, updateLastWrite bool) error {

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := identity.LastWriteAt == nil || time.Now().After(identity.LastWriteAtDate().Add(10*time.Minute))
	lastSeenMoreThan10MinutesAgo := identity.LastSeenAt == nil || time.Now().After(identity.LastSeenAtDate().Add(10*time.Minute))

	now := time.Now().UnixMilli()
	var lastSeenAt *int64
	var lastWriteAt *int64
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		lastSeenAt = &now
	}
	if updateLastWrite && lastWriteMoreThan10MinutesAgo {
		lastWriteAt = &now
	}

	db.Queries.UpdateIdentityDateFlags(context.TODO(), control.UpdateIdentityDateFlagsParams{
		LastSeenAt:  lastSeenAt,
		LastWriteAt: lastWriteAt,
		ID:          identity.ID,
	})

	return nil
}

////
// Test methods
////

func (db *ControlDB) TestOnlyDeleteAllUsers(sites []*SiteDB) error {
	if len(sites) == 0 {
		return nil
	}
	db.sites = make(map[string]*SiteDB)
	err := db.Queries.TESTTruncateSites(context.TODO())

	return err
}
