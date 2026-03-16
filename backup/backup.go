// Make the backgorund jobs orthogonal to the DB
package backup

import (
	"context"
	"errors"
	"fmt"
	"math"

	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"nighttroop.t1d.nz/config"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/utils"
)

type BackupJobs struct {
	controlDB *db.ControlDB
	ctx       context.Context

	ControlBackupInterval time.Duration
	BatchBackupInterval   time.Duration
	MaxUserAge            time.Duration

	cancel func()
	wg     sync.WaitGroup

	S3Client *utils.S3Client
}

// BackupJobs
// Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupJobs(config config.ServerConfig, controlDB *db.ControlDB, s3Client *utils.S3Client) *BackupJobs {

	jobs := &BackupJobs{
		controlDB:             controlDB,
		ControlBackupInterval: 10 * time.Second, // TODO replace with config
		BatchBackupInterval:   60 * time.Second, // TODO replace with config
		MaxUserAge:            30 * time.Minute, // TODO replace with config

		S3Client: s3Client,
	}

	jobs.ctx, jobs.cancel = context.WithCancel(context.Background())

	return jobs
}

func (b *BackupJobs) Start() {
	// Start backup Jobs
	// CotrolDB backup and Sync Loop
	if b.ControlBackupInterval > 0 {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			b.controlBackupLoop()
		}()
	}

	if b.BatchBackupInterval > 0 {
		b.wg.Add(1)
		// The goal here is over the BatchBackupInterval we will backup all pools
		// But we will stagger them so that we do not swamp the CPU or network
		staggeredInterval := b.BatchBackupInterval / time.Duration(utils.USER_BATCH_LENGTH)
		go func() {
			defer b.wg.Done()
			b.userBackupLoop(staggeredInterval)
		}()
	}

}

// //
// Control Backup
// //

func (b *BackupJobs) controlBackupLoop() {
	syncTicker := time.NewTicker(b.ControlBackupInterval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-b.ctx.Done():
			return
		case <-syncTicker.C:
		}

		// Backup DB to S3
		if err := b.backupControl(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}

	}

}

func (b *BackupJobs) backupControl() error {
	ctx := context.TODO()

	// Backup the Database
	err := utils.SqliteBackup(ctx, b.controlDB.SQLiteDB, utils.CONTROL_DB_PATH, utils.CONTROL_DB_BACKUP_PATH)
	if err != nil {
		return fmt.Errorf("cannot backup control db: %w", err)
	}

	// Copy the control db to S3 path
	err = b.S3Client.UploadToS3(ctx, b.S3Client.ControlS3Path(), utils.CONTROL_DB_BACKUP_PATH)
	if err != nil {
		return fmt.Errorf("cannot sync control db: %w", err)
	}

	return nil
}

// ///
// User Backup
// ///

func (b *BackupJobs) userBackupLoop(interval time.Duration) {
	syncTicker := time.NewTicker(interval)
	defer syncTicker.Stop()

	index := 0
	for {
		// Wait for ticker or context close.
		select {
		case <-b.ctx.Done():
			return
		case <-syncTicker.C:
		}
		char := utils.USER_CHAR_SET[index]

		// Backup the dbs, upload to S3
		err := b.BackupUserBatch(string(char))

		if err != nil {
			fmt.Println("Error backing up user batch", "error", err, "char", char)
			continue
		}

		// TODO soft delete users from the DB (for rehydrations later)
		removedUsers, err := b.controlDB.archive(db.S3Client, b.MaxUserAge)
		if err != nil {
			fmt.Println("Error archiving user pool", "error", err)
		} else {
			err := b.DeleteAllUsers(removedUsers)
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
		index = (index + 1) % utils.USER_BATCH_LENGTH
	}
}

func (b *BackupJobs) BackupUserBatch(char string) error {
	defer utils.TrackTime(time.Now(), "backup user pool "+char)
	users, err := b.controlDB.GetUserBatch(char)
	if err != nil {
		return fmt.Errorf("cannot get user batch: %w", err)
	}
	if len(users) == 0 {
		return nil
	}

	paths := []string{}

	for _, user := range users {
		// Have to add it anyway because we need all paths
		paths = append(paths, user.BackupPath)
	}

	err = s3Client.SyncAllFilesToSingleLZ4(up.S3Path, paths)
	if err != nil {
		return fmt.Errorf("cannot sync user pool: %w %v", err, string(up.char))
	}

	// Copy the control db to the backup location
	return nil
}

func (b *BackupJobs) BackupToS3() error {
	defer utils.TrackTime(time.Now(), "ForceBackupSync")

	b.UpdateModified()
	// Go through every user and force a WAL dump with `PRAGMA wal_checkpoint(FULL);`
	// Then check to see if they have changed (not sure if the wal_checkpoint itself causes modified date to change)
	err := b.backup()
	if err != nil {
		fmt.Println("Error forcing backup:", err)
		return fmt.Errorf("cannot force backup: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())

	// Limit the number of concurrent goroutines to the number of CPU cores - 1
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("ForceBackupSync running", len(b.userPools), limit)
	for _, pool := range b.allPools() {

		g.Go(func() error {
			_, err := pool.updateModified()
			if err != nil {
				fmt.Println("Error updating user modified", "error", err)
			}

			return pool.backup(b.S3Client)
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

// This method looks though all the ages of the DBs and removes user that are older than the max age from the circulation
// TODO It should then group these user up together, zip them and then upload them to S3 while logging their location for future hydration
func archiveUsers(s3Client *S3Client, maxUserAge time.Duration) ([]*UserDB, error) {
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

func (b *BackupJobs) Stop() {

	// This will send a cancel signal to context
	b.cancel()
	// This will wait
	b.wg.Wait()

}
