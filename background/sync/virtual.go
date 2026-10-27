package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
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

// VirtualCarerDevice is a second uploader, a carer's phone, so device statuses come from more than one device
const VirtualCarerDevice = "virtual://carer-phone"

// The slots of the glucose wave with a meal (as glucose starts to rise) and a correction (at its peak)
const (
	virtualMealSlot       = 0
	virtualCorrectionSlot = 7
)

// A 2 hour glucose wave in mg/dL, one value per 5 minute slot.
var virtualGlucose = []float64{118, 122, 127, 133, 139, 144, 148, 150, 149, 146, 141, 135, 129, 123, 117, 112, 108, 105, 104, 105, 108, 112, 115, 117}

type virtualDoc struct {
	collection nsutils.Collection
	data       map[string]any
}

// VirtualDocuments returns the glucose readings and device statuses (from both
// devices) for every 5 minute slot in the span ending at end, a meal and a
// correction bolus once per wave, and, at the start of the window, the temp
// basal, temp target, profile switch and profile.
func VirtualDocuments(end time.Time, span time.Duration) []virtualDoc {
	end = end.UTC().Truncate(virtualSlot)
	start := end.Add(-span + virtualSlot)
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
		{nsutils.Profile, map[string]any{"defaultProfile": "Virtual", "date": startMs, "device": VirtualDevice, "store": map[string]any{"Virtual": profile}}},
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
			virtualDoc{nsutils.Devicestatus, map[string]any{"device": VirtualCarerDevice, "date": t.UnixMilli(), "uploaderBattery": 80 - idx}},
		)
		switch idx {
		case virtualMealSlot:
			docs = append(docs, virtualDoc{nsutils.Treatments, map[string]any{"eventType": "Meal Bolus", "carbs": 40, "insulin": 4, "date": t.UnixMilli(), "device": VirtualDevice}})
		case virtualCorrectionSlot:
			docs = append(docs, virtualDoc{nsutils.Treatments, map[string]any{"eventType": "Correction Bolus", "insulin": 1, "date": t.UnixMilli(), "device": VirtualDevice}})
		}
	}
	return docs
}

// WriteVirtualData writes the virtual documents for the window into the
// site through the Nightscout collections, returning how many were new.
// Collections named in skip are left out.
func WriteVirtualData(ctx context.Context, store *nightscoutapi.NightTroopAPIStore, siteContext *httputils.SiteContext, end time.Time, span time.Duration, skip ...nsutils.Collection) (int, error) {
	added := 0
	for _, doc := range VirtualDocuments(end, span) {
		if slices.Contains(skip, doc.collection) {
			continue
		}
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

// syncVirtual is the 5 minute job: it writes the current slot's glucose
// reading, device status, temp basal, temp target and profile switch. The
// profile store is only written when the source is enabled. Slots already
// written are deduped, so running more often than every 5 minutes is a no-op.
func (b *SyncJobManager) syncVirtual(syncInfo *SyncJobInfo) error {
	if len(syncInfo.UserSyncSources) == 0 {
		return fmt.Errorf("Virtual: No user sync sources found for site_id %s type %s", syncInfo.SiteID, syncInfo.Type)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for _, userSyncSource := range syncInfo.UserSyncSources {
		siteContext := virtualSiteContext(userSyncSource)
		// The first sync fills the last 30 minutes, profile included, so the site has data to show straight away
		span, skip := virtualSlot, []nsutils.Collection{nsutils.Profile}
		if syncInfo.firstSync {
			span, skip = 30*time.Minute, nil
		}
		added, err := WriteVirtualData(ctx, b.nightscoutStore, siteContext, time.Now(), span, skip...)
		if err != nil {
			return err
		}
		if added > 0 {
			slog.Info("Added virtual data for user", "subdomain", userSyncSource.SiteModel.ID, "added", added)
		}
	}
	return nil
}

// virtualSiteContext is what the virtual source writes to the site as
func virtualSiteContext(userSyncSource *SyncUserInfo) *httputils.SiteContext {
	return &httputils.SiteContext{
		Subdomain:     userSyncSource.SiteModel.ID,
		IdentityModel: &controldb.SiteIdentity{Name: "virtual_sync"},
		ControlDB:     userSyncSource.ControlDB,
		SiteDB:        userSyncSource.SiteDB,
		SiteModel:     userSyncSource.SiteModel,
		HTTPAction:    httputils.HTTPActionWrite,
	}
}
