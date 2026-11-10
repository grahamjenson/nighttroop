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
