package db

import (
	"fmt"
	"sort"
	"sync"
	"time"

	mtsql "nighttroop.t1d.nz/sql"
)

////
// User Pool
////

type UserPool struct {
	user   map[string]*UserDB
	rwLock sync.RWMutex

	S3Path string
	Path   string
	char   byte
}

func NewUserPool(char byte) *UserPool {
	return &UserPool{
		char:   char,
		user:   make(map[string]*UserDB),
		S3Path: "userpool/" + string(char) + ".tar.lz4",
		Path:   "data/db/user/" + string(char),
	}
}

func (up *UserPool) GetNUsersDBs() int {
	l := 0
	up.rwLock.RLock()
	l = len(up.user)
	up.rwLock.RUnlock()
	return l
}

func (up *UserPool) allUsers() []*UserDB {
	users := make([]*UserDB, 0)
	up.rwLock.RLock()
	for _, user := range up.user {
		users = append(users, user)
	}
	up.rwLock.RUnlock()
	sort.Slice(users, func(i, j int) bool {
		return users[i].Subdomain < users[j].Subdomain
	})
	return users
}

func (up *UserPool) GetUser(subdomain string) *UserDB {
	up.rwLock.RLock()
	user := up.user[subdomain]
	up.rwLock.RUnlock()

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

func (up *UserPool) CreateUser(subdomain string) (*UserDB, string, error) {
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

	return userDB, apiKey, nil
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

func (up *UserPool) backup(s3Client *S3Client) error {
	user := up.allUsers()
	if len(user) == 0 {
		return nil
	}
	defer trackTime(time.Now(), "backup user pool "+string(up.char))

	var err error
	paths := []string{}

	for _, user := range user {
		// Have to add it anyway because we need all paths
		paths = append(paths, user.BackupPath)
	}

	if len(paths) == 0 {
		fmt.Println("No user to backup")
		return nil
	}

	err = s3Client.SyncAllFilesToSingleLZ4(up.S3Path, paths)
	if err != nil {
		return fmt.Errorf("cannot sync user pool: %w %v", err, string(up.char))
	}

	// Copy the control db to the backup location
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
