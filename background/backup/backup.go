// Make the backgorund jobs orthogonal to the DB
package backup

import (
	"context"
	"fmt"
	"log/slog"

	"time"

	"nighttroop.t1d.nz/utils"
)

// This is the last method called to backup both control and user dbs
func (b *BackupJobManager) BackupToS3() {
	err := b.BackupControl()
	if err != nil {
		slog.Error("Backup error", "error", err)
	}

	for _, char := range utils.USER_CHAR_SET {
		err := b.BackupUserBatch(string(char))

		if err != nil {
			slog.Error("Error backing up user batch", "error", err, "char", char)
			continue
		}
	}
}

// //
// Control Backup
// //

func (b *BackupJobManager) BackupControl() error {

	// Backup the Database
	err := utils.SqliteBackup(b.ctx, b.controlDB.SQLiteDB, utils.CONTROL_DB_PATH, utils.CONTROL_DB_BACKUP_PATH)
	if err != nil {
		return fmt.Errorf("cannot backup control db: %w", err)
	}

	// Copy the control db to S3 path
	err = b.S3Client.UploadFilesToS3(b.S3Client.ControlS3Path(), []string{utils.CONTROL_DB_BACKUP_PATH})
	if err != nil {
		return fmt.Errorf("cannot sync control db: %w", err)
	}

	return nil
}

// ///
// User Backup
// ///

func (b *BackupJobManager) BackupUserBatch(char string) error {
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

	slog.Info("Backing up user batch", "char", char, "user_count", len(users))

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

	err = b.S3Client.UploadFilesToS3(b.S3Client.UserBatchS3Path(char), paths)
	if err != nil {
		return fmt.Errorf("cannot sync user pool: %w %v", err, string(char))
	}

	// Copy the control db to the backup location
	return nil
}
