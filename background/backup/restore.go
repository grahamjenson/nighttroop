package backup

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/s3utils"
)

func TryRestoreAllFiles(s3Client *s3utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "TryRestoreAllFiles", nil, nil)
	fmt.Println("TryRestoreAllFiles")

	err := TryDownloadExtractControlDbs(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	err = TryDownloadExrtractUserDbs(s3Client)

	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
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

	found, err := s3Client.DownloadFilesFromS3(s3Client.ControlS3Path(), utils.DB_FOLDER)
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

func TryDownloadExrtractUserDbs(s3Client *s3utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "TryDownloadExrtractUserDbs", nil, nil)

	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)

	fmt.Println("TryDownloadExrtractUserDbs running", utils.USER_BATCH_LENGTH, limit)
	for _, char := range utils.USER_CHAR_SET {
		g.Go(func() error {
			_, err := TryDownloadExrtractUserDbsForBatch(s3Client, string(char))
			if err != nil {
				return fmt.Errorf("cannot restore user pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the TryDownloadExrtractUserDbs  tasks failed: %v\n", err)
		return fmt.Errorf("cannot restore and open all users: %w", err)
	} else {
		fmt.Println("All TryDownloadExrtractUserDbs tasks succeeded!")
	}

	return nil
}

func TryDownloadExrtractUserDbsForBatch(s3Client *s3utils.S3Client, char string) (bool, error) {

	path, _ := utils.UserBatchPaths(char)
	s3Path := s3Client.UserBatchS3Path(char)

	found, err := s3Client.DownloadFilesFromS3(s3Path, path)

	if err != nil {
		// TODO dont fail here if just not found
		return false, fmt.Errorf("cannot download user db: %w", err)
	}

	return found, nil
}
