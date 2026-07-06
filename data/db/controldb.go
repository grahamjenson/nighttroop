package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

type ControlDB struct {
	SQLiteDB     *sql.DB
	serverConfig config.ServerConfig

	// previously modified
	users map[string]*SiteDB

	usersRWLock sync.RWMutex

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
		users:        make(map[string]*SiteDB),
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
	err = controlDB.createAdminIfNoneExists()
	if err != nil {
		return nil, fmt.Errorf("cannot create admin user: %w", err)
	}

	// This will go over all non-deleted users and add them to the control db users
	_, err = controlDB.fillUsers()
	if err != nil {
		return nil, fmt.Errorf("cannot fill users: %w", err)
	}
	return controlDB, nil
}

// This will fetch all the users in
func (db *ControlDB) fillUsers() ([]string, error) {
	users := []string{}
	res, err := db.Queries.GetNotDeletedSites(context.TODO())

	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	for _, model := range res {
		users = append(users, model.Subdomain)
		// User DB has to be avaialble
		userDB, err := NewSiteDB(context.TODO(), model.Subdomain, true, db.serverConfig)
		if err != nil {
			return users, fmt.Errorf("SiteDB failed to be added to the ControlDB list, %w", err)
		}
		db.users[model.Subdomain] = userDB
	}

	return users, nil
}

func (db *ControlDB) createAdminIfNoneExists() error {
	ctx := context.TODO()
	adminName := consts.FIRST_ADMIN_NAME
	model, err := db.Queries.GetAdmin(ctx, control.GetAdminParams{
		Name: adminName,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("cannot query admin: %w", err)
	}
	if model.Name == adminName {
		return nil
	}

	password, hashedPassword, err := db.serverConfig.Security.GeneratePassword()
	if err != nil {
		return fmt.Errorf("cannot generate admin password: %w", err)
	}

	totpSecret, encTOTPSecret, err := db.serverConfig.Security.TOTPGenerate(fmt.Sprintf("%s@%s", adminName, db.serverConfig.HostName))
	if err != nil {
		return fmt.Errorf("cannot generate admin TOTP: %w", err)
	}

	_, err = db.Queries.CreateAdmin(ctx, control.CreateAdminParams{
		Name:       adminName,
		Password:   hashedPassword,
		Role:       "admin",
		TotpSecret: encTOTPSecret,
	})

	if err != nil {
		return fmt.Errorf("cannot create admin: %w", err)
	}

	bundle := fmt.Sprintf("Admin Username: %s\nAdmin Password: %s\nAdmin TOTP Secret: %s\n", adminName, password, totpSecret)
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

type UserCreateMaterial struct {
	Subdomain     string
	BackupKeyHash string
	Password      string
}

type UserCreateResult struct {
	SiteDB         *SiteDB
	AdminPassword  string
	TotpSecret     string
	PwdNSToken     string
	BackupPassword string
}

func (db *ControlDB) CreateUser() (*UserCreateResult, error) {
	subdomain := utils.NextSiteSubdomain(db.serverConfig.IsDev() || db.serverConfig.IsTest())
	if db.SubdomainUsed(subdomain) {
		// Try Again with an actual random ID
		subdomain = utils.NextSiteSubdomain(false)
		if db.SubdomainUsed(subdomain) {
			return nil, fmt.Errorf("user already exists")
		}
	}

	backupPassword, backupKeyHash, err := security.SecureCredentialsWithPrefix(subdomain + ":")
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	userDB, err := NewSiteDB(context.TODO(), subdomain, false, db.serverConfig)
	if err != nil {

		userDB.TryCloseAndDelete()
		return nil, fmt.Errorf("User DB Creation Error: %w", err)
	}

	// Lock and write user to the db and add to the in memory map
	db.usersRWLock.Lock()
	defer db.usersRWLock.Unlock()

	_, err = db.Queries.CreateSite(context.TODO(), control.CreateSiteParams{
		Subdomain:     subdomain,
		BackupKeyHash: backupKeyHash,
	})

	if err != nil {
		// Clean up the created DB file
		// If anything happens later the archiving process can clean it up

		userDB.TryCloseAndDelete()

		return nil, fmt.Errorf("cannot create user: %w", err)
	}

	////
	// Setup the new SiteDB
	////

	adminIdentity, pwdIdentity, err := db.createDefaultIdentities(subdomain)

	if err != nil {
		userDB.TryCloseAndDelete()
		return nil, fmt.Errorf("cannot create admin identity: %w", err)
	}

	db.users[subdomain] = userDB

	return &UserCreateResult{
		SiteDB:         userDB,
		AdminPassword:  adminIdentity.Password,
		TotpSecret:     *adminIdentity.TotpSecret,
		PwdNSToken:     *pwdIdentity.NSToken,
		BackupPassword: backupPassword,
	}, nil
}

////
// Delete USER methods
////

// TODO think through implications of errors at different places in this method
func (db *ControlDB) DeleteUsers(subdomains []string) error {
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
		db.usersRWLock.Lock()
		user, ok := db.users[subdomain]
		if !ok {
			db.usersRWLock.Unlock()
			return fmt.Errorf("user not found")
		}
		delete(db.users, subdomain)
		db.usersRWLock.Unlock()

		user.Close()

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

func (db *ControlDB) GetNUsersDBs() int {
	l := 0
	db.usersRWLock.RLock()
	l = len(db.users)
	db.usersRWLock.RUnlock()
	return l
}

func (db *ControlDB) AllUsers() []*SiteDB {
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	users := []*SiteDB{}
	for _, user := range db.users {
		users = append(users, user)
	}
	return users
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

func (db *ControlDB) UpdateUserDateFlags(userModel *control.Site, updateLastSeen bool, updateLastWrite bool) error {
	subdomain := userModel.Subdomain

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := userModel.LastWriteAt == nil || time.Now().After(userModel.LastWriteAtDate().Add(-10*time.Minute))
	lastSeenMoreThan10MinutesAgo := userModel.LastSeenAt == nil || time.Now().After(userModel.LastSeenAtDate().Add(-10*time.Minute))

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

func (db *ControlDB) GetUser(subdomain string) (*SiteDB, *control.Site, error) {
	// Check if user exists in the db
	model, err := db.Queries.GetSite(context.TODO(), control.GetSiteParams{
		Subdomain: subdomain,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get user from db: %w", err)
	}
	if model.Subdomain == "" {
		return nil, nil, fmt.Errorf("user not found")
	}

	userDb, err := db.getSiteDBFromMap(subdomain)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot get SiteDB from memory: %w", err)
	}
	return userDb, model, nil
}

func (db *ControlDB) getSiteDBFromMap(subdomain string) (*SiteDB, error) {
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	userdb := db.users[subdomain]
	if userdb == nil {
		return nil, fmt.Errorf("user not found in memory")
	}

	return userdb, nil
}

func (db *ControlDB) HasUsersInBatch(char string) (bool, error) {
	res, err := db.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return false, fmt.Errorf("cannot query sites: %w", err)
	}
	return len(res) > 0, nil
}

func (db *ControlDB) GetUserBatch(char string) ([]*SiteDB, error) {

	users := []*SiteDB{}
	res, err := db.Queries.GetBatchSites(context.TODO(), control.GetBatchSitesParams{
		Batch: char,
	})
	if err != nil {
		return users, fmt.Errorf("cannot query sites: %w", err)
	}

	return getSiteDBsFromModels(db, res)
}

func (db *ControlDB) GetNeverSeenUsers(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*SiteDB, error) {
	res, err := db.Queries.GetNeverSeenSites(context.TODO(), control.GetNeverSeenSitesParams{
		LastSeenAt:  lastSeenMinutes.Minutes(),
		LastWriteAt: lastWriteMinutes.Minutes(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen sites: %w", err)
	}

	return getSiteDBsFromModels(db, res)
}

func (db *ControlDB) GetInactiveUsers(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*SiteDB, error) {
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
	users := []*SiteDB{}
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	for _, model := range models {
		user := db.users[model.Subdomain]
		if user == nil {
			return users, fmt.Errorf("user not found in memory")
		}
		users = append(users, user)
	}

	return users, nil
}

func (db *ControlDB) GetAllUserSyncSources(ctx context.Context, subdomain string) (nightscout *control.SiteSyncSource, dexcom *control.SiteSyncSource, err error) {
	// TODO: This might become a problem we could make faster with a dirty flag or something
	sync_sources, err := db.Queries.GetAllUserSyncSourcesForUser(ctx, control.GetAllUserSyncSourcesForUserParams{
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
// End of Get USER Methods
////

func (db *ControlDB) Close() error {
	// Keep it locked so no new users can be created
	for _, user := range db.users {
		err := user.Close()
		if err != nil {
			slog.Error("Failed to close user db", "subdomain", user.Subdomain, "error", err)
		}
	}

	return db.SQLiteDB.Close()
}

// ////
// Identities
// ////

func (db *ControlDB) createDefaultIdentities(subdomain string) (*IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateIdentity(subdomain, consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateIdentity(subdomain, consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	////
	// Create Sync Identities (used to trail)
	////
	_, err = db.CreateIdentity(subdomain, consts.DEXCOM_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create dexcom sync identity: %w", err)
	}

	_, err = db.CreateIdentity(subdomain, consts.NIGHTSCOUT_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create nightscout sync identity: %w", err)
	}

	_, err = db.CreateIdentity(subdomain, consts.LIBRE_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create libre sync identity: %w", err)
	}

	return adminIdentity, pwdIdentity, nil
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

	err = db.Queries.CreateSiteIdentity(context.TODO(), control.CreateSiteIdentityParams{
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

func (db *ControlDB) UpdateIdentityDateFlags(identity *control.SiteIdentity, updateLastSeen bool, updateLastWrite bool) error {

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := identity.LastWriteAt == nil || time.Now().After(identity.LastWriteAtDate().Add(-10*time.Minute))
	lastSeenMoreThan10MinutesAgo := identity.LastSeenAt == nil || time.Now().After(identity.LastSeenAtDate().Add(-10*time.Minute))

	now := time.Now().UnixMilli()
	var lastSeenAt *int64
	var lastWriteAt *int64
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		lastSeenAt = &now
	}
	if updateLastWrite && lastWriteMoreThan10MinutesAgo {
		lastWriteAt = &now
	}

	db.Queries.UpdateSiteIdentityDateFlags(context.TODO(), control.UpdateSiteIdentityDateFlagsParams{
		LastSeenAt:  lastSeenAt,
		LastWriteAt: lastWriteAt,
		ID:          identity.ID,
	})

	return nil
}

////
// Test methods
////

func (db *ControlDB) TestOnlyDeleteAllUsers(users []*SiteDB) error {
	if len(users) == 0 {
		return nil
	}
	db.users = make(map[string]*SiteDB)
	err := db.Queries.TESTTruncateSites(context.TODO())

	return err
}
