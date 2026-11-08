package db_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestSiteSearchesFindWhatTheGeneralSearchesFind checks the site's own searches, which its logs pages use, find the
// same rows, in the same order, as the general searches given the site, for every filter they take
func TestSiteSearchesFindWhatTheGeneralSearchesFind(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, logs, q := context.TODO(), controlDB.Logs, controlDB.Logs.Queries
	site, other, batch := "searchsite", "searchother", "searchbatch"

	for i := range 3 {
		for _, s := range []string{site, other} {
			logs.Log(ctx, &s, db.LogLevelInfo, fmt.Sprintf("synced %d", i), map[string]any{"readings": i})
			logs.ConnectorError(ctx, s, consts.ConnectorTypeDexcom, fmt.Errorf("sync failed %d", i))
			logs.ServerError(ctx, &s, 500, "GET", "/", "request", "")
			logs.SiteDisabledEvent(ctx, s, "admin")
			logs.SiteCreatedEvent(ctx, s)
		}
		logs.Log(ctx, nil, db.LogLevelWarn, "server wide synced", nil)
		logs.SiteBatchBackedUpEvent(ctx, batch, 2, time.Now(), false)
		logs.SiteBatchBackupError(ctx, batch, errors.New("batch failed"))
	}
	allEvents, err := q.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 1000})
	require.NoError(t, err)

	str := func(s string) *string { return &s }
	containsFold := func(s, sub string) bool { return strings.Contains(strings.ToLower(s), strings.ToLower(sub)) }
	tests := []struct {
		name   string
		typ    *string
		search *string
		skip   int64
		limit  int64
	}{
		{"everything", nil, nil, 0, 100},
		{"a page", nil, nil, 2, 3},
		{"searched", nil, str("SYNC"), 0, 100},
		{"searched in attributes", nil, str("readings"), 0, 100},
		{"searched for nothing there", nil, str("nope"), 0, 100},
		{"an error kind or event name", str("connector"), nil, 0, 100},
		{"an event name", str("site_batch_backed_up"), nil, 0, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantLogs, err := q.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &site, Search: tt.search, Skip: tt.skip, Limit: tt.limit})
			require.NoError(t, err)
			gotLogs, err := q.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &site, Search: tt.search, Skip: tt.skip, Limit: tt.limit})
			require.NoError(t, err)
			require.Equal(t, wantLogs, gotLogs)

			wantErrors, err := q.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Type: tt.typ, Search: tt.search, Skip: tt.skip, Limit: tt.limit})
			require.NoError(t, err)
			gotErrors, err := q.SearchSiteErrors(ctx, logsdb.SearchSiteErrorsParams{SiteID: &site, Type: tt.typ, Search: tt.search, Skip: tt.skip, Limit: tt.limit})
			require.NoError(t, err)
			require.Equal(t, wantErrors, gotErrors)

			// The site's events and its batch's, though only the names of the batch's are searched
			wantEvents := []*logsdb.Event{}
			for _, event := range allEvents {
				own := event.SiteID != nil && *event.SiteID == site
				batched := event.SiteID == nil && event.Batch != nil && *event.Batch == batch
				searched := tt.search == nil || containsFold(event.Name, *tt.search) || (own && event.Attributes != nil && containsFold(*event.Attributes, *tt.search))
				if (own || batched) && (tt.typ == nil || event.Name == *tt.typ) && searched {
					wantEvents = append(wantEvents, event)
				}
			}
			wantEvents = wantEvents[min(tt.skip, int64(len(wantEvents))):]
			wantEvents = wantEvents[:min(tt.limit, int64(len(wantEvents)))]
			gotEvents, err := q.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &site, Batch: &batch, Type: tt.typ, Search: tt.search, Skip: tt.skip, Limit: tt.limit})
			require.NoError(t, err)
			if len(wantEvents) == 0 {
				require.Empty(t, gotEvents)
			} else {
				require.Equal(t, wantEvents, gotEvents)
			}
		})
	}
}

// BenchmarkSiteLogsSearch compares a site's page of logs read with the general search and with the site's own,
// among the rows of many other sites, as a full logs DB has. Run with go test -bench SiteLogsSearch ./data/db
func BenchmarkSiteLogsSearch(b *testing.B) {
	_, controlDB := testutils.NewControlDB(b)
	ctx, logs := context.TODO(), controlDB.Logs
	site := "benchsite"

	// A full logs DB's rows, a few of them the site's and the rest of a thousand other sites
	tx, err := logs.GetRawDB().BeginTx(ctx, nil)
	require.NoError(b, err)
	insert := logs.Queries.WithTx(tx)
	for i := range 100000 {
		rowSite := fmt.Sprintf("othersite%d", i%1000)
		if i%10000 == 0 {
			rowSite = site
		}
		require.NoError(b, insert.InsertLog(ctx, logsdb.InsertLogParams{SiteID: &rowSite, Level: "info", Message: "synced"}))
		require.NoError(b, insert.InsertError(ctx, logsdb.InsertErrorParams{SiteID: &rowSite, Message: "sync failed"}))
	}
	require.NoError(b, tx.Commit())

	b.Run("general logs", func(b *testing.B) {
		for b.Loop() {
			rows, err := logs.Queries.SearchLogs(ctx, logsdb.SearchLogsParams{SiteID: &site, Limit: 100})
			require.NoError(b, err)
			require.Len(b, rows, 10)
		}
	})
	b.Run("site logs", func(b *testing.B) {
		for b.Loop() {
			rows, err := logs.Queries.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &site, Limit: 100})
			require.NoError(b, err)
			require.Len(b, rows, 10)
		}
	})
	b.Run("general errors", func(b *testing.B) {
		for b.Loop() {
			rows, err := logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &site, Limit: 100})
			require.NoError(b, err)
			require.Len(b, rows, 10)
		}
	})
	b.Run("site errors", func(b *testing.B) {
		for b.Loop() {
			rows, err := logs.Queries.SearchSiteErrors(ctx, logsdb.SearchSiteErrorsParams{SiteID: &site, Limit: 100})
			require.NoError(b, err)
			require.Len(b, rows, 10)
		}
	})
}
