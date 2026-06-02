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
	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

type ControlDB struct {
	SQLiteDB     *sql.DB
	serverConfig config.ServerConfig

	// previously modified
	users map[string]*UserDB

	usersRWLock sync.RWMutex

	Queries *control.Queries
}

func NewControlDB(serverConfig config.ServerConfig) (*ControlDB, error) {
	var err error

	sqldb, err := OpenSqliteDB(utils.CONTROL_DB_PATH, 2000)
	if err != nil {
		return nil, fmt.Errorf("cannot open control db: %w", err)
	}
	// Create a new control db
	controlDB := &ControlDB{
		SQLiteDB:     sqldb,
		users:        make(map[string]*UserDB),
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
	res, err := db.Queries.GetNotDeletedUsers(context.TODO())

	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	for _, model := range res {
		users = append(users, model.Subdomain)
		// User DB has to be avaialble
		userDB, err := NewUserDB(context.TODO(), model.Subdomain, true, db.serverConfig)
		if err != nil {
			return users, fmt.Errorf("UserDB failed to be added to the ControlDB list, %w", err)
		}
		db.users[model.Subdomain] = userDB
	}

	return users, nil
}

func (db *ControlDB) createAdminIfNoneExists() error {
	ctx := context.TODO()
	adminName := consts.FIRST_ADMIN_NAME
	model, err := db.Queries.GetAdmin(ctx, adminName)
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

func (db *ControlDB) CreateUser() (userdb *UserDB, adminPassword, backupPassword, totpSecret string, err error) {
	subdomain := utils.NextUserID(db.serverConfig.IsDev() || db.serverConfig.IsTest())
	if db.UserExists(subdomain) {
		// Try Again with an actual random ID
		subdomain = utils.NextUserID(false)
		if db.UserExists(subdomain) {
			return nil, "", "", "", fmt.Errorf("user already exists")
		}
	}

	backupPassword, backupKeyHash, err := security.SecureCredentialsWithPrefix(subdomain + ":")
	if err != nil {
		return nil, "", "", "", fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	userDB, err := NewUserDB(context.TODO(), subdomain, false, db.serverConfig)
	if err != nil {

		userDB.TryCloseAndDelete()
		return nil, "", "", "", fmt.Errorf("User DB Creation Error: %w", err)
	}

	// Lock and write user to the db and add to the in memory map
	db.usersRWLock.Lock()
	defer db.usersRWLock.Unlock()

	_, err = db.Queries.CreateUser(context.TODO(), subdomain, backupKeyHash)

	if err != nil {
		// Clean up the created DB file
		// If anything happens later the archiving process can clean it up

		userDB.TryCloseAndDelete()

		return nil, "", "", "", fmt.Errorf("cannot create user: %w", err)
	}

	////
	// Setup the new UserDB
	////

	adminPassword, totpSecret, err = createDefaultUsers(subdomain, db.serverConfig.HostName, userDB, db.serverConfig.Security)

	if err != nil {
		userDB.TryCloseAndDelete()
		return nil, "", "", "", fmt.Errorf("cannot create admin identity: %w", err)
	}

	db.users[subdomain] = userDB

	return userDB, adminPassword, backupPassword, totpSecret, nil
}

func createDefaultUsers(subdomain, hostname string, userDB *UserDB, security *security.SecurityConfig) (adminPassword, totpSecret string, err error) {
	adminPassword, hashedPassword, err := security.GeneratePassword()
	if err != nil {
		return "", "", fmt.Errorf("cannot generate admin password: %w", err)
	}

	totpSecret, encTOTPSecret, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", consts.FIRST_ADMIN_NAME, subdomain, hostname))
	if err != nil {
		return "", "", fmt.Errorf("cannot generate admin TOTP: %w", err)
	}

	err = userDB.Queries.CreateIdentity(context.TODO(), user.CreateIdentityParams{
		Name:       consts.FIRST_ADMIN_NAME,
		Password:   hashedPassword,
		TotpSecret: encTOTPSecret,
		Role:       consts.IdentityRoleAdmin,
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot create admin identity: %w", err)
	}

	err = userDB.Queries.CreateIdentity(context.TODO(), user.CreateIdentityParams{
		Name:       consts.DEXCOM_SYNC_IDENTITY_NAME,
		Password:   "",
		TotpSecret: "",
		Role:       consts.IdentityRoleSync,
	})
	if err != nil {
		return "", "", fmt.Errorf("cannot create dexcom sync identity: %w", err)
	}

	err = userDB.Queries.CreateIdentity(context.TODO(), user.CreateIdentityParams{
		Name:       consts.NIGHTSCOUT_SYNC_IDENTITY_NAME,
		Password:   "",
		TotpSecret: "",
		Role:       consts.IdentityRoleSync,
	})
	if err != nil {
		return "", "", fmt.Errorf("cannot create nightscout sync identity: %w", err)
	}

	err = userDB.Queries.CreateIdentity(context.TODO(), user.CreateIdentityParams{
		Name:       consts.LIBRE_SYNC_IDENTITY_NAME,
		Password:   "",
		TotpSecret: "",
		Role:       consts.IdentityRoleSync,
	})
	if err != nil {
		return "", "", fmt.Errorf("cannot create libre sync identity: %w", err)
	}
	return adminPassword, totpSecret, nil
}

////
// Delete USER methods
////

// TODO think through implications of errors at different places in this method
func (db *ControlDB) DeleteUsers(subdomains []string) error {
	// Set the deleted at timestammp in the control DB
	bytes, _ := json.Marshal(subdomains)
	deletedSubdomains, err := db.Queries.DeleteUsers(context.TODO(), string(bytes))
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

		path, backupPath := utils.UserDBPaths(subdomain)

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

func (db *ControlDB) AllUsers() []*UserDB {
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	users := []*UserDB{}
	for _, user := range db.users {
		users = append(users, user)
	}
	return users
}

func (db *ControlDB) UserExists(subdomain string) bool {
	db.usersRWLock.RLock()
	_, ok := db.users[subdomain]
	db.usersRWLock.RUnlock()
	return ok
}

func (db *ControlDB) GetUser(subdomain string, updateLastSeen bool) (*UserDB, *control.User, error) {
	// Check if user exists in the db
	model, err := db.Queries.GetUser(context.TODO(), subdomain)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get user: %w", err)
	}
	if model.Subdomain == "" {
		return nil, nil, fmt.Errorf("user not found")
	}

	lastSeenMoreThan10MinutesAgo := model.LastSeenAt == nil || time.Now().After(model.LastSeenAtDate().Add(-10*time.Minute))
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		db.Queries.UpdateUserLastSeenAt(context.TODO(), subdomain)
	}

	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	user := db.users[subdomain]
	if user == nil {
		return nil, nil, fmt.Errorf("user not found in memory")
	}

	return user, model, nil
}

func (db *ControlDB) HasUsersInBatch(char string) (bool, error) {
	res, err := db.Queries.GetBatchUsers(context.TODO(), char)
	if err != nil {
		return false, fmt.Errorf("cannot query users: %w", err)
	}
	return len(res) > 0, nil
}

func (db *ControlDB) GetUserBatch(char string) ([]*UserDB, error) {

	users := []*UserDB{}
	res, err := db.Queries.GetBatchUsers(context.TODO(), char)
	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	return getUserDBsFromModels(db, res)
}

func (db *ControlDB) GetNeverSeenUsers(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*UserDB, error) {
	res, err := db.Queries.GetNeverSeenUsers(context.TODO(), lastSeenMinutes.Minutes(), lastWriteMinutes.Minutes())
	if err != nil {
		return nil, fmt.Errorf("cannot query never seen users: %w", err)
	}

	return getUserDBsFromModels(db, res)
}

func (db *ControlDB) GetInactiveUsers(lastSeenMinutes time.Duration, lastWriteMinutes time.Duration) ([]*UserDB, error) {
	res, err := db.Queries.GetInactiveUsers(context.TODO(), lastSeenMinutes.Minutes(), lastWriteMinutes.Minutes())
	if err != nil {
		return nil, fmt.Errorf("cannot query inactive users: %w", err)
	}

	return getUserDBsFromModels(db, res)

}

func getUserDBsFromModels(db *ControlDB, models []*control.User) ([]*UserDB, error) {
	users := []*UserDB{}
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

func (db *ControlDB) GetAllUserSyncSources(ctx context.Context, subdomain string) (nightscout *control.UserSyncSource, dexcom *control.UserSyncSource, err error) {
	// TODO: This might become a problem we could make faster with a dirty flag or something
	sync_sources, err := db.Queries.GetAllUserSyncSourcesForUser(ctx, subdomain)
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

////
// Test methods
////

func (db *ControlDB) TestOnlyDeleteAllUsers(users []*UserDB) error {
	if len(users) == 0 {
		return nil
	}
	db.users = make(map[string]*UserDB)
	err := db.Queries.TESTTruncateUsers(context.TODO())

	return err
}
