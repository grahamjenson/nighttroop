package connectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

// A backfill fetches a connector's records from before the site's oldest glucose reading. A source has one queued
// or running at a time, and each is recorded on the source and as an event, so one that backfills too much can be
// slowed or blocked.

const (
	// backfillTimeout stays under the job queue's 60 second visibility timeout, so a backfill never runs twice at once
	backfillTimeout = 50 * time.Second
	// maxBackfillPages caps the Nightscout requests one backfill makes
	maxBackfillPages = 200
	// dexcomBackfillMinutes is as far back as Dexcom Share has readings
	dexcomBackfillMinutes = 24 * 60
	// libreBackfillMinutes is as far back as the LibreLinkUp graph goes
	libreBackfillMinutes = 12 * 60
	// recentBackfillDays are the last days a backfill fetches again, to fill gaps the site missed, as while it was down
	// for a deploy
	recentBackfillDays = 2
)

// BackfillParams is when, in ms, a backfill fetches records from and up to, and the recent days it fetches again,
// when they are not in that range already
type BackfillParams struct {
	From       int64
	To         int64
	RecentFrom int64 `json:",omitempty"`
	RecentTo   int64 `json:",omitempty"`
}

// ranges are the backfill's times, from and to, and its recent days when it has them
func (p BackfillParams) ranges() [][2]time.Time {
	ranges := [][2]time.Time{{time.UnixMilli(p.From), time.UnixMilli(p.To)}}
	if p.RecentTo > 0 {
		ranges = append(ranges, [2]time.Time{time.UnixMilli(p.RecentFrom), time.UnixMilli(p.RecentTo)})
	}
	return ranges
}

// eachRange fetches each of the backfill's ranges, adding up the records it added, until one fails
func eachRange(backfill BackfillParams, fetch func(from, to time.Time) (int, error)) (int, error) {
	added := 0
	for _, r := range backfill.ranges() {
		records, err := fetch(r[0], r[1])
		if added += records; err != nil {
			return added, err
		}
	}
	return added, nil
}

// ErrBackfillRunning is a backfill asked for while the source has one queued or running
var ErrBackfillRunning = errors.New("a backfill is already running")

// BackfillRange is what a backfill of the source would fetch now: the backfillDays before the site's oldest glucose
// reading, or before now when it has none, and the recentBackfillDays up to now again, to fill gaps, joined into one
// range when they meet. Dexcom only has the last 24 hours and Libre the last 12, so their backfills fetch those.
func BackfillRange(ctx context.Context, siteDB *db.SiteDB, connectorType consts.ConnectorType, backfillDays int) (BackfillParams, error) {
	now := time.Now().UnixMilli()
	switch connectorType {
	case consts.ConnectorTypeDexcom:
		return BackfillParams{From: now - dexcomBackfillMinutes*time.Minute.Milliseconds(), To: now}, nil
	case consts.ConnectorTypeLibre:
		return BackfillParams{From: now - libreBackfillMinutes*time.Minute.Milliseconds(), To: now}, nil
	}
	oldest, err := siteDB.Queries().OldestGlucoseReadingAt(ctx)
	if err != nil {
		return BackfillParams{}, err
	}
	if oldest == 0 {
		oldest = now
	}
	backfill := BackfillParams{From: oldest - int64(backfillDays)*24*time.Hour.Milliseconds(), To: oldest}
	if recent := now - recentBackfillDays*24*time.Hour.Milliseconds(); recent <= oldest {
		backfill.To = now
	} else {
		backfill.RecentFrom, backfill.RecentTo = recent, now
	}
	return backfill, nil
}

// QueueBackfill starts a backfill of the site's source and queues its job, or is ErrBackfillRunning when the source
// has one queued or running, so a double click starts only one
func QueueBackfill(ctx context.Context, controlDB *db.ControlDB, siteID string, connectorType consts.ConnectorType, backfill BackfillParams) error {
	result, err := controlDB.Queries.StartConnectorBackfill(ctx, controldb.StartConnectorBackfillParams{From: &backfill.From, To: &backfill.To, Type: connectorType, SiteID: siteID})
	if err != nil {
		return err
	}
	if started, err := result.RowsAffected(); err != nil || started == 0 {
		return ErrBackfillRunning
	}
	job, err := json.Marshal(ConnectorJobParams{SiteID: siteID, Type: connectorType, Backfill: &backfill})
	if err == nil {
		err = liteq.New(controlDB).QueueJob(ctx, liteq.QueueJobParams{Queue: CONNECTOR_JOB, Job: string(job), DedupingKey: liteq.IgnoreDuplicate(fmt.Sprintf("%s-%s-backfill", siteID, connectorType))})
	}
	if err != nil {
		// Not queued, so it is finished rather than blocking the next for an hour
		message := utilerrors.Truncate(err.Error())
		_ = controlDB.Queries.FinishConnectorBackfill(ctx, controldb.FinishConnectorBackfillParams{Error: &message, Type: connectorType, SiteID: siteID})
	}
	return err
}

// backfill runs a queued backfill, recording how long it took, how many records it added and why it stopped early,
// if it did, on the source and as an event
func (b *ConnectorJobManager) backfill(jobInfo *ConnectorJobInfo, backfill BackfillParams) error {
	started := time.Now()
	ctx, cancel := context.WithTimeout(jobInfo.ctx, backfillTimeout)
	defer cancel()

	records, err := 0, fmt.Errorf("%s cannot be backfilled", jobInfo.Type)
	switch jobInfo.Type {
	case consts.ConnectorTypeNightscout:
		records, err = eachRange(backfill, func(from, to time.Time) (int, error) { return b.backfillNightscout(ctx, jobInfo, from, to) })
	case consts.ConnectorTypeDexcom:
		records, err = copyDexcomEGVs(ctx, jobInfo, dexcomBackfillMinutes)
	case consts.ConnectorTypeLibre:
		records, err = b.copyLibreReadings(ctx, jobInfo)
	case consts.ConnectorTypeVirtual:
		records, err = eachRange(backfill, func(from, to time.Time) (int, error) { return b.backfillVirtual(ctx, jobInfo, from, to) })
	}

	// Recorded with the sync's own context, as a backfill that ran out of time has none left
	recordCtx := jobInfo.ctx
	var message *string
	if err != nil {
		truncated := utilerrors.Truncate(err.Error())
		message = &truncated
		b.controlDB.Logs.ConnectorError(recordCtx, jobInfo.SiteID, jobInfo.Type, err)
	}
	b.controlDB.Logs.ConnectorBackfilledEvent(recordCtx, jobInfo.SiteID, jobInfo.Type, started, backfill.From, backfill.To, records, err)
	durationMs, added := time.Since(started).Milliseconds(), int64(records)
	err = b.controlDB.Queries.FinishConnectorBackfill(recordCtx, controldb.FinishConnectorBackfillParams{DurationMs: &durationMs, Records: &added, Error: message, Type: jobInfo.Type, SiteID: jobInfo.SiteID})
	b.publishConnector(jobInfo.SiteID, jobInfo.Type)
	return err
}

// backfillNightscout copies each collection's documents dated from from to to, newest first a page at a time
func (b *ConnectorJobManager) backfillNightscout(ctx context.Context, jobInfo *ConnectorJobInfo, from, to time.Time) (int, error) {
	nsc := b.nightscoutClientFor(jobInfo)
	nsc.Lock()
	defer nsc.Unlock()

	added, pages := 0, 0
	for _, c := range nightscoutCollections {
		before := to.UnixMilli()
		for {
			if pages++; pages > maxBackfillPages {
				return added, fmt.Errorf("Nightscout backfill stopped after %d pages", maxBackfillPages)
			}
			docs, err := nsc.GetDocuments(ctx, string(c.collection), from, time.UnixMilli(before), c.backfillPage)
			if err != nil {
				return added, fmt.Errorf("NightscoutBackfill Error: %w", err)
			}
			oldest := oldestDate(docs, before)
			created, err := b.createNightscoutDocuments(ctx, jobInfo, c.collection, docs, c.backfillSave)
			added += created
			if err != nil {
				return added, err
			}
			// A page reaching back before from has all the range, and the next would ask for documents dated after from
			// and before its oldest, which none are and which a Nightscout server can search until the backfill times out
			if len(docs) < c.backfillPage || oldest < from.UnixMilli() {
				break
			}
			// The next page includes the oldest date, so documents sharing it are not missed, unless the whole page had it
			before = min(oldest, before-1)
		}
	}
	return added, nil
}

// oldestDate is the oldest date of the documents, or before when none is older
func oldestDate(docs []map[string]any, before int64) int64 {
	oldest := before
	for _, doc := range docs {
		if date, ok := doc["date"].(float64); ok {
			oldest = min(oldest, int64(date))
		}
	}
	return oldest
}

// backfillVirtual writes the virtual source's documents from from to to into the site
func (b *ConnectorJobManager) backfillVirtual(ctx context.Context, jobInfo *ConnectorJobInfo, from, to time.Time) (int, error) {
	return WriteVirtualData(ctx, b.nightscoutStore, jobInfo.Site.siteContext(virtualIdentity), to, to.Sub(from))
}
