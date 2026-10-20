package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
)

// waveStart is a time in the first slot of the 2 hour glucose wave
var waveStart = time.Unix(int64(len(virtualGlucose))*int64(virtualSlot.Seconds())*100000, 0).UTC()

func TestVirtualDocumentsMealsAndCorrections(t *testing.T) {
	tests := []struct {
		name        string
		end         time.Time
		span        time.Duration
		meals       []time.Time
		corrections []time.Time
	}{
		{name: "a whole wave has one meal then one correction", end: waveStart.Add(23 * virtualSlot), span: 2 * time.Hour,
			meals: []time.Time{waveStart}, corrections: []time.Time{waveStart.Add(7 * virtualSlot)}},
		{name: "the meal slot alone", end: waveStart, span: virtualSlot, meals: []time.Time{waveStart}},
		{name: "the correction slot alone", end: waveStart.Add(7 * virtualSlot), span: virtualSlot, corrections: []time.Time{waveStart.Add(7 * virtualSlot)}},
		{name: "a 30 minute window after both has neither", end: waveStart.Add(20 * virtualSlot), span: 30 * time.Minute},
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
