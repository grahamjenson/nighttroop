package backup

import (
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
)

func (b *BackupJobs) ArchiveInactiveUsers() error {
	defer utils.TrackTime(time.Now(), "ArchiveInactiveUsers", nil, nil)

	deleteList, err := b.findUsersToDeleteWithoutArchiving()
	if err != nil {
		return fmt.Errorf("Cant find users to delete without archiving %w", err)
	}

	// Just delete these users without an archive, as they have not been used in so many minutes
	err = b.controlDB.DeleteUsers(deleteList)
	if err != nil {
		slog.Error("Failed to delete user without archiving", "error", err, "users", deleteList)
	}

	archiveList, err := b.findUsersToArchive()
	if err != nil {
		return fmt.Errorf("Cant find users to archive %w", err)
	}

	// Backup and upload the archived users
	removeList, err := b.archiveUsers(archiveList)
	if err != nil {
		return fmt.Errorf("Cant archive users %w", err)
	}

	// Delete the users from the DB, and their user DB files (including WAL and BACKUP)
	err = b.controlDB.DeleteUsers(removeList)
	if err != nil {
		slog.Error("Failed to delete user", "subdomain", removeList, "error", err)
	} else {
		if len(removeList) > 0 {
			slog.Info("Deleted user", "subdomain", removeList)
		}
	}

	return nil
}

// The users we delete without archiving are users that
func (b *BackupJobs) findUsersToDeleteWithoutArchiving() ([]string, error) {
	users, err := b.controlDB.GetNeverSeenUsers(b.MaxUserNeverSeen, b.MaxUserNeverWritten)
	if err != nil {
		return nil, fmt.Errorf("Cant get never seen users %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	deleteUsers := []string{}
	for _, user := range users {
		if user.Subdomain == "" {
			continue
		}
		deleteUsers = append(deleteUsers, user.Subdomain)
	}

	return deleteUsers, nil

}

func (b *BackupJobs) findUsersToArchive() ([]*db.UserDB, error) {
	users, err := b.controlDB.GetInactiveUsers(b.MaxUserLastSeen, b.MaxUserNeverWritten)
	if err != nil {
		return nil, fmt.Errorf("Cant get inactive users %w", err)
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users, nil
}

func (b *BackupJobs) archiveUsers(archiveList []*db.UserDB) ([]string, error) {
	removeList := []string{}
	for _, user := range archiveList {
		if user.Subdomain == "" {
			continue
		}
		path, backupPath := utils.UserDBPaths(user.Subdomain)

		// Last Backup
		err := utils.SqliteBackup(b.ctx, user.SQLiteDB, path, backupPath)
		if err != nil {
			return nil, fmt.Errorf("cannot backup user db: %w", user.Subdomain, err)
		}

		// Save as backup file and upload to final localtion
		err = b.S3Client.UploadFilesToS3(b.S3Client.UserArchiveS3Path(user.Subdomain), []string{backupPath})
		if err != nil {
			return nil, fmt.Errorf("cannot upload archived user db: %w", user.Subdomain, err)
		}

		removeList = append(removeList, user.Subdomain)
	}
	return removeList, nil
}
