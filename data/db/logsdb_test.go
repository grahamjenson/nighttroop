package db_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestLogsDBRecordsBySite(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs, site := context.TODO(), controlDB.Logs, "somesite"

	logs.Log(ctx, &site, db.LogLevelInfo, "synced", map[string]any{"readings": 3})
	logs.Log(ctx, nil, db.LogLevelWarn, "server wide", nil)
	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.SyncError(ctx, site, consts.SyncSourceTypeDexcom, errors.New("sync failed"))

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

	events, err := logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, string(db.EventSiteDisabled), events[0].Name)
	require.Equal(t, `{"by":"admin"}`, *events[0].Attributes)
	// Disabling is over as soon as it happens, so it has a start but no finish
	require.NotZero(t, events[0].StartedAt)
	require.Nil(t, events[0].FinishedAt)

	siteErrors, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sync failed", siteErrors[0].Message)
	require.Equal(t, `{"kind":"sync","sync_type":"dexcom"}`, *siteErrors[0].Attributes)
}

// TestLogsDBKeepsWithinLimits checks a long message is cut short and long attributes are replaced, rather than
// the row being refused
func TestLogsDBKeepsWithinLimits(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs := context.TODO(), controlDB.Logs

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
			_, controlDB, err := testutils.GetConfigControlDBForTest()
			require.NoError(t, err)
			ctx, logs := context.TODO(), controlDB.Logs
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

func TestCreateSiteRecordsEvent(t *testing.T) {
	_, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain

	events, err := controlDB.Logs.Queries.SearchEvents(context.TODO(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "site_created", events[0].Name)
}

// TestSiteSeesOnlyItsOwnLogs checks a site sees its own logs, errors and events, and its batch's events, but nothing
// of another site, another batch or the server
func TestSiteSeesOnlyItsOwnLogs(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs, q := context.TODO(), controlDB.Logs, controlDB.Logs.Queries
	site, other, batch, otherBatch := "qownsite", "qothersite", "q", "r"

	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.SiteDisabledEvent(ctx, other, "admin")
	logs.SiteBatchBackedUpEvent(ctx, batch, 2, time.Now(), false)
	logs.SiteBatchBackedUpEvent(ctx, otherBatch, 3, time.Now(), false)
	logs.ControlBackedUpEvent(ctx, time.Now(), false)
	logs.SyncError(ctx, other, consts.SyncSourceTypeDexcom, errors.New("other failed"))
	logs.SiteBatchBackupError(ctx, batch, errors.New("batch failed"))
	logs.Log(ctx, &other, db.LogLevelInfo, "other synced", nil)

	events, err := q.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &site, Batch: &batch, Limit: 100})
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, string(db.EventSiteBatchBackedUp), events[0].Name)
	require.Equal(t, batch, *events[0].Batch)
	require.Equal(t, site, *events[1].SiteID)

	// Another site's or batch's event, or the server's, cannot be read by its ID either
	allEvents, err := q.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 100})
	require.NoError(t, err)
	for _, event := range allEvents {
		_, err := q.GetSiteEvent(ctx, logsdb.GetSiteEventParams{ID: event.ID, SiteID: &site, Batch: &batch})
		if event.ID == events[0].ID || event.ID == events[1].ID {
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
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs, site, batch := context.TODO(), controlDB.Logs, "typedsite", "typedbatch"

	logs.SiteCreatedEvent(ctx, site)
	logs.SiteDisabledEvent(ctx, site, "admin")
	logs.SyncError(ctx, site, consts.SyncSourceTypeDexcom, errors.New("sync failed"))
	logs.ServerError(ctx, &site, 500, "GET", "/", "request")

	str := func(s string) *string { return &s }
	tests := []struct {
		name       string
		typ        *string
		eventNames []string
		errorKinds []string
	}{
		{"every type", nil, []string{"site_disabled", "site_created"}, []string{"server", "sync"}},
		{"one event name", str("site_created"), []string{"site_created"}, nil},
		{"one error kind", str("sync"), nil, []string{"sync"}},
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
