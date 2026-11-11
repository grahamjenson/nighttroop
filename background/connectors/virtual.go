package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// The virtual connector writes fixed, repeatable data so a site can be
// tried out without a CGM. Nothing here is random: every value is looked up
// from the 5 minute slot it belongs to, so writing the same window twice
// produces the same documents and the collections dedupe them.

const VirtualDevice = "virtual://nighttroop"
const virtualSlot = 5 * time.Minute

// virtualIdentity is who the virtual source writes to the site as
const virtualIdentity = "virtual_connector"

// VirtualCarerDevice is a second uploader, a carer's phone, so device statuses come from more than one device
const VirtualCarerDevice = "virtual://carer-phone"

// The virtual glucose is two waves added to 135 mg/dL, so it changes through the day and from one day to the next, as
// real glucose does: a 6 hour wave, rising 90 mg/dL after a meal as it starts to rise, with a correction at its peak,
// then dipping up to 50 mg/dL, sharply and briefly as a low does; and a slower 25 hour one of 40 mg/dL, so no two days
// are the same. Over days about 2% of readings are very low (below 54 mg/dL), 4% low, 63% in range, 27% high and 4%
// very high (above 250 mg/dL), so every range has readings.
const (
	virtualWaveSlots      = 72  // 6 hours of 5 minute slots
	virtualDriftSlots     = 300 // 25 hours
	virtualMealSlot       = 0
	virtualCorrectionSlot = virtualWaveSlots / 4
)

// A finger prick meter reading is taken twice a day, at 07:00 and 19:00 UTC, the slot this many into each 12 hours,
// 2 minutes after that slot's CGM reading and 6 mg/dL above it, so the two are not taken for the same glucose
const (
	virtualMeterSlots  = 144
	virtualMeterSlot   = 84
	virtualMeterAfter  = 2 * time.Minute
	virtualMeterHigher = 6
)

// virtualGlucose is the glucose in mg/dL in the 5 minute slot, counted from 1970, no lower than a CGM reads
func virtualGlucose(slot int64) float64 {
	wave := math.Sin(2 * math.Pi * float64(slot%virtualWaveSlots) / virtualWaveSlots)
	if wave > 0 {
		wave *= 90
	} else {
		wave = 50 * wave * wave * wave
	}
	drift := 40 * math.Sin(2*math.Pi*float64(slot%virtualDriftSlots)/virtualDriftSlots)
	return max(40, math.Round(135+wave+drift))
}

type virtualDoc struct {
	collection nsutils.Collection
	data       map[string]any
}

// VirtualDocuments returns the glucose readings and device statuses (from both
// devices) for every 5 minute slot in the span ending at end, a meal and a
// correction bolus once per wave, and, at the start of the window, the temp
// basal, temp target, profile switch and profile.
func VirtualDocuments(end time.Time, span time.Duration) []virtualDoc {
	// The window is its slots after end less span up to end's, every one on the 5 minute grid, so windows of any span
	// that overlap, as a backfill's and the sync's do, write the same documents at the same times and they dedupe
	start := end.UTC().Add(-span).Truncate(virtualSlot).Add(virtualSlot)
	end = end.UTC().Truncate(virtualSlot)
	startMs := start.UnixMilli()
	block := func(value float64) []map[string]any {
		return []map[string]any{{"time": "00:00", "timeAsSeconds": 0, "value": value}}
	}
	profile := map[string]any{
		"dia": 5, "units": "mg/dL", "timezone": "UTC", "carbs_hr": 20,
		"basal": block(0.8), "sens": block(50), "carbratio": block(10), "target_low": block(100), "target_high": block(120),
	}
	// The profile switch carries the profile it switched to, as AAPS sends it
	profileJSON, _ := json.Marshal(profile)
	docs := []virtualDoc{
		{nsutils.Treatments, map[string]any{"eventType": "Temp Basal", "rate": 0.85, "absolute": 0.85, "duration": 30, "date": startMs, "device": VirtualDevice}},
		{nsutils.Treatments, map[string]any{"eventType": "Temporary Target", "targetBottom": 100, "targetTop": 120, "units": "mg/dL", "duration": 30, "reason": "Virtual", "date": startMs, "device": VirtualDevice}},
		{nsutils.Treatments, map[string]any{"eventType": "Profile Switch", "profile": "Virtual", "percentage": 100, "duration": 0, "profileJson": string(profileJSON), "date": startMs, "device": VirtualDevice}},
		// Its Loop settings, without where pushes go, so it registers no push device
		{nsutils.Profile, map[string]any{"defaultProfile": "Virtual", "date": startMs, "device": VirtualDevice, "store": map[string]any{"Virtual": profile},
			"loopSettings": map[string]any{"dosingEnabled": true, "dosingStrategy": "automaticBolus", "maximumBasalRatePerHour": 3, "maximumBolus": 10, "minimumBGGuard": 80, "preMealTargetRange": []float64{80, 90}}}},
	}
	for t := start; !t.After(end); t = t.Add(virtualSlot) {
		slot := t.Unix() / int64(virtualSlot.Seconds())
		// idx is the slot of the 6 hour wave, and the device's values cycle every 2 hours of it
		idx := int(slot % virtualWaveSlots)
		sgv := virtualGlucose(slot)
		delta := sgv - virtualGlucose(slot-1)
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
				"uploader": map[string]any{"battery": 100 - idx%24, "name": "Virtual", "timestamp": iso},
				"pump":     map[string]any{"reservoir": 150 - idx%24, "clock": iso, "battery": map[string]any{"percent": 90}},
				"loop": map[string]any{"name": "Virtual", "timestamp": iso,
					"iob": map[string]any{"iob": 0.5 + float64(idx%24)*0.05, "timestamp": iso},
					"cob": map[string]any{"cob": 12 - idx%12, "timestamp": iso}},
			}},
			virtualDoc{nsutils.Devicestatus, map[string]any{"device": VirtualCarerDevice, "date": t.UnixMilli(), "uploaderBattery": 80 - idx%24}},
		)
		if slot%virtualMeterSlots == virtualMeterSlot {
			docs = append(docs, virtualDoc{nsutils.Entries, map[string]any{"type": "mbg", "mbg": sgv + virtualMeterHigher, "date": t.Add(virtualMeterAfter).UnixMilli(), "device": VirtualDevice}})
		}
		switch idx {
		case virtualMealSlot:
			docs = append(docs, virtualDoc{nsutils.Treatments, map[string]any{"eventType": "Meal Bolus", "carbs": 40, "insulin": 4, "date": t.UnixMilli(), "device": VirtualDevice}})
		case virtualCorrectionSlot:
			docs = append(docs, virtualDoc{nsutils.Treatments, map[string]any{"eventType": "Correction Bolus", "insulin": 1, "date": t.UnixMilli(), "device": VirtualDevice}})
		}
	}
	return docs
}

// WriteVirtualData writes the virtual documents for the window into the site through the Nightscout collections,
// returning how many were new
func WriteVirtualData(ctx context.Context, store *nightscoutapi.NightTroopAPIStore, siteContext *httputils.SiteContext, end time.Time, span time.Duration) (int, error) {
	return writeVirtualDocuments(ctx, store, siteContext, VirtualDocuments(end, span))
}

// writeVirtualDocuments writes the documents into the site, but no device statuses older than deviceStatusBackfillAge,
// returning how many were new
func writeVirtualDocuments(ctx context.Context, store *nightscoutapi.NightTroopAPIStore, siteContext *httputils.SiteContext, docs []virtualDoc) (int, error) {
	added, oldestStatus := 0, time.Now().Add(-deviceStatusBackfillAge).UnixMilli()
	for _, doc := range docs {
		// As a Nightscout backfill copies, no device status older than deviceStatusBackfillAge
		if date, _ := doc.data["date"].(int64); doc.collection == nsutils.Devicestatus && date < oldestStatus {
			continue
		}
		collectionStore, ok := store.GetStore(doc.collection)
		if !ok {
			return added, fmt.Errorf("virtual: collection %s not found", doc.collection)
		}
		_, duplicate, herr := collectionStore.Create(ctx, siteContext, doc.data)
		if herr != nil {
			return added, fmt.Errorf("virtual: failed to create %s: %s", doc.collection, herr.Error())
		}
		if !duplicate {
			added++
		}
	}
	return added, nil
}

// syncVirtual is the 5 minute job: it writes the current slot's glucose
// reading, device status, temp basal, temp target and profile switch. The
// profile store is only written when the source is enabled. Slots already
// written are deduped, so running more often than every 5 minutes is a no-op.
func (b *ConnectorJobManager) syncVirtual(jobInfo *ConnectorJobInfo) (int, error) {
	ctx, cancel := context.WithTimeout(jobInfo.ctx, 20*time.Second)
	defer cancel()

	added, err := writeVirtualDocuments(ctx, b.nightscoutStore, jobInfo.Site.siteContext(virtualIdentity), virtualSyncDocuments(time.Now(), jobInfo.firstSync))
	if err != nil {
		return 0, err
	}
	if added > 0 {
		slog.InfoContext(ctx, "Added virtual data for user", "added", added)
	}
	return added, nil
}

// virtualSyncDocuments are what a sync at now writes: its slot's, without the profile; or, the first sync, the last 30
// minutes', profile included, so the site has data to show straight away. The first sync leaves out a meter reading,
// so it is the same 6 readings whenever it runs, not 7 in the half hour after 07:00 and 19:00 UTC; a backfill writes it.
func virtualSyncDocuments(now time.Time, firstSync bool) []virtualDoc {
	if !firstSync {
		return slices.DeleteFunc(VirtualDocuments(now, virtualSlot), func(doc virtualDoc) bool { return doc.collection == nsutils.Profile })
	}
	return slices.DeleteFunc(VirtualDocuments(now, 30*time.Minute), func(doc virtualDoc) bool { return doc.data["type"] == "mbg" })
}
