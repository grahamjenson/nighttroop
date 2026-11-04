package db_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkProfileLookups finds the newest profile store, as a Nightscout profile request does, and the newest change
// to any profile, as each Nightscout lastModified request does, among a year of a looper's profile uploads. Presets are
// never truncated, so they keep growing.
// Run with go test -run XXX -bench ProfileLookups ./data/db
func BenchmarkProfileLookups(b *testing.B) {
	_, _, site := testutils.NewControlDBWithSite(b)
	ctx := b.Context()
	const hour = int64(60 * 60 * 1000)
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	require.NoError(b, err)
	insert := site.SiteDB.QueriesWithTx(tx)
	// A day's basal rates, every hour, and the profile as the loop sent it
	var rates []string
	for h := range 24 {
		rates = append(rates, fmt.Sprintf(`{"time":"%02d:00","value":0.85}`, h))
	}
	basals := "[" + strings.Join(rates, ",") + "]"
	raw := `{"basal":` + basals + `,"notes":"` + strings.Repeat("x", 1000) + `"}`
	// Four uploads a day of a store of three profiles
	for store := range int64(1500) {
		for i, name := range []string{"Default", "Exercise", "Sick"} {
			_, err := insert.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: fmt.Sprint("p", store, "-", i),
				Name: name, ProfileStoreID: fmt.Sprint("s", store), IsDefault: i == 0, BasalRatesJson: &basals,
				RawJsonData: &raw, OccurredAt: store * 6 * hour, CreatedByIdentityName: "test"})
			require.NoError(b, err)
		}
	}
	require.NoError(b, tx.Commit())
	q := site.SiteDB.Queries()

	b.Run("newest profile store", func(b *testing.B) {
		for b.Loop() {
			presets, err := q.GetMostRecentProfileStore(b.Context())
			require.NoError(b, err)
			require.Len(b, presets, 3)
			require.Equal(b, "s1499", presets[0].ProfileStoreID)
			require.True(b, presets[0].IsDefault)
		}
	})
	b.Run("last modified", func(b *testing.B) {
		for b.Loop() {
			lastModified, err := q.LastModifiedCollections(b.Context())
			require.NoError(b, err)
			require.NotZero(b, lastModified.ProfilePresetUpdatedAt)
		}
	})
}
