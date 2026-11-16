package sitedb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// A site's DB refuses a row whose foreign key has no row it refers to, in memory, as every DB in tests is, as on file
func TestSiteDBForeignKeys(t *testing.T) {
	for _, mode := range []consts.ServerMode{consts.Test, consts.Development} {
		t.Run(string(mode), func(t *testing.T) {
			site, err := sitedb.NewDB(t.Context(), "abcdefgh", &config.ServerConfig{Mode: mode, DataDir: t.TempDir()})
			require.NoError(t, err)
			t.Cleanup(func() { site.Close() })
			require.Equal(t, mode == consts.Test, site.InMemory)
			_, err = site.Queries().CreateEventGroup(t.Context(), sitedb.CreateEventGroupParams{
				ID: "group", GroupType: "Note", GlucoseReadingID: new("no such reading"), OccurredAt: 1794788205631,
				CreatedByIdentityName: "admin",
			})
			require.ErrorContains(t, err, "FOREIGN KEY constraint failed")
		})
	}
}
