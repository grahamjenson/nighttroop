package sitedb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// Profile stores come out in the order the query returned them, newest first, every time
func TestManyPresetsToWrappersKeepsOrder(t *testing.T) {
	presets := []*sitedb.ProfilePreset{
		{ID: "a1", ProfileStoreID: "newest"},
		{ID: "b1", ProfileStoreID: "middle"},
		{ID: "a2", ProfileStoreID: "newest", IsDefault: true},
		{ID: "c1", ProfileStoreID: "oldest"},
	}
	// Map order is random, so one lucky run is not enough
	for range 20 {
		wrappers := sitedb.ManyPresetsToMayWrappers(presets)
		require.Len(t, wrappers, 3)
		order := []string{}
		for _, wrapper := range wrappers {
			order = append(order, (*wrapper)[0].ProfileStoreID)
		}
		require.Equal(t, []string{"newest", "middle", "oldest"}, order)
		// The default preset still leads its store
		require.Equal(t, "a2", (*wrappers[0])[0].ID)
	}
}
