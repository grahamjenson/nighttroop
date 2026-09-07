package backup_control

import (
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/s3utils"
)

func TryRestoreControlDB(s3Client *s3utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "TryRestoreControlDB", nil, nil)
	slog.Info("TryRestoreControlDB")

	err := TryDownloadExtractControlDbs(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download control db: %w", err)
	}

	return nil
}

func TryDownloadExtractControlDbs(s3Client *s3utils.S3Client) error {
	// If it already exists dont download
	if utils.FileExists(utils.CONTROL_DB_PATH) {
		return nil
	}

	err := utils.MkParentDirs(utils.CONTROL_DB_PATH)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	err = utils.MkParentDirs(utils.CONTROL_DB_BACKUP_PATH)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	_, found, err := s3Client.DownloadFilesFromS3(s3Client.ControlS3Path(), utils.DB_FOLDER)
	if !found && err == nil {
		slog.Info("No control db backup found")
		return nil
	}
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}
