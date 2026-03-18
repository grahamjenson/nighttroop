// Make the backgorund jobs orthogonal to the DB
package backup

import (
	"context"
	"fmt"

	"sync"
	"time"

	"nighttroop.t1d.nz/config"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/utils"
)

type BackupJobs struct {
	controlDB *db.ControlDB
	ctx       context.Context

	ControlBackupInterval time.Duration
	BatchBackupInterval   time.Duration
	MaxUserInactive       time.Duration

	cancel func()
	wg     sync.WaitGroup

	S3Client *utils.S3Client
}

// BackupJobs
// Takes a ControlDB and backs up the db and all user dbs to a file and syncs to S3
func SetupBackupJobs(config config.ServerConfig, controlDB *db.ControlDB, s3Client *utils.S3Client) *BackupJobs {

	jobs := &BackupJobs{
		controlDB:             controlDB,
		ControlBackupInterval: time.Duration(config.ControlBackupIntervalInSeconds) * time.Second,
		BatchBackupInterval:   time.Duration(config.BatchBackupIntervalInSeconds) * time.Second,
		MaxUserInactive:       time.Duration(config.MaxUserInactiveInMins) * time.Minute,

		S3Client: s3Client,
	}

	jobs.ctx, jobs.cancel = context.WithCancel(context.Background())

	return jobs
}

func (b *BackupJobs) Start() {
	// Start backup Jobs
	// CotrolDB backup and Sync Loop
	if b.ControlBackupInterval > 0 {
		b.wg.Go(func() {
			b.controlBackupLoop()
		})
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

// This is the last method called to backup both control and user dbs
func (b *BackupJobs) BackupToS3() {
	err := b.backupControl()
	if err != nil {
		fmt.Println("Backup error", "error", err)
	}

	for _, char := range utils.USER_CHAR_SET {
		err := b.BackupUserBatch(string(char))

		if err != nil {
			fmt.Println("Error backing up user batch", "error", err, "char", char)
			continue
		}
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
		err := b.backupControl()
		if err != nil {
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
	err = b.S3Client.UploadToS3(ctx, utils.CONTROL_DB_BACKUP_PATH, b.S3Client.ControlS3Path())
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

		index = (index + 1) % utils.USER_BATCH_LENGTH

		// Go over all users once per cycle (at the end)
		if index == 0 {
			err := b.ArchiveInactiveUsers()
			if err != nil {
				fmt.Println("Error archiving", "error", err, "char", char)
				continue
			}
		}

	}
}

func (b *BackupJobs) BackupUserBatch(char string) error {
	trackTrim := false
	defer utils.TrackTime(time.Now(), "backup user pool "+char, &trackTrim, nil)
	ctx := context.TODO()

	users, err := b.controlDB.GetUserBatch(char)

	if err != nil {
		return fmt.Errorf("cannot get user batch: %w", err)
	}
	if len(users) == 0 {
		return nil
	}

	fmt.Printf("Backing up user batch %s with %d users\n", char, len(users))

	paths := []string{}

	for _, user := range users {
		if user.Subdomain == "" {
			continue
		}
		path, backupPath := utils.UserDBPaths(user.Subdomain)

		// Have to add it anyway because we need all paths
		err := utils.SqliteBackup(ctx, user.SQLiteDB, path, backupPath)
		if err != nil {
			return fmt.Errorf("cannot backup user db: %w", user.Subdomain, err)
		}

		paths = append(paths, backupPath)
	}

	trackTrim = len(paths) > 0

	err = b.S3Client.SyncAllFilesToSingleLZ4(b.S3Client.UserBatchS3Path(char), paths)
	if err != nil {
		return fmt.Errorf("cannot sync user pool: %w %v", err, string(char))
	}

	// Copy the control db to the backup location
	return nil
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
