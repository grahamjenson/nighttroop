package restore

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// RemoveStaging removes the folders a start or archive killed part way leaves beside the DBs: a download's
// `.<folder>-restore-*` staging folder (see s3utils.DownloadFilesFromS3) and an archive's `.archive-*` folder (see
// ControlDB.ArchiveSite). It is run before any restore or archive, so none of them is still in use.
func RemoveStaging(serverConfig *config.ServerConfig) {
	// The control DB downloads to its DB folder, and each batch to its folder; an archive is beside its backup
	restored := []string{serverConfig.DBDir()}
	patterns := []string{}
	for _, char := range utils.SITE_CHAR_SET {
		restored = append(restored, sitedb.GetBatchFolder(string(char), serverConfig))
		patterns = append(patterns, filepath.Join(filepath.Dir(sitedb.GetBackupPath(string(char), serverConfig)), ".archive-*"))
	}
	for _, outpath := range restored {
		patterns = append(patterns, filepath.Join(filepath.Dir(outpath), "."+filepath.Base(outpath)+"-restore-*"))
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			// Both are only ever made as folders, so anything else of the same name is left
			if info, err := os.Lstat(match); err != nil || !info.IsDir() {
				continue
			}
			if err := os.RemoveAll(match); err != nil {
				slog.Error("Cannot remove a leftover staging folder", "path", match, "error", err)
			} else {
				slog.Info("Removed a leftover staging folder", "path", match)
			}
		}
	}
}

func TryRestoreControlDB(s3Client *s3utils.S3Client) error {
	defer aslog.TrackTime(context.Background(), time.Now(), "TryRestoreControlDB", nil, nil)
	slog.Info("TryRestoreControlDB")

	err := TryDownloadExtractControlDbs(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download control db: %w", err)
	}

	return nil
}

func TryDownloadExtractControlDbs(s3Client *s3utils.S3Client) error {
	path := controldb.GetPath(s3Client.ServerConfig)
	backuppath := controldb.GetBackupPath(s3Client.ServerConfig)
	s3Path := controldb.GetS3Path(s3Client.ServerConfig)

	// If it already exists dont download
	if utils.FileExists(path) {
		return nil
	}

	err := utils.MkParentDirs(path)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	err = utils.MkParentDirs(backuppath)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	_, found, err := s3Client.DownloadFilesFromS3(context.Background(), s3Path, s3Client.ServerConfig.DBDir())
	if !found && err == nil {
		slog.Info("No control db backup found")
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}

func TryRestoreAllFiles(s3Client *s3utils.S3Client) error {
	defer aslog.TrackTime(context.Background(), time.Now(), "TryRestoreAllFiles", nil, nil)
	slog.Info("TryRestoreAllFiles")

	// A batch that failed is left without a folder, so the next start downloads it again
	if err := TryDownloadExrtractUserDbs(s3Client); err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	// Create the rest of the folders here
	for _, char := range utils.SITE_CHAR_SET {
		if err := utils.MkDirs(sitedb.GetBatchFolder(string(char), s3Client.ServerConfig)); err != nil {
			return err
		}
	}

	return nil
}

func TryDownloadExrtractUserDbs(s3Client *s3utils.S3Client) error {
	defer aslog.TrackTime(context.Background(), time.Now(), "TryDownloadExrtractUserDbs", nil, nil)

	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)

	slog.Info("TryDownloadExrtractUserDbs running", "site_batch_length", utils.SITE_BATCH_LENGTH, "limit", limit)
	for _, char := range utils.SITE_CHAR_SET {
		g.Go(func() error {
			_, err := TryDownloadExrtractUserDbsForBatch(s3Client, string(char))
			if err != nil {
				return fmt.Errorf("cannot restore site pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		slog.Error("One of the TryDownloadExrtractUserDbs tasks failed", "error", err)
		return fmt.Errorf("cannot restore and open all users: %w", err)
	} else {
		slog.Info("All TryDownloadExrtractUserDbs tasks succeeded!")
	}

	return nil
}

func TryDownloadExrtractUserDbsForBatch(s3Client *s3utils.S3Client, char string) (bool, error) {

	path := sitedb.GetBatchFolder(char, s3Client.ServerConfig)
	if utils.FileExists(path) {
		return true, nil
	}

	s3Path := sitedb.GetS3BatchPath(char, s3Client.ServerConfig)

	_, found, err := s3Client.DownloadFilesFromS3(context.Background(), s3Path, path)

	// A batch with no backup is not found, not an error
	if err != nil {
		return false, fmt.Errorf("cannot download user db: %w", err)
	}

	return found, nil
}
