package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	mtsql "nighttroop.t1d.nz/sql"
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
	UsersUpdateLastSeen *controlsql.UsersUpdateLastSeenStmt
}

func NewControlDB(endpoint, bucket string) (*ControlDB, error) {
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

	controlDB.UsersGetBatch, err = control.NewUsersGetBatch(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users get batch statement: %w", err)
	}

	controlDB.UsersUpdateLastSeen, err = control.NewUsersUpdateLastSeen(sqldb)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare users update last seen statement: %w", err)
	}

	return controlDB, nil
}

// This will fetch all the users in
func (db *ControlDB) FillUsers(char byte) ([]string, error) {
	users := []string{}
	res, err := db.UsersGetBatch.Query(context.TODO(), string(char))

	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	for _, model := range res {
		users = append(users, model.Subdomain)
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

	return user, nil
}

func (db *ControlDB) CreateUser() (string, string, error) {
	userID := NextUserID()
	_, err := db.Exec(mtsql.ControlUserInsert, userID, int(userID[0]))
	if err != nil {
		return "", "", fmt.Errorf("cannot create user: %w", err)
	}

	tp := db.userPools[userID[0]]
	var err error
	userDB := NewUserDB(subdomain)
	if userDB.DBFileExists() {
		return nil, "", fmt.Errorf("user db file already exists")
	}

	// Create the DB file
	err = userDB.Open(&mtsql.UserCreate)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create user db: %w", err)
	}

	apiKey, err := userDB.CreateAPIToken()
	if err != nil {
		userDB.Close()
		return nil, "", fmt.Errorf("cannot create user db: %w", err)
	}

	// Add the user to the pool
	up.rwLock.Lock()
	up.user[subdomain] = userDB
	up.rwLock.Unlock()

	if err != nil {
		return "", "", fmt.Errorf("cannot create user: %w", err)
	}
	return userID, apiKey, nil
}

func (db *ControlDB) Close() error {
	// Keep it locked so no new users can be created
	for _, pool := range db.allPools() {
		pool.Close()
	}

	return db.sqldb.Close()
}

func (db *ControlDB) allPools() []*UserPool {
	var allPools []*UserPool
	for _, pool := range db.userPools {
		allPools = append(allPools, pool)
	}

	sort.Slice(allPools, func(i, j int) bool {
		return allPools[i].char < allPools[j].char
	})
	return allPools
}

func (db *ControlDB) DeleteAllUsers(users []*UserDB) error {
	if len(users) == 0 {
		return nil
	}
	for _, user := range users {
		_, err := db.Exec(mtsql.ControlUserDelete, user.Subdomain)
		if err != nil {
			fmt.Println("cannot delete user: %w", err)
		}
	}
	return nil
}

// USERS POOL METHODS
func (db *ControlDB) GetUserBatchPaths(pool byte) (string, string) {
	S3Path := "userpool/" + string(pool) + ".tar.lz4"
	Path := "data/db/user/" + string(pool)
	return S3Path, Path
}

func (db *ControlDB) GetNUsersDBs() int {
	l := 0
	db.usersRWLock.RLock()
	l = len(db.users)
	db.usersRWLock.RUnlock()
	return l
}

func (db *ControlDB) allUsers() []*UserDB {
	users := make([]*UserDB, 0)
	db.usersRWLock.RLock()
	for _, user := range db.users {
		users = append(users, user)
	}
	db.usersRWLock.RUnlock()
	sort.Slice(users, func(i, j int) bool {
		return users[i].Subdomain < users[j].Subdomain
	})
	return users
}

func (db *ControlDB) GetUser(subdomain string) *UserDB {
	db.usersRWLock.RLock()
	user := db.users[subdomain]
	db.usersRWLock.RUnlock()

	return user
}

func (up *UserPool) RestoreAndOpenAllUsers(s3Client *S3Client, subdomains []string) error {
	if len(subdomains) == 0 {
		return nil
	}

	var err error
	// Check to see if user are already on disk
	for _, subdomain := range subdomains {
		_, err := up.OpenUser(subdomain)
		if err != nil {
			fmt.Printf("Error opening user db %v %v %v\n", string(up.char), subdomain, err)
			return fmt.Errorf("cannot open user db: %w", err)
		}
		// We ignore if the DB was not found to be created
	}

	restored := up.GetNUsersDBs()
	if restored > 0 {
		if restored != len(subdomains) {
			fmt.Printf("Not all user were restored for Tennant Pool %v %v %v\n", string(up.char), restored, len(subdomains))
			return fmt.Errorf("not all user were restored")
		}
		// If they should all be on disk if any are
		return nil
	}

	// Now we need to see if we want to restore from S3
	fmt.Println("Trying to restore user from backup for pool", string(up.char))

	found, err := s3Client.DownloadAndExtractLZ4(up.S3Path, up.Path)
	fmt.Println("Finished s3 download backup for pool", string(up.char))
	if !found && err == nil {
		fmt.Println("No user backup found")
		return nil
	}
	if err != nil {
		// TODO dont fail here if just not found
		return fmt.Errorf("cannot download user db: %w", err)
	}

	for _, subdomain := range subdomains {

		// At this point in time we are ignoring errors
		user, err := up.OpenUser(subdomain)
		if err != nil {
			fmt.Printf("Error opening user db %v %v %v\n", string(up.char), subdomain, err)
			return fmt.Errorf("cannot open user db: %w", err)
		}
		if user == nil {
			fmt.Printf("User db %v %v not found\n", string(up.char), subdomain)
			return fmt.Errorf("cannot open user db: %w", err)
		}
	}

	return nil

}

func (up *UserPool) OpenUser(subdomain string) (*UserDB, error) {
	userDB := NewUserDB(subdomain)
	if userDB.DBFileExists() {
		// If the DB file is there, write the user
		err := userDB.Open(nil)
		if err != nil {
			return nil, fmt.Errorf("cannot open user db: %w", err)
		}
		// User db opened and ready to use
		up.rwLock.Lock()
		up.user[subdomain] = userDB
		up.rwLock.Unlock()
		return userDB, nil
	}
	return nil, nil
}

func (up *UserPool) Close() error {
	up.rwLock.Lock()
	defer up.rwLock.Unlock()

	for _, user := range up.user {
		err := user.Close()
		if err != nil {
			fmt.Println("cannot close user db", err)
		}
	}

	return nil
}

func (up *UserPool) updateModified() (bool, error) {

	user := up.allUsers()
	if len(user) == 0 {
		return false, nil
	}

	anyModified := false
	for _, user := range user {
		modified, err := user.UpdateModified()
		if err != nil {
			// TODO error group
			fmt.Println("cannot update user modified db:", user.Subdomain, err)
		}
		if modified {
			anyModified = true
		}
	}
	return anyModified, nil
}

// This method looks though all the ages of the DBs and removes user that are older than the max age from the circulation
// TODO It should then group these user up together, zip them and then upload them to S3 while logging their location for future hydration
func (up *UserPool) archive(s3Client *S3Client, maxUserAge time.Duration) ([]*UserDB, error) {
	removeList := []*UserDB{}

	user := up.allUsers()
	if len(user) == 0 {
		return removeList, nil
	}

	for _, user := range user {
		// Check if the user is older than the max age
		lm := time.Unix(user.lastModifiedAt, 0)
		if time.Since(lm) > maxUserAge {
			// Archive the user
			removeList = append(removeList, user)
		}
	}
	up.rwLock.Lock()
	defer up.rwLock.Unlock()
	for _, user := range removeList {
		fmt.Println("Removing user", user.Subdomain)
		delete(up.user, user.Subdomain)
		err := user.Close()
		if err != nil {
			fmt.Println("cannot close user db", err)
		}

		err = s3Client.SyncAllFilesToSingleLZ4(user.ArchivePath(), []string{user.Path})
		if err != nil {
			fmt.Println("cannot sync to s3 user archive", err)
		}
	}

	return removeList, nil
}

//
//
//
