package restore

import (
	"log/slog"
	"os"
	"path/filepath"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// RemoveStaging removes the folders a start or archive killed part way leaves beside the DBs: a download's
// `.<folder>-restore-*` staging folder (see s3utils.DownloadFilesFromS3) and an archive's `.archive-*` folder (see
// ControlDB.ArchiveSite). It is run before any restore or archive, so none of them is still in use.
func RemoveStaging(serverConfig *config.ServerConfig) {
	// Each site downloads to its letter's folder; an archive is beside its backup
	restored := []string{}
	patterns := []string{}
	for _, char := range utils.SITE_CHAR_SET {
		restored = append(restored, sitedb.GetSitesFolder(string(char), serverConfig))
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
