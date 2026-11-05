package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/clients/libre_client"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestBackfillStartsOneAtATimeAndRecordsWhatItDid(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeVirtual
	jobs := SetupSyncJobManager(cfg, controlDB)
	source := func() *controldb.SiteSyncSource {
		s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s
	}
	grabJobs := func() []*jobsdb.Job {
		queued, err := liteq.New(controlDB).GrabJobs(ctx, liteq.GrabJobsParams{Queue: SYNC_JOB, Count: 10})
		require.NoError(t, err)
		return queued
	}
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, Enabled: true}))

	// With no readings, a one day backfill fetches the day before now
	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 1)
	require.NoError(t, err)
	require.InDelta(t, time.Now().UnixMilli(), backfill.To, 1000)
	require.Equal(t, backfill.To-24*time.Hour.Milliseconds(), backfill.From)

	// It starts once: asking again while it is queued, as a double click does, starts nothing
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.ErrorIs(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill), ErrBackfillRunning)
	queued := grabJobs()
	require.Len(t, queued, 1)
	require.NotNil(t, source().BackfillStartedAt)
	require.Equal(t, int64(1), source().TotalBackfills)

	// Running it writes the day's virtual readings, and records how it went on the source and as an event
	require.NoError(t, jobs.workerFunc()(ctx, queued[0]))
	require.NoError(t, controlDB.Jobs.Queries.CompleteJob(ctx, jobsdb.CompleteJobParams{ID: queued[0].ID}))
	readings, err := site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 1000})
	require.NoError(t, err)
	require.Len(t, readings, 24*12)
	s := source()
	require.Nil(t, s.BackfillStartedAt)
	require.Nil(t, s.LastBackfillError)
	require.Equal(t, backfill.From, *s.LastBackfillFrom)
	require.Equal(t, backfill.To, *s.LastBackfillTo)
	require.NotNil(t, s.LastBackfillDuration)
	require.Greater(t, *s.LastBackfillRecords, int64(len(readings)), "readings, device statuses and treatments")
	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sync_backfilled", events[0].Name)
	require.NotNil(t, events[0].FinishedAt)
	require.Contains(t, *events[0].Attributes, `"sync_type":"virtual"`)

	// The next backfill starts again, from the oldest reading now in the site
	oldest := readings[len(readings)-1].OccurredAt
	next, err := BackfillRange(ctx, site.SiteDB, typ, 1)
	require.NoError(t, err)
	require.Equal(t, oldest, next.To)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, next))
	require.Len(t, grabJobs(), 1)
	require.Equal(t, int64(2), source().TotalBackfills)
}

func TestBackfillRangeOfDexcomIsTheLastDay(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	backfill, err := BackfillRange(t.Context(), site.SiteDB, consts.SyncSourceTypeDexcom, 7)
	require.NoError(t, err)
	require.InDelta(t, time.Now().UnixMilli(), backfill.To, 1000)
	require.Equal(t, backfill.To-24*time.Hour.Milliseconds(), backfill.From)
}

// LibreLinkUp's graph is the last 12 hours, so a Libre backfill fetches those, and runs as a backfill should
func TestLibreBackfillFetchesTheLastTwelveHours(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeLibre
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	region := "EU"
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(libre_client.TestValidEmail), PasswordEnc: enc("p"), Region: &region, Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.InDelta(t, time.Now().UnixMilli(), backfill.To, 1000)
	require.Equal(t, backfill.To-12*time.Hour.Milliseconds(), backfill.From)

	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))
	s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Nil(t, s.LastBackfillError, "a Libre backfill runs rather than saying Libre cannot be backfilled")
	require.Nil(t, s.BackfillStartedAt)
	require.Equal(t, int64(1), s.TotalBackfills)
}

func TestNightscoutBackfillCopiesOlderDocuments(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	threeDaysAgo := float64(time.Now().Add(-72 * time.Hour).UnixMilli())
	nightscout_client.TestDocuments = map[string][]any{
		"entries":    {map[string]any{"type": "sgv", "sgv": 140, "date": threeDaysAgo}},
		"treatments": {map[string]any{"eventType": "Correction Bolus", "insulin": 2, "date": threeDaysAgo}},
	}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))

	readings, err := site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	require.Equal(t, int64(threeDaysAgo), readings[0].OccurredAt)
	s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Equal(t, int64(2), *s.LastBackfillRecords)
}

func TestOldestDate(t *testing.T) {
	tests := []struct {
		name   string
		docs   []map[string]any
		oldest int64
	}{
		{"no documents", nil, 1000},
		{"the oldest of them", []map[string]any{{"date": 900.0}, {"date": 700.0}, {"date": 800.0}}, 700},
		{"none older", []map[string]any{{"date": 1000.0}}, 1000},
		{"without a date", []map[string]any{{"sgv": 100}}, 1000},
	}
	for _, test := range tests {
		require.Equal(t, test.oldest, oldestDate(test.docs, 1000), test.name)
	}
}

// A Nightscout server can return a document dated before the backfill's start, e.g. a device status whose date is
// not what it filters on. The backfill stops there, rather than asking for documents dated after its start and
// before that one, which none can be, and which a Nightscout server can search until the backfill times out.
func TestNightscoutBackfillStopsAtAPageOlderThanItsStart(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.SyncSourceTypeNightscout
	jobs := SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	// A full page of statuses from the last day, and one from a month ago; the fake server returns it for any dates
	statuses := []any{map[string]any{"device": "openaps://phone", "date": float64(time.Now().Add(-30 * 24 * time.Hour).UnixMilli())}}
	for i := 1; i < 50; i++ {
		statuses = append(statuses, map[string]any{"device": "openaps://phone", "date": float64(time.Now().Add(-time.Duration(i) * time.Minute).UnixMilli())})
	}
	nightscout_client.TestDocuments = map[string][]any{"devicestatus": statuses}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.SyncSource(ctx, SyncJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))

	s, err := q.GetSyncSource(ctx, controldb.GetSyncSourceParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Nil(t, s.LastBackfillError, "the backfill stops at the page older than its start")
}
