package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"nighttroop.t1d.nz/sql/control"
	controlsql "nighttroop.t1d.nz/sql/control"
	"nighttroop.t1d.nz/sql/migrations"
	"nighttroop.t1d.nz/utils"
)

type ControlDB struct {
	SQLiteDB *sql.DB

	// previously modified
	users map[string]*UserDB

	usersRWLock sync.RWMutex

	// prepared statements
	UsersCreate         *controlsql.UsersCreateStmt
	UsersGet            *controlsql.UsersGetStmt
	UsersGetBatch       *controlsql.UsersGetBatchStmt
	UsersAll            *controlsql.UsersAllStmt
	UsersGetInactive    *controlsql.UsersGetInactiveStmt
	UsersUpdateLastSeen *controlsql.UsersUpdateLastSeenStmt
	UsersDelete         *controlsql.UsersDeleteStmt

	// For Tests
	TESTONLYUsersTruncateAll *controlsql.UsersTruncateAllStmt
}

func NewControlDB() (*ControlDB, error) {
	var err error

	sqldb, err := OpenSqliteDB(utils.CONTROL_DB_PATH, 2000)
	if err != nil {
		return nil, fmt.Errorf("cannot open control db: %w", err)
	}
	// Create a new control db
	controlDB := &ControlDB{
		SQLiteDB: sqldb,
		users:    make(map[string]*UserDB),
	}

	// Migrate the DB, run any unrun migrations
	err = migrations.MigrateControl(context.TODO(), sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot migrate control db: %w", err)
	}

	var errs []error

	// Create the prepared statements
	controlDB.UsersCreate = control.NewUsersCreate(sqldb, &errs)
	controlDB.UsersDelete = control.NewUsersDelete(sqldb, &errs)
	controlDB.UsersGet = control.NewUsersGet(sqldb, &errs)
	controlDB.UsersAll = control.NewUsersAll(sqldb, &errs)
	controlDB.UsersGetBatch = control.NewUsersGetBatch(sqldb, &errs)
	controlDB.UsersUpdateLastSeen = control.NewUsersUpdateLastSeen(sqldb, &errs)
	controlDB.TESTONLYUsersTruncateAll = control.NewUsersTruncateAll(sqldb, &errs)
	controlDB.UsersGetInactive = control.NewUsersGetInactive(sqldb, &errs)

	if finalErr := errors.Join(errs...); finalErr != nil {
		return nil, finalErr
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
	res, err := db.UsersAll.Query(context.TODO())

	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	for _, model := range res {
		users = append(users, model.Subdomain)
		// User DB has to be avaialble
		userDB, err := NewUserDB(context.TODO(), model.Subdomain, true)
		if err != nil {
			return users, fmt.Errorf("UserDB failed to be added to the ControlDB list, %w", err)
		}
		db.users[model.Subdomain] = userDB
	}

	return users, nil
}

func (db *ControlDB) CreateUser() (*UserDB, error) {
	subdomain := utils.NextUserID()
	backupKeyHash, err := utils.GenerateSecureAPIKey(12)
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key hash: %w", err)
	}

	_, err = db.UsersCreate.Exec(context.TODO(), subdomain, backupKeyHash)
	if err != nil {
		return nil, fmt.Errorf("cannot create user: %w", err)
	}

	userDB, err := NewUserDB(context.TODO(), subdomain, false)
	if err != nil {
		return nil, fmt.Errorf("user db file already exists")
	}

	// TODO create API token

	// Add the user to the pool
	db.usersRWLock.Lock()
	db.users[subdomain] = userDB
	db.usersRWLock.Unlock()

	return userDB, nil
}

////
// Delete USER methods
////

func (db *ControlDB) DeleteUser(subdomain string) error {
	db.usersRWLock.Lock()
	user, ok := db.users[subdomain]
	if !ok {
		db.usersRWLock.Unlock()
		return fmt.Errorf("user not found")
	}
	delete(db.users, subdomain)
	db.usersRWLock.Unlock()

	user.Close()

	_, err := db.UsersDelete.Exec(context.TODO(), subdomain)
	if err != nil {
		return fmt.Errorf("cannot delete user: %w", err)
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

func (db *ControlDB) AllInactiveUsers(maxAge time.Duration) ([]*UserDB, error) {
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	inactiveUsers, err := db.UsersGetInactive.Query(context.TODO(), maxAge.Minutes())
	if err != nil {
		return nil, fmt.Errorf("InactiveUser Error %w", err)
	}
	users := []*UserDB{}
	for _, iu := range inactiveUsers {
		user := db.users[iu.Subdomain]
		users = append(users, user)
	}

	return users, nil
}

func (db *ControlDB) GetUser(subdomain string, updateLastSeen bool) (*UserDB, error) {
	// Check if user exists in the db
	model, err := db.UsersGet.QueryRow(context.TODO(), subdomain)
	if err != nil {
		return nil, fmt.Errorf("cannot get user: %w", err)
	}
	if model.Subdomain == "" {
		return nil, fmt.Errorf("user not found")
	}
	fmt.Println("Found User: ", model.Subdomain, model.LastSeenAt)
	lastSeenMoreThan10MinutesAgo := model.LastSeenAt.Add(10 * time.Minute).After(time.Now())
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		db.UsersUpdateLastSeen.Exec(context.TODO(), subdomain, time.Now())
	}

	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	user := db.users[subdomain]
	if user == nil {
		return nil, fmt.Errorf("user not found in memory")
	}

	return user, nil
}

func (db *ControlDB) GetUserBatch(char string) ([]*UserDB, error) {

	users := []*UserDB{}
	res, err := db.UsersGetBatch.Query(context.TODO(), char)
	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	for _, model := range res {
		user := db.users[model.Subdomain]
		if user == nil {
			return users, fmt.Errorf("user not found in memory")
		}
		users = append(users, user)
	}

	return users, nil
}

func (db *ControlDB) GetInactiveUsers(duration time.Duration) ([]*UserDB, error) {
	res, err := db.UsersGetInactive.Query(context.TODO(), duration.Minutes())
	if err != nil {
		return nil, fmt.Errorf("cannot query inactive users: %w", err)
	}

	users := []*UserDB{}
	db.usersRWLock.RLock()
	defer db.usersRWLock.RUnlock()

	for _, model := range res {
		user := db.users[model.Subdomain]
		if user == nil {
			return users, fmt.Errorf("user not found in memory")
		}
		users = append(users, user)
	}

	return users, nil

}

////
// End of Get USER Methods
////

func (db *ControlDB) Close() error {
	// Keep it locked so no new users can be created
	for _, user := range db.users {
		user.Close()
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
	_, err := db.TESTONLYUsersTruncateAll.Exec(context.TODO())

	return err
}
