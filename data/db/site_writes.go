package db

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
)

// RecordModelWrites records each site written by a model change on bus, such as a connector's sync or a live update,
// which no request records, as written to, at most every 10 minutes, as its requests record their writes. The site
// backup job backs up the sites written to.
func (db *ControlDB) RecordModelWrites(bus *events.SiteBus) {
	var mu sync.Mutex
	recorded := map[string]int64{}
	bus.SubscribeAll(func(event events.Event) {
		now := time.Now().UnixMilli()
		mu.Lock()
		due := event.Action != "" && now > recorded[event.SiteID]+(10*time.Minute).Milliseconds()
		if due {
			recorded[event.SiteID] = now
		}
		mu.Unlock()
		// Publishing waits for each handler, so the write is not waited for
		if due {
			go func() {
				if err := db.Queries.UpdateSiteDateFlags(context.Background(), controldb.UpdateSiteDateFlagsParams{LastWriteAt: &now, ID: event.SiteID}); err != nil {
					slog.Error("Cannot record a site's write", "site_id", event.SiteID, "error", err)
				}
			}()
		}
	})
}
