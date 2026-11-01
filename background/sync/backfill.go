package sync

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

// A backfill fetches a sync source's records from before the site's oldest glucose reading. A source has one queued
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
)

// BackfillParams is when, in ms, a backfill fetches records from and up to
type BackfillParams struct {
	From int64
	To   int64
}

// ErrBackfillRunning is a backfill asked for while the source has one queued or running
var ErrBackfillRunning = errors.New("a backfill is already running")

// BackfillRange is what a backfill of the source would fetch now: the backfillDays before the site's oldest glucose
// reading, or before now when it has none. Dexcom only has the last 24 hours and Libre the last 12, so their backfills
// fetch those.
func BackfillRange(ctx context.Context, siteDB *db.SiteDB, syncType consts.SyncSourceType, backfillDays int) (BackfillParams, error) {
	now := time.Now().UnixMilli()
	switch syncType {
	case consts.SyncSourceTypeDexcom:
		return BackfillParams{From: now - dexcomBackfillMinutes*time.Minute.Milliseconds(), To: now}, nil
	case consts.SyncSourceTypeLibre:
		return BackfillParams{From: now - libreBackfillMinutes*time.Minute.Milliseconds(), To: now}, nil
	}
	oldest, err := siteDB.Queries().OldestGlucoseReadingAt(ctx)
	if err != nil {
		return BackfillParams{}, err
	}
	if oldest == 0 {
		oldest = now
	}
	return BackfillParams{From: oldest - int64(backfillDays)*24*time.Hour.Milliseconds(), To: oldest}, nil
}

// QueueBackfill starts a backfill of the site's source and queues its job, or is ErrBackfillRunning when the source
// has one queued or running, so a double click starts only one
func QueueBackfill(ctx context.Context, controlDB *db.ControlDB, siteID string, syncType consts.SyncSourceType, backfill BackfillParams) error {
	result, err := controlDB.Queries.StartSyncSourceBackfill(ctx, controldb.StartSyncSourceBackfillParams{From: &backfill.From, To: &backfill.To, Type: syncType, SiteID: siteID})
	if err != nil {
		return err
	}
	if started, err := result.RowsAffected(); err != nil || started == 0 {
		return ErrBackfillRunning
	}
	job, err := json.Marshal(SyncJobParams{SiteID: siteID, Type: syncType, Backfill: &backfill})
	if err == nil {
		err = liteq.New(controlDB).QueueJob(ctx, liteq.QueueJobParams{Queue: SYNC_JOB, Job: string(job), DedupingKey: liteq.IgnoreDuplicate(fmt.Sprintf("%s-%s-backfill", siteID, syncType))})
	}
	if err != nil {
		// Not queued, so it is finished rather than blocking the next for an hour
		message := utilerrors.Truncate(err.Error())
		_ = controlDB.Queries.FinishSyncSourceBackfill(ctx, controldb.FinishSyncSourceBackfillParams{Error: &message, Type: syncType, SiteID: siteID})
	}
	return err
}

// backfill runs a queued backfill, recording how long it took, how many records it added and why it stopped early,
// if it did, on the source and as an event
func (b *SyncJobManager) backfill(syncInfo *SyncJobInfo, backfill BackfillParams) error {
	started := time.Now()
	ctx, cancel := context.WithTimeout(syncInfo.ctx, backfillTimeout)
	defer cancel()

	from, to := time.UnixMilli(backfill.From), time.UnixMilli(backfill.To)
	records, err := 0, fmt.Errorf("%s cannot be backfilled", syncInfo.Type)
	switch syncInfo.Type {
	case consts.SyncSourceTypeNightscout:
		records, err = b.backfillNightscout(ctx, syncInfo, from, to)
	case consts.SyncSourceTypeDexcom:
		records, err = copyDexcomEGVs(ctx, syncInfo, dexcomBackfillMinutes)
	case consts.SyncSourceTypeLibre:
		records, err = b.copyLibreReadings(ctx, syncInfo)
	case consts.SyncSourceTypeVirtual:
		records, err = b.backfillVirtual(ctx, syncInfo, from, to)
	}

	// Recorded with the sync's own context, as a backfill that ran out of time has none left
	recordCtx := syncInfo.ctx
	var message *string
	if err != nil {
		truncated := utilerrors.Truncate(err.Error())
		message = &truncated
		b.controlDB.Logs.SyncError(recordCtx, syncInfo.SiteID, syncInfo.Type, err)
	}
	b.controlDB.Logs.SyncBackfilledEvent(recordCtx, syncInfo.SiteID, syncInfo.Type, started, backfill.From, backfill.To, records, err)
	durationMs, added := time.Since(started).Milliseconds(), int64(records)
	return b.controlDB.Queries.FinishSyncSourceBackfill(recordCtx, controldb.FinishSyncSourceBackfillParams{DurationMs: &durationMs, Records: &added, Error: message, Type: syncInfo.Type, SiteID: syncInfo.SiteID})
}

// backfillNightscout copies each collection's documents dated from from to to, newest first a page at a time
func (b *SyncJobManager) backfillNightscout(ctx context.Context, syncInfo *SyncJobInfo, from, to time.Time) (int, error) {
	nsc := b.nightscoutClientFor(syncInfo)
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
			created, err := b.createNightscoutDocuments(ctx, syncInfo, c.collection, docs)
			added += created
			if err != nil {
				return added, err
			}
			if len(docs) < c.backfillPage {
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
func (b *SyncJobManager) backfillVirtual(ctx context.Context, syncInfo *SyncJobInfo, from, to time.Time) (int, error) {
	return WriteVirtualData(ctx, b.nightscoutStore, syncInfo.Site.siteContext(virtualIdentity), to, to.Sub(from))
}
