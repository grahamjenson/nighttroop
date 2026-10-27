package sync

import (
	"context"
	"fmt"
	"log/slog"
	gosync "sync"
	"time"

	"nighttroop.t1d.nz/clients/libre_client"
	"nighttroop.t1d.nz/models/libremodels"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

// libreClient is a site's LibreLinkUp client, kept between syncs so its login is reused
// until it expires, as LibreLinkUp limits logins. Its lock stops two syncs using it at once.
type libreClient struct {
	gosync.Mutex
	email, password, region string
	*libre_client.Client
}

// libreClientFor returns the site's kept client, or a new one when there is none or its details have changed.
// There is at most one per site, so saving new details many times does not grow them.
func (b *SyncJobManager) libreClientFor(syncInfo *SyncJobInfo) (*libreClient, error) {
	b.libreClientsLock.Lock()
	defer b.libreClientsLock.Unlock()
	c := b.libreClients[syncInfo.SiteID]
	if c == nil || c.email != syncInfo.urlOrUsername || c.password != syncInfo.password || c.region != syncInfo.region {
		client, err := libre_client.NewClient(syncInfo.urlOrUsername, syncInfo.password, syncInfo.region)
		if err != nil {
			return nil, err
		}
		c = &libreClient{email: syncInfo.urlOrUsername, password: syncInfo.password, region: syncInfo.region, Client: client}
		b.libreClients[syncInfo.SiteID] = c
	}
	return c, nil
}

func (b *SyncJobManager) syncLibre(syncInfo *SyncJobInfo) error {
	if len(syncInfo.UserSyncSources) == 0 {
		return fmt.Errorf("Libre, no user sync sources found for site_id %s type %s", syncInfo.SiteID, syncInfo.Type)
	}
	if syncInfo.region == "" {
		syncInfo.region = consts.LibreRegionGlobal
	}
	lc, err := b.libreClientFor(syncInfo)
	if err != nil {
		return fmt.Errorf("Libre Error: %w", err)
	}
	lc.Lock()
	defer lc.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	readings, err := lc.FetchReadings(ctx)
	if err != nil {
		return fmt.Errorf("Libre Error: %w", err)
	}
	// The graph is the last 12 hours, all of it imported each time so a gap in syncing is filled; readings already
	// imported have the same IDs, so they are skipped as duplicates
	for _, userSyncSource := range syncInfo.UserSyncSources {
		for _, r := range readings {
			createParams := libremodels.ReadingToCreate(r, "libre_sync")
			_, err := userSyncSource.SiteDB.Queries.CreateGlucoseReading(context.TODO(), createParams)
			if err != nil && !utilerrors.IsSQLDuplicateID(err) {
				slog.Error("Error creating entry in database", "error", err, "id", createParams.ID, "site_id", syncInfo.SiteID, "sync_type", syncInfo.Type)
			}
		}
	}
	return nil
}
