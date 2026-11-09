package connectors

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
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeVirtual
	jobs := SetupConnectorJobManager(cfg, controlDB)
	source := func() *controldb.SiteConnector {
		s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s
	}
	grabJobs := func() []*jobsdb.Job {
		queued, err := liteq.New(controlDB).GrabJobs(ctx, liteq.GrabJobsParams{Queue: CONNECTOR_JOB, Count: 10})
		require.NoError(t, err)
		return queued
	}
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))

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
	require.Equal(t, *s.LastBackfillRecords, s.TotalRecords, "its records are added to the source's total")
	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "connector_backfilled", events[0].Name)
	require.NotNil(t, events[0].FinishedAt)
	require.Contains(t, *events[0].Attributes, `"connector_type":"virtual"`)

	// The next backfill starts again, from the day before the oldest reading now in the site, which is within the
	// recent days it fetches again, so it fetches up to now in one range
	oldest := readings[len(readings)-1].OccurredAt
	next, err := BackfillRange(ctx, site.SiteDB, typ, 1)
	require.NoError(t, err)
	require.Equal(t, oldest-24*time.Hour.Milliseconds(), next.From)
	require.InDelta(t, time.Now().UnixMilli(), next.To, 1000)
	require.Zero(t, next.RecentTo)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, next))
	require.Len(t, grabJobs(), 1)
	require.Equal(t, int64(2), source().TotalBackfills)
}

// A site down for a while, as for a deploy, misses its connector's records, so a backfill also fetches the recent days
// again, as well as the days before its oldest reading, leaving the days between them as they are
func TestBackfillFillsGapsInTheRecentDays(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeVirtual
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, Enabled: true}))
	day, now := 24*time.Hour.Milliseconds(), time.Now().UnixMilli()
	oldest := now - 5*day
	_, err := site.SiteDB.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "old", Mgdl: 120, ReadingType: "sgv", OccurredAt: oldest, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 1)
	require.NoError(t, err)
	require.Equal(t, BackfillParams{From: oldest - day, To: oldest, RecentFrom: backfill.RecentTo - recentBackfillDays*day, RecentTo: backfill.RecentTo}, backfill)
	require.InDelta(t, now, backfill.RecentTo, 1000)

	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, SetupConnectorJobManager(cfg, controlDB).RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))
	count := func(from, to int64) int {
		readings, err := site.SiteDB.Queries().SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: &from, DateLT: &to, Limit: 1000})
		require.NoError(t, err)
		return len(readings)
	}
	require.Equal(t, 24*12, count(oldest-day, oldest), "the day before the oldest reading")
	require.Zero(t, count(oldest+1, backfill.RecentFrom), "the days between are left as they are")
	require.InDelta(t, recentBackfillDays*24*12, count(backfill.RecentFrom, backfill.RecentTo+1), 1, "the recent days")
}

func TestBackfillRangeOfDexcomIsTheLastDay(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	backfill, err := BackfillRange(t.Context(), site.SiteDB, consts.ConnectorTypeDexcom, 7)
	require.NoError(t, err)
	require.InDelta(t, time.Now().UnixMilli(), backfill.To, 1000)
	require.Equal(t, backfill.To-24*time.Hour.Milliseconds(), backfill.From)
}

// LibreLinkUp's graph is the last 12 hours, so a Libre backfill fetches those, and runs as a backfill should
func TestLibreBackfillFetchesTheLastTwelveHours(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeLibre
	jobs := SetupConnectorJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	region := "EU"
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(libre_client.TestValidEmail), PasswordEnc: enc("p"), Region: &region, Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.InDelta(t, time.Now().UnixMilli(), backfill.To, 1000)
	require.Equal(t, backfill.To-12*time.Hour.Milliseconds(), backfill.From)

	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))
	s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Nil(t, s.LastBackfillError, "a Libre backfill runs rather than saying Libre cannot be backfilled")
	require.Nil(t, s.BackfillStartedAt)
	require.Equal(t, int64(1), s.TotalBackfills)
}

func TestNightscoutBackfillCopiesOlderDocuments(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	jobs := SetupConnectorJobManager(cfg, controlDB)
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
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))

	readings, err := site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	require.Equal(t, int64(threeDaysAgo), readings[0].OccurredAt)
	s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
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
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	jobs := SetupConnectorJobManager(cfg, controlDB)
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
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	backfill, err := BackfillRange(ctx, site.SiteDB, typ, 7)
	require.NoError(t, err)
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))

	s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
	require.NoError(t, err)
	require.Nil(t, s.LastBackfillError, "the backfill stops at the page older than its start")
}

// A Nightscout backfill copies old device statuses with their IOB, COB and pump's battery, which summaries show, but
// none of their loop status or pump as sent, which are of no use once old; a status the site already had keeps its own
func TestNightscoutBackfillKeepsNoLoopStatus(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	threeDaysAgo, twoDaysAgo := time.Now().Add(-72*time.Hour).UnixMilli(), time.Now().Add(-48*time.Hour).UnixMilli()
	openAPS := map[string]any{"iob": map[string]any{"iob": 1.5, "basaliob": 0.5}, "suggested": map[string]any{"COB": 20, "reason": "COB 20", "predBGs": map[string]any{"IOB": []any{120, 118}}}}
	nightscout_client.TestDocuments = map[string][]any{"devicestatus": {
		map[string]any{"device": "openaps://phone", "date": float64(threeDaysAgo), "openaps": openAPS, "pump": map[string]any{"battery": map[string]any{"percent": 60}}},
		map[string]any{"device": "openaps://phone", "date": float64(twoDaysAgo), "openaps": openAPS},
	}}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })
	// The site already has the later one, as a sync copied it, with its loop status
	kept := `{"iob":{"iob":1.5}}`
	_, err := site.SiteDB.Queries().CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "synced", OccurredAt: twoDaysAgo, CreatedByDeviceName: new("openaps://phone"),
		LoopName: new("openaps"), LoopStatusJson: &kept, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	now := time.Now()
	backfill := BackfillParams{From: now.Add(-7 * 24 * time.Hour).UnixMilli(), To: now.Add(-time.Hour).UnixMilli()}
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	require.NoError(t, SetupConnectorJobManager(cfg, controlDB).RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))

	reports, err := site.SiteDB.Queries().AllDeviceReports(ctx, sitedb.AllDeviceReportsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, reports, 2)
	byDate := map[int64]*sitedb.DeviceReport{}
	for _, report := range reports {
		byDate[report.OccurredAt] = report
	}
	backfilled := byDate[threeDaysAgo]
	require.Nil(t, backfilled.LoopStatusJson, "a backfilled status keeps no loop status")
	require.Equal(t, 1.5, *backfilled.LoopIob, "its IOB is kept")
	require.Equal(t, 20.0, *backfilled.LoopCob, "and its COB")
	require.Nil(t, backfilled.PumpJson, "it keeps no pump as sent")
	require.Equal(t, 60.0, *backfilled.PumpBatteryPercent, "but keeps its pump's battery")
	require.Equal(t, kept, *byDate[twoDaysAgo].LoopStatusJson, "a status the site had keeps its own")
}

// A backfill copies no device status older than 30 days, from a Nightscout, though its server returns one dated
// before the range, or from the virtual source; it copies a newer one without its loop status or pump as sent
func TestBackfillCopiesNoDeviceStatusOlderThan30Days(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	daysAgo := func(days int) int64 { return time.Now().Add(-time.Duration(days) * 24 * time.Hour).UnixMilli() }
	status := func(date int64) map[string]any {
		return map[string]any{"device": "openaps://phone", "date": float64(date), "openaps": map[string]any{"iob": map[string]any{"iob": 1.5}},
			"pump": map[string]any{"battery": map[string]any{"percent": 60}}}
	}
	nightscout_client.TestDocuments = map[string][]any{
		"entries":      {map[string]any{"type": "sgv", "sgv": 140, "date": float64(daysAgo(40))}},
		"devicestatus": {status(daysAgo(3)), status(daysAgo(40))},
	}
	t.Cleanup(func() { nightscout_client.TestDocuments = map[string][]any{} })
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	backfill := BackfillParams{From: daysAgo(60), To: daysAgo(1)}
	require.NoError(t, QueueBackfill(ctx, controlDB, siteID, typ, backfill))
	jobs := SetupConnectorJobManager(cfg, controlDB)
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Backfill: &backfill}))
	reports, err := site.SiteDB.Queries().AllDeviceReports(ctx, sitedb.AllDeviceReportsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, reports, 1, "only the device status from the last 30 days is copied")
	require.Equal(t, daysAgo(3)/1000, reports[0].OccurredAt/1000)
	require.Nil(t, reports[0].LoopStatusJson, "without its loop status")
	require.Nil(t, reports[0].PumpJson, "or its pump as sent")
	require.Equal(t, 1.5, *reports[0].LoopIob, "but with its IOB")
	readings, err := site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1, "an older reading is still copied")

	// The virtual source's 10 minutes 31 days ago are its readings, without its device statuses
	connectorSite, err := jobs.connectorSite(ctx, siteID)
	require.NoError(t, err)
	_, err = WriteVirtualData(ctx, jobs.nightscoutStore, connectorSite.siteContext(virtualIdentity), time.UnixMilli(daysAgo(31)), 10*time.Minute)
	require.NoError(t, err)
	reports, err = site.SiteDB.Queries().AllDeviceReports(ctx, sitedb.AllDeviceReportsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, reports, 1)
	readings, err = site.SiteDB.Queries().AllGlucoseReadings(ctx, sitedb.AllGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Greater(t, len(readings), 1)
}
