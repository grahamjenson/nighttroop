package backup

import (
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/utils"
)

func (b *BackupJobManager) DeleteUnusedSites() error {
	defer utils.TrackTime(time.Now(), "DeleteUnusedSites", nil, nil)

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

	return nil
}

// The sites we delete without archiving are sites that
func (b *BackupJobManager) findSitesToDeleteWithoutArchiving() ([]string, error) {
	sites, err := GetNeverUsedSites(b.controlDB, b.MaxSiteNeverSeen, b.MaxSiteNeverWritten)
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
