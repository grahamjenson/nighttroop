package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"nighttroop.t1d.nz/clients/libre_client"
	"nighttroop.t1d.nz/models/libremodels"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

// libreClient is a site's LibreLinkUp client, kept between syncs so its login is reused
// until it expires, as LibreLinkUp limits logins. Its lock stops two syncs using it at once.
type libreClient struct {
	sync.Mutex
	email, password, region string
	*libre_client.Client
}

// libreClientFor returns the site's kept client, or a new one when there is none or its details have changed.
// There is at most one per site, so saving new details many times does not grow them, and at most
// maxNightscoutClients sites keep one.
func (b *ConnectorJobManager) libreClientFor(jobInfo *ConnectorJobInfo) (*libreClient, error) {
	b.libreClientsLock.Lock()
	defer b.libreClientsLock.Unlock()
	c, _ := b.libreClients.Get(jobInfo.SiteID)
	if c == nil || c.email != jobInfo.urlOrUsername || c.password != jobInfo.password || c.region != jobInfo.region {
		client, err := libre_client.NewClient(jobInfo.urlOrUsername, jobInfo.password, jobInfo.region)
		if err != nil {
			return nil, err
		}
		c = &libreClient{email: jobInfo.urlOrUsername, password: jobInfo.password, region: jobInfo.region, Client: client}
		b.libreClients.Add(jobInfo.SiteID, c)
	}
	return c, nil
}

// syncLibre imports the followed patient's readings, returning how many were new
func (b *ConnectorJobManager) syncLibre(jobInfo *ConnectorJobInfo) (int, error) {
	ctx, cancel := context.WithTimeout(jobInfo.ctx, syncTimeout)
	defer cancel()
	return b.copyLibreReadings(ctx, jobInfo)
}

// copyLibreReadings imports the followed patient's graph, the last 12 hours, all of it each time so a gap in syncing is
// filled, returning how many readings were new. Readings already imported have the same IDs, so they are skipped.
func (b *ConnectorJobManager) copyLibreReadings(ctx context.Context, jobInfo *ConnectorJobInfo) (int, error) {
	if jobInfo.region == "" {
		jobInfo.region = consts.LibreRegionGlobal
	}
	lc, err := b.libreClientFor(jobInfo)
	if err != nil {
		return 0, fmt.Errorf("Libre Error: %w", err)
	}
	lc.Lock()
	defer lc.Unlock()

	refused := false
	refuse := func(err error, record any) {
		if !refused {
			refused = true
			recordRefused(ctx, jobInfo, err, record)
		}
	}
	lc.Unreadable = func(raw json.RawMessage, err error) { refuse(fmt.Errorf("Libre: cannot read a reading: %w", err), raw) }
	readings, err := lc.FetchReadings(ctx)
	if err != nil {
		return 0, fmt.Errorf("Libre Error: %w", err)
	}
	added := 0
	for _, r := range readings {
		createParams := libremodels.ReadingToCreate(r, "libre_connector")
		row, err := jobInfo.Site.SiteDB.Queries().CreateGlucoseReading(ctx, createParams)
		if err == nil {
			added++
			jobInfo.Site.published(row)
		} else if !utilerrors.IsSQLDuplicateID(err) {
			refuse(fmt.Errorf("Libre: cannot keep a reading: %w", err), r)
		}
	}
	return added, nil
}
