package backup

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/utils"
)

// This method looks though all the ages of the DBs and removes user that are older than the max age from the circulation
// TODO we should have a way to unarchive a user
func (b *BackupJobs) ArchiveInactiveUsers() error {
	defer utils.TrackTime(time.Now(), "ArchiveInactiveUsers", nil, nil)

	ctx := context.TODO()
	users, err := b.controlDB.GetInactiveUsers(b.MaxUserInactive)
	if err != nil {
		return fmt.Errorf("Cant get inactive users %w", err)
	}
	if len(users) == 0 {
		return nil
	}

	removeList := []string{}

	for _, user := range users {
		if user.Subdomain == "" {
			continue
		}
		path, backupPath := utils.UserDBPaths(user.Subdomain)

		// One more backup
		err := utils.SqliteBackup(ctx, user.SQLiteDB, path, backupPath)
		if err != nil {
			return fmt.Errorf("cannot backup user db: %w", user.Subdomain, err)
		}

		// Save as backup file and upload to final localtion
		err = b.S3Client.UploadFilesToS3(b.S3Client.UserArchiveS3Path(user.Subdomain), []string{backupPath})
		if err != nil {
			return fmt.Errorf("cannot upload archived user db: %w", user.Subdomain, err)
		}

		removeList = append(removeList, user.Subdomain)
	}

	// Delete the user from the DB
	for _, subdomain := range removeList {
		err := b.controlDB.DeleteUser(subdomain)
		if err != nil {
			fmt.Printf("Failed to delete user %s: %v\n", subdomain, err)
		} else {
			fmt.Printf("Deleted user %s\n", subdomain)
		}
	}

	// Delete the Use DB WAL and BACKUP files
	for _, subdomain := range removeList {
		path, backupPath := utils.UserDBPaths(subdomain)

		err := utils.DeleteFiles([]string{path, backupPath})
		if err != nil {
			fmt.Printf("Failed to delete user db files for %s: %v\n", subdomain, err)
		} else {
			fmt.Printf("Deleted user db files for %s\n", subdomain)
		}

	}

	return nil
}
