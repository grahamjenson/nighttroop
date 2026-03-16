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
	PoolBackupInterval    time.Duration
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
		PoolBackupInterval:    60 * time.Second, // TODO replace with config
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
			b.backupLoop()
		}()
	}

	if b.PoolBackupInterval > 0 {
		b.wg.Add(1)
		// The goal here is over the PoolBackupInterval we will backup all pools
		// But we will stagger them so that we do not swamp the CPU or network
		staggeredInterval := b.PoolBackupInterval / time.Duration(len(b.userPools))
		go func() {
			defer b.wg.Done()
			b.tpBackupLoop(staggeredInterval)
		}()
	}

}
func (b *BackupJobs) backupLoop() {
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
		if err := b.backup(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}

	}

}

func (b *BackupJobs) tpBackupLoop(interval time.Duration) {
	syncTicker := time.NewTicker(interval)
	defer syncTicker.Stop()

	allPools := db.allPools()

	index := 0
	for {
		// Wait for ticker or context close.
		select {
		case <-b.ctx.Done():
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
		removedUsers, err := tp.archive(db.S3Client, b.MaxUserAge)
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
		index = (index + 1) % len(allPools)
	}
}

func (b *BackupJobs) backup() error {

	var err error
	modified, err := b.UpdateModified()
	if err != nil {
		fmt.Println("Error updating control db modified", err)
		return fmt.Errorf("cannot update control db modified: %w", err)
	}

	if modified {
		err = b.S3Client.SyncAllFiles([]string{b.BackupPath})
		if err != nil {
			return fmt.Errorf("cannot sync control db: %w", err)
		}
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

func (up *UserPool) backup(s3Client *S3Client) error {
	user := up.allUsers()
	if len(user) == 0 {
		return nil
	}
	defer utils.TrackTime(time.Now(), "backup user pool "+string(up.char))

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

func (b *BackupJobs) Stop() {

	// This will send a cancel signal to context
	b.cancel()
	// This will wait
	b.wg.Wait()

}
