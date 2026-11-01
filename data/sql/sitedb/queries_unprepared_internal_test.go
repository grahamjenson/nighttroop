package sitedb

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// A site's queries run unprepared: preparing all of them, for each of thousands of sites, took a site's first query
// 30ms and kept about 3.5MB of memory for each site
func TestSiteQueriesAreUnprepared(t *testing.T) {
	db, err := NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: consts.Test, DataDir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Queries().AllActivities(t.Context(), AllActivitiesParams{Limit: 1})
	require.NoError(t, err)
	require.Nil(t, db.Queries().allActivitiesStmt)
}
