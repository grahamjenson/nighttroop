package db_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestLogsDBRecordsBySite(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)
	ctx, logs, site := context.TODO(), controlDB.Logs, "somesite"

	logs.Log(ctx, &site, db.LogLevelInfo, "synced", map[string]any{"readings": 3})
	logs.Log(ctx, nil, db.LogLevelWarn, "server wide", nil)
	logs.Event(ctx, &site, "site_disabled", nil)
	logs.Error(ctx, &site, errors.New("sync failed"), map[string]any{"sync_type": "dexcom"})

	siteLogs, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Len(t, siteLogs, 1)
	require.Equal(t, "synced", siteLogs[0].Message)
	require.Equal(t, db.LogLevelInfo, siteLogs[0].Level)
	require.Equal(t, `{"readings":3}`, *siteLogs[0].Attributes)

	allLogs, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, allLogs, 2)
	require.Nil(t, allLogs[0].SiteID)

	events, err := logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_disabled", events[0].Name)

	siteErrors, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "sync failed", siteErrors[0].Message)
	require.Equal(t, `{"sync_type":"dexcom"}`, *siteErrors[0].Attributes)
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
				logs.Event(ctx, nil, string(rune('a'+i)), nil)
				logs.Error(ctx, nil, errors.New(string(rune('a'+i))), nil)
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
