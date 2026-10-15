package sync

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// The virtual sync source writes fixed, repeatable data so a site can be
// tried out without a CGM. Nothing here is random: every value is looked up
// from the 5 minute slot it belongs to, so writing the same window twice
// produces the same documents and the collections dedupe them.

const VirtualDevice = "virtual://nighttroop"
const virtualSlot = 5 * time.Minute

// A 2 hour glucose wave in mg/dL, one value per 5 minute slot.
var virtualGlucose = []float64{118, 122, 127, 133, 139, 144, 148, 150, 149, 146, 141, 135, 129, 123, 117, 112, 108, 105, 104, 105, 108, 112, 115, 117}

type virtualDoc struct {
	collection nsutils.Collection
	data       map[string]any
}

// VirtualDocuments returns the glucose readings, device statuses and, at the
// start of the window, the temp basal, temp target, profile switch and
// profile for every 5 minute slot in the span ending at end.
func VirtualDocuments(end time.Time, span time.Duration) []virtualDoc {
	end = end.UTC().Truncate(virtualSlot)
	start := end.Add(-span + virtualSlot)
	startMs := start.UnixMilli()
	block := func(value float64) []map[string]any {
		return []map[string]any{{"time": "00:00", "timeAsSeconds": 0, "value": value}}
	}
	docs := []virtualDoc{
		{nsutils.Treatments, map[string]any{"eventType": "Temp Basal", "rate": 0.85, "absolute": 0.85, "duration": 30, "date": startMs, "device": VirtualDevice}},
		{nsutils.Treatments, map[string]any{"eventType": "Temporary Target", "targetBottom": 100, "targetTop": 120, "units": "mg/dL", "duration": 30, "reason": "Virtual", "date": startMs, "device": VirtualDevice}},
		{nsutils.Treatments, map[string]any{"eventType": "Profile Switch", "profile": "Virtual", "percentage": 100, "duration": 0, "date": startMs, "device": VirtualDevice}},
		{nsutils.Profile, map[string]any{"defaultProfile": "Virtual", "date": startMs, "device": VirtualDevice, "store": map[string]any{"Virtual": map[string]any{
			"dia": 5, "units": "mg/dL", "timezone": "UTC", "carbs_hr": 20,
			"basal": block(0.8), "sens": block(50), "carbratio": block(10), "target_low": block(100), "target_high": block(120),
		}}}},
	}
	for t := start; !t.After(end); t = t.Add(virtualSlot) {
		idx := int(t.Unix()/int64(virtualSlot.Seconds())) % len(virtualGlucose)
		sgv := virtualGlucose[idx]
		delta := sgv - virtualGlucose[(idx+len(virtualGlucose)-1)%len(virtualGlucose)]
		direction := "Flat"
		if delta > 3 {
			direction = "FortyFiveUp"
		} else if delta < -3 {
			direction = "FortyFiveDown"
		}
		iso := t.Format(time.RFC3339)
		docs = append(docs,
			virtualDoc{nsutils.Entries, map[string]any{"type": "sgv", "sgv": sgv, "direction": direction, "date": t.UnixMilli(), "device": VirtualDevice}},
			virtualDoc{nsutils.Devicestatus, map[string]any{"device": VirtualDevice, "date": t.UnixMilli(),
				"uploader": map[string]any{"battery": 100 - idx, "name": "Virtual", "timestamp": iso},
				"pump":     map[string]any{"reservoir": 150 - idx, "clock": iso, "battery": map[string]any{"percent": 90}},
				"loop": map[string]any{"name": "Virtual", "timestamp": iso,
					"iob": map[string]any{"iob": 0.5 + float64(idx)*0.05, "timestamp": iso},
					"cob": map[string]any{"cob": 12 - idx%12, "timestamp": iso}},
			}},
		)
	}
	return docs
}

// WriteVirtualData writes the virtual documents for the window into the
// site through the Nightscout collections, returning how many were new.
func WriteVirtualData(ctx context.Context, store *nightscoutapi.NightTroopAPIStore, siteContext *httputils.SiteContext, end time.Time, span time.Duration) (int, error) {
	added := 0
	for _, doc := range VirtualDocuments(end, span) {
		collectionStore, ok := store.GetStore(doc.collection)
		if !ok {
			return added, fmt.Errorf("virtual: collection %s not found", doc.collection)
		}
		_, duplicate, herr := collectionStore.Create(ctx, siteContext, doc.data, false)
		if herr != nil {
			return added, fmt.Errorf("virtual: failed to create %s: %s", doc.collection, herr.Error())
		}
		if !duplicate {
			added++
		}
	}
	return added, nil
}
