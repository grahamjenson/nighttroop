package retention

import (
	"math"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

// A packed day reads back as it was, with its gaps, says its version and shape, and is far smaller than its rows
func TestPackGrid(t *testing.T) {
	day := cgmDay(1)
	packed := PackGrid(Glucose, day)
	unpacked, err := UnpackGrid(packed)
	require.NoError(t, err)
	require.Len(t, unpacked, 288)
	same := func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) }
	require.True(t, slices.EqualFunc(day, unpacked, same))
	// Version 1, a series, then how it was packed, a step of 300 seconds (2 bytes), a scale of 1, and 288 steps (2 bytes)
	require.Equal(t, []byte{1, 1}, packed[:2])
	require.Equal(t, []byte{0xac, 0x02, 1, 0xa0, 0x02}, packed[3:8])
	// 276 readings are 61 KB as rows (222 bytes each with their indexes, measured), and under 200 bytes packed
	require.Less(t, len(packed), 200)
	t.Logf("a day of glucose: %d bytes packed, %d bytes as rows", len(packed), 276*222)

	// IOB to 0.01 U, which is smooth, so it is packed as the change of its change
	iob := make([]float64, 288)
	for i := range iob {
		iob[i] = math.Round(100*(2.5+2*math.Sin(float64(i)/20))) / 100
	}
	packedIOB := PackGrid(Series{Name: "iob", Step: fiveMinutes, Scale: 100}, iob)
	unpackedIOB, err := UnpackGrid(packedIOB)
	require.NoError(t, err)
	require.InDeltaSlice(t, iob, unpackedIOB, 0.0001)
	require.NotZero(t, packedIOB[2]&secondOrder)
	require.Less(t, len(packedIOB), 200)
	t.Logf("a day of IOB: %d bytes packed", len(packedIOB))

	// A day of one reading is a few bytes, and a day with none has every step, each with none
	one := Glucose.Empty()
	one[100] = 120
	packedOne := PackGrid(Glucose, one)
	require.Less(t, len(packedOne), 30)
	unpackedOne, err := UnpackGrid(packedOne)
	require.NoError(t, err)
	require.True(t, slices.EqualFunc(one, unpackedOne, same))
	empty, err := UnpackGrid(PackGrid(Glucose, Glucose.Empty()))
	require.NoError(t, err)
	require.Len(t, empty, 288)
	require.True(t, math.IsNaN(empty[0]) && math.IsNaN(empty[287]))

	// What is not a series of a version this code reads is an error, never a day of wrong numbers: no bytes, a later
	// version, another shape, a packing not known, a scale of 0, more steps than a day has, and bytes cut short
	for _, bad := range [][]byte{nil, {1, 1}, {2, 1, 0, 60, 1, 1, 1, 2}, {1, 2, 0, 60, 1, 1, 1, 2}, {1, 1, 4, 60, 1, 1, 1, 2}, {1, 1, 0, 60, 0, 1, 1, 2},
		{1, 1, 0, 60, 1, 0xa1, 0x0b}, {1, 1, 0, 60, 1, 8}, {1, 1, 0, 60, 1, 1, 1}, {1, 1, deflated, 60, 1, 1, 0xff, 0xff}, packed[:50]} {
		_, err := UnpackGrid(bad)
		require.Error(t, err, "%v", bad)
	}
}

// Whatever the values, a day reads back as it was packed
func TestPackGridOfAnyDay(t *testing.T) {
	random := rand.New(rand.NewSource(7))
	for range 200 {
		series := Series{Step: []time.Duration{time.Minute, fiveMinutes, 15 * time.Minute, hour}[random.Intn(4)], Scale: []float64{1, 10, 100, 1000}[random.Intn(4)]}
		values := series.Empty()
		for i := range values {
			if random.Intn(4) > 0 {
				values[i] = math.Round((random.Float64()*2000-1000)*series.Scale) / series.Scale
			}
		}
		unpacked, err := UnpackGrid(PackGrid(series, values))
		require.NoError(t, err)
		require.InDeltaSlice(t, nanAsZero(values), nanAsZero(unpacked), 1e-9)
	}
}

// nanAsZero is the values with a far number where there is none, to compare two days
func nanAsZero(values []float64) []float64 {
	compared := slices.Clone(values)
	for i, value := range compared {
		if math.IsNaN(value) {
			compared[i] = -1e9
		}
	}
	return compared
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
