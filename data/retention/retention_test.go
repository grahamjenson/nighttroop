package retention

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
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
		policy, ok := Policies[table.Name]
		require.True(t, ok, "%s has no retention policy", table.Name)
		if policy.Class == Signal || policy.Class == Telemetry {
			require.NotEmpty(t, policy.Series, "%s is deleted when its day is sealed, so it must fold into a series", table.Name)
			require.NotEmpty(t, policy.Lost, "%s must say what is lost", table.Name)
		}
	}
	require.Len(t, Policies, seen, "a policy for a table that is gone")
}

// cgmDay is a day of a CGM's glucose every 5 minutes, as a person's is: a slow wave, a few mg/dL of noise, and an
// hour with no readings
func cgmDay(seed int64) []float64 {
	random, values := rand.New(rand.NewSource(seed)), make([]float64, 288)
	for i := range values {
		values[i] = math.Round(140 + 60*math.Sin(float64(i)/40) + 25*math.Sin(float64(i)/11) + random.Float64()*6)
		if i >= 100 && i < 112 {
			values[i] = math.NaN()
		}
	}
	return values
}

// A packed day reads back as it was, with its gaps, and is far smaller than its rows
func TestPackGrid(t *testing.T) {
	day := cgmDay(1)
	packed := PackGrid(day, 1)
	unpacked, err := UnpackGrid(packed, 1)
	require.NoError(t, err)
	require.Len(t, unpacked, 288)
	for i := range day {
		require.True(t, day[i] == unpacked[i] || (math.IsNaN(day[i]) && math.IsNaN(unpacked[i])), "step %d: %v != %v", i, day[i], unpacked[i])
	}
	// 276 readings are 61 KB as rows (222 bytes each with their indexes, measured), and under 400 bytes packed
	require.Less(t, len(packed), 400)
	t.Logf("a day of glucose: %d bytes packed, %d bytes as rows", len(packed), 276*222)

	// IOB to 0.01 U
	iob := make([]float64, 288)
	for i := range iob {
		iob[i] = math.Round(100*(2.5+2*math.Sin(float64(i)/20))) / 100
	}
	packedIOB := PackGrid(iob, 100)
	unpackedIOB, err := UnpackGrid(packedIOB, 100)
	require.NoError(t, err)
	require.InDeltaSlice(t, iob, unpackedIOB, 0.0001)
	require.Less(t, len(packedIOB), 400)

	// An empty day, and bytes that are not a grid
	empty, err := UnpackGrid(PackGrid(make([]float64, 0), 1), 1)
	require.NoError(t, err)
	require.Empty(t, empty)
	for _, bad := range [][]byte{nil, {9, 1}, {gridVersion}, {gridVersion, 200, 2}, packed[:50]} {
		_, err := UnpackGrid(bad, 1)
		require.Error(t, err)
	}
}

// The stats of two weeks merged are the stats of all their values, so a month or a year needs no row read again, and
// the time in a range the person changes later is still known
func TestStatsMerge(t *testing.T) {
	var all, merged Stats
	for seed := int64(1); seed <= 14; seed++ {
		var day Stats
		for _, value := range cgmDay(seed) {
			if !math.IsNaN(value) {
				day.Add(value)
				all.Add(value)
			}
		}
		merged.Merge(day)
	}
	merged.Merge(Stats{})
	require.Equal(t, all.Count, merged.Count)
	require.Equal(t, all.Bins, merged.Bins)
	require.InDelta(t, all.Mean(), merged.Mean(), 1e-9)
	require.InDelta(t, all.SD(), merged.SD(), 1e-9)
	require.Equal(t, all.Min, merged.Min)
	require.Equal(t, all.Max, merged.Max)
	require.InDelta(t, 100, merged.PercentBetween(0, 1000), 1e-9)
	// In range by today's range, and by a tighter one chosen later
	require.InDelta(t, all.PercentBetween(70, 180), merged.PercentBetween(70, 180), 1e-9)
	require.Greater(t, merged.PercentBetween(70, 180), merged.PercentBetween(70, 140))
}
