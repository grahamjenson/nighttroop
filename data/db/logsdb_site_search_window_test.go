package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site's search looks through only its newest 10,000 logs, so its admin cannot make every search read a table it
// filled with its own rows, while paging without a search still reaches every row
func TestSiteSearchesLookThroughOnlyTheNewestRows(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	logs, q, ctx := controlDB.Logs, controlDB.Logs.Queries, t.Context()
	site, other := "windowsite", "windowother"
	logs.Log(ctx, &site, db.LogLevelInfo, "oldest needle", nil)
	for range 10000 {
		logs.Log(ctx, &site, db.LogLevelInfo, "newer", nil)
		// Another site's rows between them do not count towards the site's newest
		logs.Log(ctx, &other, db.LogLevelInfo, "other", nil)
	}
	logs.Log(ctx, &site, db.LogLevelInfo, "newest needle", nil)

	needle := "NEEDLE"
	found, err := q.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &site, Search: &needle, Limit: 10})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "newest needle", found[0].Message)

	oldest, err := q.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &site, Skip: 10001, Limit: 10})
	require.NoError(t, err)
	require.Len(t, oldest, 1, "without a search every row is paged to")
	require.Equal(t, "oldest needle", oldest[0].Message)

	// A site with fewer rows than that searches them all
	otherLog := "OTHER"
	all, err := q.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &other, Search: &otherLog, Limit: 1000})
	require.NoError(t, err)
	require.Len(t, all, 1000)
}
