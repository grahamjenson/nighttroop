package backup

import (
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils"
)

func (b *BackupJobManager) ArchiveInactiveSites() error {
	defer utils.TrackTime(time.Now(), "ArchiveInactiveSites", nil, nil)

	deleteList, err := b.findSitesToDeleteWithoutArchiving()
	if err != nil {
		return fmt.Errorf("cant find sites to delete without archiving %w", err)
	}
	slog.Info("Delete site without archiving", "subdomains", deleteList)

	// Just delete these sites without an archive, as they have not been used in so many minutes
	err = b.controlDB.DeleteSites(deleteList)
	if err != nil {
		slog.Error("Failed to delete site without archiving", "error", err, "sites", deleteList)
	}

	archiveList, err := b.findSitesToArchive()
	if err != nil {
		return fmt.Errorf("cant find sites to archive %w", err)
	}
	archiveSubdomains := []string{}
	for _, site := range archiveList {
		archiveSubdomains = append(archiveSubdomains, site.Subdomain)
	}

	slog.Info("Delete site WITH archiving", "subdomains", archiveSubdomains)

	// Backup and upload the archived sites
	removeList, err := b.archiveSites(archiveList)
	if err != nil {
		return fmt.Errorf("cant archive sites %w", err)
	}

	// Delete the sites from the DB, and their site DB files (including WAL and BACKUP)
	err = b.controlDB.DeleteSites(removeList)
	if err != nil {
		slog.Error("Failed to delete site", "subdomain", removeList, "error", err)
	} else {
		if len(removeList) > 0 {
			slog.Info("Deleted site", "subdomain", removeList)
		}
	}

	return nil
}

// The sites we delete without archiving are sites that
func (b *BackupJobManager) findSitesToDeleteWithoutArchiving() ([]string, error) {
	sites, err := b.controlDB.GetNeverSeenSites(b.MaxSiteNeverSeen, b.MaxSiteNeverWritten)
	if err != nil {
		return nil, fmt.Errorf("cant get never seen sites %w", err)
	}

	if len(sites) == 0 {
		return nil, nil
	}

	deleteSites := []string{}
	for _, site := range sites {
		if site.Subdomain == "" {
			continue
		}
		deleteSites = append(deleteSites, site.Subdomain)
	}

	return deleteSites, nil

}

func (b *BackupJobManager) findSitesToArchive() ([]*db.SiteDB, error) {
	sites, err := b.controlDB.GetInactiveSites(b.MaxSiteLastSeen, b.MaxSiteNeverWritten)
	if err != nil {
		return nil, fmt.Errorf("cant get inactive sites %w", err)
	}
	if len(sites) == 0 {
		return nil, nil
	}
	return sites, nil
}

func (b *BackupJobManager) archiveSites(archiveList []*db.SiteDB) ([]string, error) {
	removeList := []string{}
	for _, site := range archiveList {
		if site.Subdomain == "" {
			continue
		}

		_, backupPath := utils.SiteDBPaths(site.Subdomain)

		// Last Backup
		err := utils.SqliteBackup(b.ctx, site.SQLiteDB, backupPath)
		if err != nil {
			return nil, fmt.Errorf("cannot backup site db %s: %w", site.Subdomain, err)
		}

		// Save as backup file and upload to final localtion
		err = b.S3Client.UploadFilesToS3(b.S3Client.SiteArchiveS3Path(site.Subdomain), []string{backupPath})
		if err != nil {
			return nil, fmt.Errorf("cannot upload archived site db %s: %w", site.Subdomain, err)
		}

		removeList = append(removeList, site.Subdomain)
	}
	return removeList, nil
}
