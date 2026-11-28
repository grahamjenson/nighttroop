package retention_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/utils/testutils"
)

// A new table of the site's DB must say how it stays small
func TestEveryTableHasAPolicy(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	tables, err := site.SiteDB.TableSizes(t.Context())
	require.NoError(t, err)
	seen := 0
	for _, table := range tables {
		if table.Name == "sqlite_schema" || table.Name == "schema_migrations" {
			continue
		}
		seen++
		policy, ok := retention.Policies[table.Name]
		require.True(t, ok, "%s has no retention policy", table.Name)
		if policy.Shape == retention.SeriesShape {
			require.NotEmpty(t, policy.Series, "%s is deleted once its day is sealed, so it must fold into a series", table.Name)
			require.NotEmpty(t, policy.Lost, "%s must say what is lost", table.Name)
		}
	}
	require.Len(t, retention.Policies, seen, "a policy for a table that is gone")
}
