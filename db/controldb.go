package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	mtsql "nighttroop.t1d.nz/sql"
)

type ControlDB struct {
	*SQLDB

	// previously modified
	ControlBackupInterval time.Duration
	PoolBackupInterval    time.Duration
	MaxUserAge            time.Duration

	userPools map[byte]*UserPool

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	S3Client *S3Client
}

func NewControlDB(endpoint, bucket string) (*ControlDB, error) {
	var err error

	// Create a new control db
	controlDB := &ControlDB{
		SQLDB: &SQLDB{
			Path:       "data/db/control.sqlite",
			BackupPath: "data/backup/control.sqlite",
		},
		ControlBackupInterval: 10 * time.Second,
		PoolBackupInterval:    60 * time.Second,
		MaxUserAge:            30 * time.Minute,
	}
	controlDB.ctx, controlDB.cancel = context.WithCancel(context.Background())

	controlDB.S3Client, err = NewS3Client(endpoint, bucket)
	if err != nil {
		return nil, fmt.Errorf("cannot create s3 client: %w", err)
	}

	// Create a pool for each starting character of a user subdomain
	// This map will never be added or removed from so no need for a lock
	controlDB.userPools = map[byte]*UserPool{}
	for _, c := range userCharSet {
		controlDB.userPools[byte(c)] = NewUserPool(byte(c))
	}

	return controlDB, nil
}

func (db *ControlDB) TryRestoreDBFile() error {
	// If it already exists dont download
	if db.DBFileExists() {
		return nil
	}
	os.MkdirAll("data/backup", os.ModePerm)
	os.MkdirAll("data/db", os.ModePerm)
	found, err := db.S3Client.DownloadFile(db.BackupPath, db.Path)
	if !found && err == nil {
		fmt.Println("No control db backup found")
		return nil
	}
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}

func (db *ControlDB) RestoreAndOpenAllUsers() error {
	defer trackTime(time.Now(), "RestoreAndOpenAllUsers")
	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("RestoreAndOpenAllUsers running", len(db.userPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			userSubdomains, err := db.GetUserPoolUsersFromDB(pool.char)
			fmt.Println("UserSubdomains", string(pool.char), userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot get user pool users: %w", err)
			}
			err = pool.RestoreAndOpenAllUsers(db.S3Client, userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot restore user pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the RestoreAndOpenAllUsers  tasks failed: %v\n", err)
		return fmt.Errorf("cannot restore and open all users: %w", err)
	} else {
		fmt.Println("All RestoreAndOpenAllUsers tasks succeeded!")
	}

	return nil
}

func (db *ControlDB) GetUserPoolUsersFromDB(char byte) ([]string, error) {
	users := []string{}
	res, err := db.Query("SELECT subdomain FROM users WHERE pool = ? AND deleted_at IS NULL ORDER BY subdomain DESC", int(char))
	if err != nil {
		return users, fmt.Errorf("cannot query users: %w", err)
	}

	defer res.Close()

	for res.Next() {
		var subdomain string
		err = res.Scan(&subdomain)
		if err != nil {
			return users, fmt.Errorf("cannot scan user: %w", err)
		}
		users = append(users, subdomain)
	}
	err = res.Err()
	if err != nil {
		return users, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

func (db *ControlDB) GetNUsersDBs() int {
	total := 0
	for _, pool := range db.userPools {
		total += pool.GetNUsersDBs()
	}
	return total
}

func (db *ControlDB) GetUser(subdomain string, updateLastSeen bool) (*UserDB, error) {
	if subdomain == "" {
		return nil, fmt.Errorf("subdomain cannot be empty")
	}

	fmt.Println("GetUser", subdomain)

	// See if the user is already there
	tp := db.userPools[subdomain[0]]
	user := tp.GetUser(subdomain)
	userInDB, err := db.UserInDB(subdomain, updateLastSeen)

	fmt.Println("UserInDB", userInDB, err)
	if err != nil {
		fmt.Println("Error getting user from DB", err)
		return nil, fmt.Errorf("cannot get user from DB: %w", err)
	}

	// Absolute User not in the ControlDB we dont return
	if !userInDB {
		fmt.Println("User not found in DB")
		return nil, fmt.Errorf("user not found in DB")
	}

	if user == nil {
		fmt.Println("User not found in pool, opening user")
		return nil, fmt.Errorf("user Not Found: %w", err)
	}

	return user, nil
}

func (db *ControlDB) UserInDB(subdomain string, updateLastSeen bool) (bool, error) {
	// Check if user exists in the db
	rows, err := db.Query("SELECT last_seen_at FROM users WHERE subdomain = ? AND deleted_at IS NULL LIMIT 1", subdomain)
	if err != nil {
		return false, fmt.Errorf("cannot get user: %w", err)
	}
	if rows == nil {
		return false, fmt.Errorf("user not found")
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return false, fmt.Errorf("rows error: %w", err)
	}

	rows.Next()

	var last_seen_at string
	err = rows.Scan(&last_seen_at)
	if err != nil {
		return false, fmt.Errorf("cannot scan user: %w", err)
	}
	fmt.Println("Found User: ", last_seen_at)

	if updateLastSeen {
		err = db.TryUpdateLastSeenAt(subdomain, last_seen_at)
		// Not going to break because this errors
		if err != nil {
			fmt.Println("Error updating last seen at", err)
		}
	}

	return true, nil
}

func (db *ControlDB) TryUpdateLastSeenAt(subdomain string, last_seen_at string) error {

	last_seen_at_time, err := time.Parse(time.RFC3339, last_seen_at)
	if err != nil {
		fmt.Println("Error parsing last seen at", err)
		return fmt.Errorf("cannot parse last seen at: %w", err)

	}

	// Only update if the last seen at is older than 10 minutes
	if last_seen_at_time.Add(10 * time.Minute).After(time.Now()) {
		return nil
	}

	// Try update lsat seen at, I am not going to throw here if not updated
	fmt.Println("Trying to updating last seen at")
	_, err = db.Exec("UPDATE users SET last_seen_at = ? WHERE subdomain = ?", time.Now().Format(time.RFC3339), subdomain)
	if err != nil {
		fmt.Println("Error updating last seen at", err)
		return fmt.Errorf("cannot update last seen at: %w", err)
	}
	fmt.Println("Last seen at updated")
	return nil
}

func (db *ControlDB) CreateUser() (string, string, error) {
	userID := NextUserID()
	_, err := db.Exec(mtsql.ControlUserInsert, userID, int(userID[0]))
	if err != nil {
		return "", "", fmt.Errorf("cannot create user: %w", err)
	}

	tp := db.userPools[userID[0]]
	_, apiKey, err := tp.CreateUser(userID)
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

func (db *ControlDB) StartBackupLoops() {
	// CotrolDB backup and Sync Loop
	if db.ControlBackupInterval > 0 {
		db.wg.Add(1)
		go func() {
			defer db.wg.Done()
			db.backupLoop()
		}()
	}

	if db.PoolBackupInterval > 0 {
		db.wg.Add(1)
		// The goal here is over the PoolBackupInterval we will backup all pools
		// But we will stagger them so that we do not swamp the CPU or network
		staggeredInterval := db.PoolBackupInterval / time.Duration(len(db.userPools))
		go func() {
			defer db.wg.Done()
			db.tpBackupLoop(staggeredInterval)
		}()
	}

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

func (db *ControlDB) tpBackupLoop(interval time.Duration) {
	syncTicker := time.NewTicker(interval)
	defer syncTicker.Stop()

	allPools := db.allPools()

	index := 0
	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}
		tp := allPools[index]

		// Force all users to sync WAL to disk and update their mod times
		// this also backs up the user to the backup path if modified
		modified, err := tp.updateModified()
		if err != nil {
			fmt.Println("Error updating user modified", "error", err)
		}

		// TODO soft delete users from the DB (for rehydrations later)
		removedUsers, err := tp.archive(db.S3Client, db.MaxUserAge)
		if err != nil {
			fmt.Println("Error archiving user pool", "error", err)
		} else {
			err := db.DeleteAllUsers(removedUsers)
			if err != nil {
				fmt.Println("Error deleting users from DB", "error", err)
			}

		}

		if modified {
			// If any user was modified we need to backup the pool
			if err = tp.backup(db.S3Client); err != nil && !errors.Is(err, context.Canceled) {
				fmt.Println("Backup error", "error", err)
			}
		}
		index = (index + 1) % len(allPools)
	}
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

func (db *ControlDB) backupLoop() {
	syncTicker := time.NewTicker(db.ControlBackupInterval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}

		// Backup DB to S3
		if err := db.backup(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}

	}

}

func (db *ControlDB) backup() error {

	var err error
	modified, err := db.UpdateModified()
	if err != nil {
		fmt.Println("Error updating control db modified", err)
		return fmt.Errorf("cannot update control db modified: %w", err)
	}

	if modified {
		err = db.S3Client.SyncAllFiles([]string{db.BackupPath})
		if err != nil {
			return fmt.Errorf("cannot sync control db: %w", err)
		}
	}

	// Copy the control db to the backup location
	return nil
}

func (db *ControlDB) StopBackupLoop() {
	// This will send a cancel signal to context
	db.cancel()
	// This will wait
	db.wg.Wait()
}

func (db *ControlDB) ForceBackupSync() error {
	defer trackTime(time.Now(), "ForceBackupSync")

	db.UpdateModified()
	// Go through every user and force a WAL dump with `PRAGMA wal_checkpoint(FULL);`
	// Then check to see if they have changed (not sure if the wal_checkpoint itself causes modified date to change)
	err := db.backup()
	if err != nil {
		fmt.Println("Error forcing backup:", err)
		return fmt.Errorf("cannot force backup: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())

	// Limit the number of concurrent goroutines to the number of CPU cores - 1
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("ForceBackupSync running", len(db.userPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			_, err := pool.updateModified()
			if err != nil {
				fmt.Println("Error updating user modified", "error", err)
			}

			return pool.backup(db.S3Client)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the ForceBackupSync tasks failed: %v\n", err)
		return fmt.Errorf("cannot sync all users: %w", err)
	} else {
		fmt.Println("All ForceBackupSync tasks succeeded!")
	}

	return nil
}
