// Make the backgorund jobs orthogonal to the DB
package backup

import (
	"context"
	"fmt"
	"log/slog"

	"time"

	"nighttroop.t1d.nz/utils"
)

// This is the last method called to backup both control and site dbs
func (b *BackupJobManager) BackupToS3() {
	err := b.BackupControl()
	if err != nil {
		slog.Error("Backup error", "error", err)
	}

	for _, char := range utils.SITE_CHAR_SET {
		err := b.BackupSiteBatch(string(char))

		if err != nil {
			slog.Error("Error backing up site batch", "error", err, "char", char)
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
// Site Backup
// ///

func (b *BackupJobManager) BackupSiteBatch(char string) error {
	trackTrim := false
	defer utils.TrackTime(time.Now(), "backup site pool "+char, &trackTrim, nil)
	ctx := context.TODO()

	sites, err := b.controlDB.GetSiteBatch(char)

	if err != nil {
		return fmt.Errorf("cannot get site batch: %w", err)
	}
	if len(sites) == 0 {
		return nil
	}

	slog.Info("Backing up site batch", "char", char, "site_count", len(sites))

	paths := []string{}

	for _, site := range sites {
		if site.Subdomain == "" {
			continue
		}
		path, backupPath := utils.SiteDBPaths(site.Subdomain)

		// Have to add it anyway because we need all paths
		err := utils.SqliteBackup(ctx, site.SQLiteDB, path, backupPath)
		if err != nil {
			return fmt.Errorf("cannot backup site db: %w", site.Subdomain, err)
		}

		paths = append(paths, backupPath)
	}

	trackTrim = len(paths) > 0

	err = b.S3Client.UploadFilesToS3(b.S3Client.SiteBatchS3Path(char), paths)
	if err != nil {
		return fmt.Errorf("cannot sync site pool: %w %v", err, string(char))
	}

	// Copy the control db to the backup location
	return nil
}
