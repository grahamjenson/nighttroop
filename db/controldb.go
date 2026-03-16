package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"nighttroop.t1d.nz/sql/control"
	controlsql "nighttroop.t1d.nz/sql/control"
	"nighttroop.t1d.nz/sql/migrations"
	"nighttroop.t1d.nz/utils"
)

type ControlDB struct {
	sqldb *sql.DB

	// previously modified
	users map[string]*UserDB

	usersRWLock sync.RWMutex

	// prepared statements
	UsersCreate         *controlsql.UsersCreateStmt
	UsersGet            *controlsql.UsersGetStmt
	UsersGetBatch       *controlsql.UsersGetBatchStmt
	UsersAll            *controlsql.UsersAllStmt
	UsersUpdateLastSeen *controlsql.UsersUpdateLastSeenStmt

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
		sqldb: sqldb,
		users: make(map[string]*UserDB),
	}

	// Migrate the DB, run any unrun migrations
	err = migrations.MigrateControl(context.TODO(), sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot migrate control db: %w", err)
	}

	// Create the prepared statements
	controlDB.UsersCreate, err = control.NewUsersCreate(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users create statement: %w", err)
	}

	controlDB.UsersGet, err = control.NewUsersGet(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users get statement: %w", err)
	}

	controlDB.UsersAll, err = control.NewUsersAll(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users get all statement: %w", err)
	}

	controlDB.UsersGetBatch, err = control.NewUsersGetBatch(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users get batch statement: %w", err)
	}

	controlDB.UsersUpdateLastSeen, err = control.NewUsersUpdateLastSeen(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users update last seen statement: %w", err)
	}

	controlDB.TESTONLYUsersTruncateAll, err = control.NewUsersTruncateAll(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users truncate all statement: %w", err)
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

func (db *ControlDB) CreateUser() (*UserDB, error) {
	subdomain := utils.NextUserID()

	_, err := db.UsersCreate.Exec(context.TODO(), subdomain)
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

func (db *ControlDB) Close() error {
	// Keep it locked so no new users can be created
	for _, user := range db.users {
		user.Close()
	}

	return db.sqldb.Close()
}

func (db *ControlDB) TestOnlyDeleteAllUsers(users []*UserDB) error {
	if len(users) == 0 {
		return nil
	}
	_, err := db.TESTONLYUsersTruncateAll.Exec(context.TODO())

	return err
}

func (db *ControlDB) GetNUsersDBs() int {
	l := 0
	db.usersRWLock.RLock()
	l = len(db.users)
	db.usersRWLock.RUnlock()
	return l
}
