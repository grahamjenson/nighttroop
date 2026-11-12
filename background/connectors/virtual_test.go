package connectors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// waveStart is a time in the first slot of the 6 hour glucose wave
var waveStart = time.Unix(virtualWaveSlots*int64(virtualSlot.Seconds())*100000, 0).UTC()

func TestVirtualDocumentsMealsAndCorrections(t *testing.T) {
	tests := []struct {
		name        string
		end         time.Time
		span        time.Duration
		meals       []time.Time
		corrections []time.Time
	}{
		{name: "a whole wave has one meal then one correction", end: waveStart.Add(71 * virtualSlot), span: 6 * time.Hour,
			meals: []time.Time{waveStart}, corrections: []time.Time{waveStart.Add(18 * virtualSlot)}},
		{name: "the meal slot alone", end: waveStart, span: virtualSlot, meals: []time.Time{waveStart}},
		{name: "the correction slot alone", end: waveStart.Add(18 * virtualSlot), span: virtualSlot, corrections: []time.Time{waveStart.Add(18 * virtualSlot)}},
		{name: "a 30 minute window after both has neither", end: waveStart.Add(30 * virtualSlot), span: 30 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meals, corrections, devices := []time.Time{}, []time.Time{}, map[any]int{}
			for _, doc := range VirtualDocuments(tt.end, tt.span) {
				switch {
				case doc.collection == nsutils.Devicestatus:
					devices[doc.data["device"]]++
				case doc.data["eventType"] == "Meal Bolus":
					require.Equal(t, 40, doc.data["carbs"])
					meals = append(meals, time.UnixMilli(doc.data["date"].(int64)).UTC())
				case doc.data["eventType"] == "Correction Bolus":
					corrections = append(corrections, time.UnixMilli(doc.data["date"].(int64)).UTC())
				}
			}
			require.ElementsMatch(t, tt.meals, meals)
			require.ElementsMatch(t, tt.corrections, corrections)

			// Both devices report once per slot
			slots := int(tt.span / virtualSlot)
			require.Equal(t, map[any]int{VirtualDevice: slots, VirtualCarerDevice: slots}, devices)
		})
	}
}

// Over a few days the virtual glucose reaches every range, from very low to very high, mostly in range or a little out,
// is never below what a CGM reads, and no day is the same as the one before
func TestVirtualGlucoseRanges(t *testing.T) {
	const day = 288
	counts := map[string]int{}
	same := true
	for slot := int64(0); slot < 4*day; slot++ {
		mgdl := virtualGlucose(slot)
		require.GreaterOrEqual(t, mgdl, 40.0)
		switch {
		case mgdl < 54:
			counts["very low"]++
		case mgdl < 70:
			counts["low"]++
		case mgdl <= 180:
			counts["in range"]++
		case mgdl <= 250:
			counts["high"]++
		default:
			counts["very high"]++
		}
		if slot >= day && mgdl != virtualGlucose(slot-day) {
			same = false
		}
	}
	for _, band := range []string{"very low", "low", "in range", "high", "very high"} {
		require.Positive(t, counts[band], band)
	}
	require.Greater(t, counts["in range"], 4*day/2, "most readings are in range")
	require.False(t, same, "each day differs from the one before")
}

// A day has two finger prick meter readings, at 07:02 and 19:02 UTC, a little above the CGM then
func TestVirtualDocumentsMeterReadings(t *testing.T) {
	day := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	var meters []time.Time
	for _, doc := range VirtualDocuments(day.Add(24*time.Hour-virtualSlot), 24*time.Hour) {
		if doc.data["type"] == "mbg" {
			at := time.UnixMilli(doc.data["date"].(int64)).UTC()
			meters = append(meters, at)
			slot := at.Add(-virtualMeterAfter).Unix() / int64(virtualSlot.Seconds())
			require.Equal(t, virtualGlucose(slot)+virtualMeterHigher, doc.data["mbg"])
		}
	}
	require.Equal(t, []time.Time{day.Add(7*time.Hour + 2*time.Minute), day.Add(19*time.Hour + 2*time.Minute)}, meters)
}

// A backfill writes nothing after the end of its window, as the meter reading 2 minutes after its slot would be, so the
// days after a backfill's oldest reading, which it fetches up to, are left as they are whatever the time it runs
func TestVirtualBackfillDocumentsEndAtTo(t *testing.T) {
	day := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		to    time.Time
		meter bool
	}{
		{name: "ending in the meter reading's slot, before it", to: day.Add(7*time.Hour + time.Minute)},
		{name: "ending just after the meter reading", to: day.Add(7*time.Hour + 3*time.Minute), meter: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meter := false
			for _, doc := range virtualBackfillDocuments(tt.to.Add(-time.Hour), tt.to) {
				date, _ := doc.data["date"].(int64)
				require.LessOrEqual(t, date, tt.to.UnixMilli(), "a %s document after the backfill's end", doc.collection)
				meter = meter || doc.data["type"] == "mbg"
			}
			require.Equal(t, tt.meter, meter)
		})
	}
}

// Windows of any span, as a backfill's ending now is, write on the 5 minute grid, so ones that overlap, as a backfill's
// recent days and the sync's do, write the same documents at the same times, which dedupe, rather than another of each
// a few minutes off
func TestVirtualDocumentsOnTheSlotGrid(t *testing.T) {
	end := waveStart.Add(3*time.Hour + 2*time.Minute + 17*time.Second)
	dates := func(span time.Duration) map[int64]bool {
		all := map[int64]bool{}
		for _, doc := range VirtualDocuments(end, span) {
			if doc.collection == nsutils.Entries && doc.data["type"] == "sgv" {
				date := doc.data["date"].(int64)
				require.Zero(t, date%virtualSlot.Milliseconds(), "a reading on the grid")
				all[date] = true
			}
		}
		return all
	}
	short, long := dates(time.Hour+3*time.Minute+41*time.Second), dates(2*time.Hour+11*time.Second)
	require.Len(t, short, 13)
	for date := range short {
		require.True(t, long[date], "the longer window has the shorter's readings at the same times")
	}
}

// The first sync writes the same 6 readings whenever it runs, even in the half hour after a meter reading, which is
// left for a backfill, so a new site's readings can be counted; a sync of the meter reading's slot writes it
func TestVirtualSyncDocumentsReadings(t *testing.T) {
	day := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		now       time.Time
		firstSync bool
		sensor    int
		meter     int
	}{
		{name: "a first sync at midday", now: day.Add(12*time.Hour + 3*time.Minute), firstSync: true, sensor: 6},
		{name: "a first sync just after midnight", now: day.Add(10 * time.Minute), firstSync: true, sensor: 6},
		{name: "a first sync just after the 07:02 meter reading", now: day.Add(7*time.Hour + 10*time.Minute), firstSync: true, sensor: 6},
		{name: "a first sync just before it", now: day.Add(7*time.Hour + time.Minute), firstSync: true, sensor: 6},
		{name: "a first sync before the 19:02 meter reading's half hour ends", now: day.Add(19*time.Hour + 29*time.Minute), firstSync: true, sensor: 6},
		{name: "a sync of the meter reading's slot", now: day.Add(7*time.Hour + 3*time.Minute), sensor: 1, meter: 1},
		{name: "a sync of another slot", now: day.Add(7*time.Hour + 10*time.Minute), sensor: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts, profiles := map[any]int{}, 0
			for _, doc := range virtualSyncDocuments(tt.now, tt.firstSync) {
				switch doc.collection {
				case nsutils.Entries:
					counts[doc.data["type"]]++
				case nsutils.Profile:
					profiles++
				}
			}
			require.Equal(t, tt.sensor, counts["sgv"])
			require.Equal(t, tt.meter, counts["mbg"])
			// Only the first sync writes the profile
			require.Equal(t, tt.firstSync, profiles == 1)
		})
	}
}
