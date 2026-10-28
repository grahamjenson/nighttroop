package nswebsockets

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// recentData is the first dataUpdate, as Nightscout sends it: entries and treatments over the history, oldest
// first and capped like a search, the latest device statuses and profile, and the food
func (ws *Websocket) recentData(ctx context.Context, history time.Duration) map[string]any {
	now := time.Now()
	since := now.Add(-history).UnixMilli()
	data := map[string]any{"lastUpdated": now.UnixMilli()}
	newest := &nsquery.SortOccurredAtDesc
	recent := func(limit int64) nsquery.Query {
		return nsquery.Query{DateGTE: &since, HasFilters: true, Limit: limit, SortOrder: newest}
	}
	queries := map[nsutils.Collection]nsquery.Query{
		nsutils.Entries:      recent(maxRecords),
		nsutils.Treatments:   recent(maxRecords),
		nsutils.Devicestatus: recent(maxDeviceStatuses),
		nsutils.Food:         {Limit: maxRecords},
		nsutils.Profile:      {Limit: 1, SortOrder: newest},
	}
	for _, key := range []string{"sgvs", "mbgs", "cals", "treatments", "devicestatus", "profiles", "food"} {
		data[key] = []map[string]any{}
	}
	for collection, query := range queries {
		records, err := ws.collections[collection].Search(ctx, httputils.GetSiteContext(ctx), query)
		if err != nil {
			slog.Error("Websocket failed to load recent data", "collection", collection, "error", err)
			continue
		}
		slices.Reverse(records)
		for _, record := range records {
			if key, value, ok := clientData(collection, record); ok {
				data[key] = append(data[key].([]map[string]any), value)
			}
		}
	}
	return data
}

// retroDeviceStatuses are the device statuses over Nightscout's two days, oldest first and capped like a search
func (ws *Websocket) retroDeviceStatuses(ctx context.Context) []map[string]any {
	since := time.Now().Add(-maxHistoryHours * time.Hour).UnixMilli()
	query := nsquery.Query{DateGTE: &since, HasFilters: true, Limit: maxRecords, SortOrder: &nsquery.SortOccurredAtDesc}
	records, err := ws.collections[nsutils.Devicestatus].Search(ctx, httputils.GetSiteContext(ctx), query)
	if err != nil {
		slog.Error("Websocket failed to load retro device statuses", "error", err)
	}
	slices.Reverse(records)
	statuses := make([]map[string]any, 0, len(records))
	for _, record := range records {
		_, status, _ := clientData(nsutils.Devicestatus, record)
		statuses = append(statuses, status)
	}
	return statuses
}

// recorder is a collection that can show a site's model as its record
type recorder interface {
	Record(model any, siteContext *httputils.SiteContext) (nsutils.Record, bool)
}

// broadcastChange sends a changed model of the site to its DataReceivers as a delta dataUpdate, and to its /storage
// subscribers, when it is a Nightscout record
func (ws *Websocket) broadcastChange(main *socketio.Namespace, event events.Event) {
	siteDB, siteModel, err := ws.controlDB.GetSite(event.SiteID)
	if err != nil {
		return
	}
	siteContext := httputils.GetSiteContext(httputils.CreateSiteContext(context.Background(), ws.config, ws.controlDB, siteDB, siteModel, event.SiteID, "", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON))
	for collection, store := range ws.collections {
		r, ok := store.(recorder)
		if !ok {
			continue
		}
		if record, ok := r.Record(event.Model, siteContext); ok {
			ws.sendDataUpdate(main, collection, record, event.Action)
			ws.sendToStorage(siteDB, siteModel, event, collection, r)
			return
		}
	}
}

// sendDataUpdate sends a changed record to the DataReceivers as a delta dataUpdate. As in Nightscout, a treatment
// says whether it was updated or removed, and other removals are not sent.
func (ws *Websocket) sendDataUpdate(main *socketio.Namespace, collection nsutils.Collection, record nsutils.Record, action events.Action) {
	key, value, ok := clientData(collection, record)
	if !ok || (action == events.Deleted && collection != nsutils.Treatments) {
		return
	}
	if collection == nsutils.Treatments && action != events.Created {
		if action == events.Deleted {
			value = pick(value, "_id", "mills")
			value["action"] = "remove"
		} else {
			value["action"] = "update"
		}
	}
	main.EmitTo(dataReceivers, "dataUpdate", map[string]any{
		"delta":       true,
		"lastUpdated": time.Now().UnixMilli(),
		key:           []map[string]any{value},
	})
}

// clientData is a record as Nightscout's client data holds it, and its key in a dataUpdate
func clientData(collection nsutils.Collection, record nsutils.Record) (string, map[string]any, bool) {
	out := record.ToMap(false)
	out["mills"] = record.OccuredAt
	switch collection {
	case nsutils.Entries:
		switch out["type"] {
		case "sgv":
			out["mgdl"] = mgdl(out["sgv"], out["units"])
			return "sgvs", pick(out, "_id", "mgdl", "mills", "device", "direction", "filtered", "unfiltered", "noise", "rssi", "type"), true
		case "mbg":
			out["mgdl"] = mgdl(out["mbg"], out["units"])
			return "mbgs", pick(out, "_id", "mgdl", "mills", "device", "type"), true
		case "cal":
			return "cals", out, true
		}
	case nsutils.Treatments:
		return "treatments", out, true
	case nsutils.Devicestatus:
		return "devicestatus", out, true
	case nsutils.Profile:
		return "profiles", out, true
	case nsutils.Food:
		return "food", out, true
	case nsutils.Activity:
		return "activity", out, true
	}
	return "", nil, false
}

// mgdl is a glucose value in mg/dL, as the client data always has it, from the site's units
func mgdl(value any, units any) any {
	glucose, ok := value.(float64)
	if !ok {
		return value
	}
	var from *consts.Units
	if parsed, err := consts.ParseUnits(fmt.Sprint(units)); err == nil {
		from = &parsed
	}
	return math.Round(from.ToMGDL(glucose))
}

// pick is the keys of m that have a value
func pick(m map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if value, ok := m[key]; ok && value != nil {
			out[key] = value
		}
	}
	return out
}
