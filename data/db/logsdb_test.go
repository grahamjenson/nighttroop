package db_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestLogsDBRecordsBySite(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs, site := t.Context(), controlDB.Logs, "somesite"

	logs.Log(ctx, &site, db.LogLevelInfo, "synced", map[string]any{"readings": 3})
	logs.Log(ctx, nil, db.LogLevelWarn, "server wide", nil)
	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.ConnectorError(ctx, site, consts.ConnectorTypeDexcom, errors.New("sync failed"))

	siteLogs, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Len(t, siteLogs, 1)
	require.Equal(t, "synced", siteLogs[0].Message)
	require.Equal(t, string(db.LogLevelInfo), siteLogs[0].Level)
	require.Equal(t, `{"readings":3}`, *siteLogs[0].Attributes)

	allLogs, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, allLogs, 2)
	require.Nil(t, allLogs[0].SiteID)

	siteEvents, err := logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, string(events.SiteDisabled), siteEvents[0].Name)
	require.Equal(t, `{"by":"admin"}`, *siteEvents[0].Attributes)
	// Disabling is over as soon as it happens, so it has a start but no finish
	require.NotZero(t, siteEvents[0].StartedAt)
	require.Nil(t, siteEvents[0].FinishedAt)

	siteErrors, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sync failed", siteErrors[0].Message)
	require.Equal(t, `{"kind":"connector","connector_type":"dexcom"}`, *siteErrors[0].Attributes)
}

// TestLogsDBKeepsWithinLimits checks a long message is cut short and long attributes are replaced, rather than
// the row being refused
func TestLogsDBKeepsWithinLimits(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs := t.Context(), controlDB.Logs

	logs.Log(ctx, nil, db.LogLevelError, strings.Repeat("é", 2000), map[string]any{"body": strings.Repeat("x", 5000)})

	rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.LessOrEqual(t, len(rows[0].Message), 1024)
	require.True(t, strings.HasSuffix(rows[0].Message, "é..."))
	require.Equal(t, `{"truncated":true}`, *rows[0].Attributes)
}

func TestLogsDBTrim(t *testing.T) {
	tests := []struct {
		name    string
		written int
		maxRows int64
		kept    int
	}{
		{"under the limit keeps all", 3, 5, 3},
		{"at the limit keeps all", 5, 5, 5},
		{"over the limit keeps the newest", 8, 5, 5},
		{"a limit of none keeps none", 3, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, controlDB := testutils.NewControlDB(t)
			ctx, logs := t.Context(), controlDB.Logs
			for i := range test.written {
				logs.Log(ctx, nil, db.LogLevelInfo, string(rune('a'+i)), nil)
				logs.SiteCreatedEvent(ctx, string(rune('a'+i)))
				logs.ControlBackupError(ctx, errors.New(string(rune('a'+i))))
			}

			require.NoError(t, logs.Trim(ctx, test.maxRows))

			rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 100})
			require.NoError(t, err)
			require.Len(t, rows, test.kept)
			if test.kept > 0 {
				require.Equal(t, string(rune('a'+test.written-1)), rows[0].Message, "the newest is kept")
			}
			events, err := logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 100})
			require.NoError(t, err)
			require.Len(t, events, test.kept)
			errs, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{Limit: 100})
			require.NoError(t, err)
			require.Len(t, errs, test.kept)
		})
	}
}

// TestLogsDBTrimAgain checks a trim after rows are written since the last keeps the newest, as the cron trims
func TestLogsDBTrimAgain(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs := t.Context(), controlDB.Logs
	for _, batch := range [][]string{{"a", "b", "c", "d"}, {"e", "f", "g"}} {
		for _, message := range batch {
			logs.Log(ctx, nil, db.LogLevelInfo, message, nil)
		}
		require.NoError(t, logs.Trim(ctx, 2))
	}

	rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "g", rows[0].Message)
	require.Equal(t, "f", rows[1].Message)
}

// BenchmarkLogsDBTrim is the trim the cron runs every minute on full tables, with a row each written since the last
func BenchmarkLogsDBTrim(b *testing.B) {
	const maxRows = 100_000
	_, controlDB := testutils.NewControlDB(b)
	ctx, logs := b.Context(), controlDB.Logs
	site, attributes := "benchsite", map[string]any{"connector_type": "dexcom", "body": strings.Repeat("x", 200)}
	write := func() {
		logs.Log(ctx, &site, db.LogLevelInfo, "synced", attributes)
		logs.ConnectorError(ctx, site, consts.ConnectorTypeDexcom, errors.New("sync failed"))
		logs.SiteDisabledEvent(ctx, site, "admin")
	}
	for range maxRows {
		write()
	}
	b.ResetTimer()
	for b.Loop() {
		write()
		require.NoError(b, logs.Trim(ctx, maxRows))
	}
}

func TestCreateSiteRecordsEvent(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	siteID := site.SiteDB.Subdomain

	events, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "site_created", events[0].Name)
}

// TestEventsArePublishedAndKept checks an event done to a site is sent through the server's events, to subscribers of
// that site, as well as kept
func TestEventsArePublishedAndKept(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	ctx, site := t.Context(), "publishedsite"
	var published []events.Event
	unsubscribe := cfg.Events.Subscribe(site, func(event events.Event) { published = append(published, event) })
	defer unsubscribe()

	controlDB.Logs.SiteDisabledEvent(ctx, site, "admin")
	require.Len(t, published, 1)
	require.Equal(t, events.SiteDisabled, published[0].Name)
	require.Equal(t, site, published[0].SiteID)
	require.Nil(t, published[0].Model, "it is not a model change")

	kept, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Len(t, kept, 1)
	require.Equal(t, string(events.SiteDisabled), kept[0].Name)
}

// TestSiteSeesOnlyItsOwnLogs checks a site sees its own logs, errors and events, and its batch's events, but nothing
// of another site, another batch or the server
func TestSiteSeesOnlyItsOwnLogs(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs, q := t.Context(), controlDB.Logs, controlDB.Logs.Queries
	site, other, batch, otherBatch := "qownsite", "qothersite", "q", "r"

	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.SiteDisabledEvent(ctx, other, "admin")
	logs.SiteBatchBackedUpEvent(ctx, batch, 2, time.Now(), false)
	logs.SiteBatchBackedUpEvent(ctx, otherBatch, 3, time.Now(), false)
	logs.ControlBackedUpEvent(ctx, time.Now(), false)
	logs.ConnectorError(ctx, other, consts.ConnectorTypeDexcom, errors.New("other failed"))
	logs.SiteBatchBackupError(ctx, batch, errors.New("batch failed"))
	logs.Log(ctx, &other, db.LogLevelInfo, "other synced", nil)

	siteEvents, err := q.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &site, Batch: &batch, Limit: 100})
	require.NoError(t, err)
	require.Len(t, siteEvents, 2)
	require.Equal(t, string(events.SiteBatchBackedUp), siteEvents[0].Name)
	require.Equal(t, batch, *siteEvents[0].Batch)
	require.Equal(t, site, *siteEvents[1].SiteID)

	// Another site's or batch's event, or the server's, cannot be read by its ID either
	allEvents, err := q.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 100})
	require.NoError(t, err)
	for _, event := range allEvents {
		_, err := q.GetSiteEvent(ctx, logsdb.GetSiteEventParams{ID: event.ID, SiteID: &site, Batch: &batch})
		if event.ID == siteEvents[0].ID || event.ID == siteEvents[1].ID {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, sql.ErrNoRows, event.Name)
		}
	}

	// A batch's errors can name its other sites, so a site sees only its own
	siteErrors, err := q.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Limit: 100})
	require.NoError(t, err)
	require.Empty(t, siteErrors)
	otherErrors, err := q.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &other, Limit: 100})
	require.NoError(t, err)
	_, err = q.GetSiteError(ctx, logsdb.GetSiteErrorParams{ID: otherErrors[0].ID, SiteID: &site})
	require.ErrorIs(t, err, sql.ErrNoRows)

	otherLogs, err := q.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &other, Limit: 100})
	require.NoError(t, err)
	_, err = q.GetSiteLog(ctx, logsdb.GetSiteLogParams{ID: otherLogs[0].ID, SiteID: &site})
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestLogsDBSearchByType checks a type keeps only the events of that name, or the errors of that kind, and no type
// keeps them all
func TestLogsDBSearchByType(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs, site, batch := t.Context(), controlDB.Logs, "typedsite", "typedbatch"

	logs.SiteCreatedEvent(ctx, site)
	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.ConnectorError(ctx, site, consts.ConnectorTypeDexcom, errors.New("sync failed"))
	logs.ServerError(ctx, &site, 500, "GET", "/", "request", "")

	str := func(s string) *string { return &s }
	tests := []struct {
		name       string
		typ        *string
		eventNames []string
		errorKinds []string
	}{
		{"every type", nil, []string{"site_disabled", "site_created"}, []string{"server", "connector"}},
		{"one event name", str("site_created"), []string{"site_created"}, nil},
		{"one error kind", str("connector"), nil, []string{"connector"}},
		{"no such type", str("nope"), nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &site, Type: tt.typ, Limit: 10})
			require.NoError(t, err)
			siteEvents, err := logs.Queries.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &site, Batch: &batch, Type: tt.typ, Limit: 10})
			require.NoError(t, err)
			errs, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Type: tt.typ, Limit: 10})
			require.NoError(t, err)

			names, siteNames, kinds := []string{}, []string{}, []string{}
			for _, e := range events {
				names = append(names, e.Name)
			}
			for _, e := range siteEvents {
				siteNames = append(siteNames, e.Name)
			}
			for _, e := range errs {
				kinds = append(kinds, strings.Split(strings.Split(*e.Attributes, `"kind":"`)[1], `"`)[0])
			}
			require.ElementsMatch(t, tt.eventNames, names)
			require.ElementsMatch(t, tt.eventNames, siteNames)
			require.ElementsMatch(t, tt.errorKinds, kinds)
		})
	}
}

// TestLogsDBSearchText checks a search keeps only rows with it in their text or attributes, ignoring case, and that
// a site cannot search the attributes of its batch's events, which it is not shown
func TestLogsDBSearchText(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs, site, batch := t.Context(), controlDB.Logs, "searchsite", "searchbatch"

	logs.Log(ctx, &site, db.LogLevelInfo, "Synced readings", map[string]any{"source": "dexcom"})
	logs.Log(ctx, &site, db.LogLevelInfo, "Something else", nil)
	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.SiteBatchBackedUpEvent(ctx, batch, 3, time.Now(), false)
	logs.ConnectorError(ctx, site, consts.ConnectorTypeNightscout, errors.New("Nightscout refused"))

	str := func(s string) *string { return &s }
	tests := []struct {
		name       string
		search     *string
		logs       int
		siteEvents int
		errors     int
	}{
		{"no search", nil, 2, 2, 1},
		{"message, ignoring case", str("synced READINGS"), 1, 0, 0},
		{"attributes", str("dexcom"), 1, 0, 0},
		{"event name", str("disabled"), 0, 1, 0},
		{"own event attributes", str("admin"), 0, 1, 0},
		{"batch event attributes are not searched", str("site_count"), 0, 0, 0},
		{"error message", str("refused"), 0, 0, 1},
		{"nothing matches", str("nope"), 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &site, Search: tt.search, Limit: 10})
			require.NoError(t, err)
			require.Len(t, found, tt.logs)
			events, err := logs.Queries.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &site, Batch: &batch, Search: tt.search, Limit: 10})
			require.NoError(t, err)
			require.Len(t, events, tt.siteEvents)
			errs, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Search: tt.search, Limit: 10})
			require.NoError(t, err)
			require.Len(t, errs, tt.errors)
		})
	}
}

// A refused record is kept on its connector error as JSON, or as the start of its text when it is too long or is not
// JSON, so it always fits
func TestRecordJSON(t *testing.T) {
	tests := []struct {
		name   string
		record any
		want   string
	}{
		{"a document is its JSON", map[string]any{"sgv": 120}, `{"sgv":120}`},
		{"a reading that did not decode is its JSON", json.RawMessage(`{"Value": "high"}`), `{"Value":"high"}`},
		{"JSON that does not parse is its text", json.RawMessage(`{"Value": `), `{"text":"{\"Value\": ","truncated":false}`},
		{"a long record is the start of its text", map[string]any{"notes": strings.Repeat("a", 9000)}, `"truncated":true`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(db.RecordJSON(test.record))
			require.NoError(t, err)
			require.Contains(t, string(raw), test.want)
			require.LessOrEqual(t, len(raw), 9*1024)
		})
	}
}
