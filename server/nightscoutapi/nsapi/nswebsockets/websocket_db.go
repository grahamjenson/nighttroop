package nswebsockets

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// dbMessage is a dbAdd, dbUpdate, dbUpdateUnset or dbRemove: the collection, the record's _id and the data to write
type dbMessage struct {
	Collection string          `json:"collection"`
	ID         json.RawMessage `json:"_id"`
	Data       json.RawMessage `json:"data"`
}

// dbWriter writes a checked dbMessage and is the reply
type dbWriter func(ctx context.Context, collection nsutils.Collection, id string, data json.RawMessage) any

// dbCollections are the collections a socket can write, as in Nightscout
var dbCollections = []nsutils.Collection{nsutils.Treatments, nsutils.Entries, nsutils.Devicestatus, nsutils.Profile, nsutils.Food, nsutils.Activity}

// ackResult is a reply that only says how an event went, as Nightscout's are
func ackResult(message string) map[string]string { return map[string]string{"result": message} }

// handleDBWrite checks a write as Nightscout does, answering why it is refused, then writes it through the collection
// like the REST API, so it is validated, deduplicated and sent on as a dataUpdate the same way
func (ws *Websocket) handleDBWrite(event string, args []json.RawMessage, auth *authorization, ctx context.Context, write dbWriter) any {
	var message dbMessage
	if len(args) == 0 || json.Unmarshal(args[0], &message) != nil {
		return ackResult("Invalid request")
	}
	collection := nsutils.Collection(message.Collection)
	if !slices.Contains(dbCollections, collection) {
		return ackResult("Wrong collection")
	}
	if auth == nil {
		return ackResult("Not authorized")
	}
	if collection == nsutils.Treatments && !auth.WriteTreatment || collection != nsutils.Treatments && !auth.Write {
		return ackResult("Not permitted")
	}
	// A websocket's events skip the HTTP rate limits, so each write counts against them here
	siteContext := httputils.GetSiteContext(ctx)
	limiters := ws.config.RateLimiters
	if allowed, _ := httputils.WithinRateLimitAll(limiters.SitePOSTLimiter(siteContext.SiteID), limiters.IdentityPOSTLimiter(siteContext.IdentityModel.ID)); !allowed {
		return ackResult("Too many requests")
	}
	var id string
	if event != "dbAdd" && (json.Unmarshal(message.ID, &id) != nil || id == "") {
		return ackResult("Missing or invalid _id")
	}
	return write(ctx, collection, id, message.Data)
}

// dbAdd creates a record, or each of a list, and answers with those stored (the existing one for a duplicate)
func (ws *Websocket) dbAdd(ctx context.Context, collection nsutils.Collection, _ string, data json.RawMessage) any {
	var documents []map[string]any
	if json.Unmarshal(data, &documents) != nil {
		var document map[string]any
		if json.Unmarshal(data, &document) != nil || document == nil {
			return ackResult("Invalid data")
		}
		documents = []map[string]any{document}
	}
	stored := []map[string]any{}
	for _, document := range documents {
		if _, ok := document["eventType"]; collection == nsutils.Treatments && !ok {
			document["eventType"] = "<none>"
		}
		if _, ok := document["created_at"]; !ok {
			document["created_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		}
		record, _, err := ws.collections[collection].Create(ctx, httputils.GetSiteContext(ctx), document, false)
		if err != nil || record == nil {
			slog.Info("Websocket dbAdd refused", "collection", collection, "error", err)
			continue
		}
		stored = append(stored, record.ToMap(false))
	}
	return stored
}

// dbUpdate sets the given fields on a record
func (ws *Websocket) dbUpdate(ctx context.Context, collection nsutils.Collection, id string, data json.RawMessage) any {
	var fields map[string]any
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return ackResult("Invalid data")
	}
	if record, err := ws.collections[collection].Patch(ctx, httputils.GetSiteContext(ctx), id, fields, false); err != nil || record == nil {
		slog.Info("Websocket dbUpdate refused", "collection", collection, "error", err)
		return ackResult("Unable to process update")
	}
	return ackResult("success")
}

// dbUpdateUnset removes the given fields from a record, by replacing it with what it is without them
func (ws *Websocket) dbUpdateUnset(ctx context.Context, collection nsutils.Collection, id string, data json.RawMessage) any {
	var fields map[string]any
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return ackResult("Invalid data")
	}
	siteContext := httputils.GetSiteContext(ctx)
	record, err := ws.collections[collection].Get(ctx, siteContext, id, nsquery.Query{})
	if err == nil && record != nil {
		document := record.ToMap(false)
		for field := range fields {
			delete(document, field)
		}
		_, _, err = ws.collections[collection].Replace(ctx, siteContext, id, document, false, false)
	}
	if err != nil || record == nil {
		slog.Info("Websocket dbUpdateUnset refused", "collection", collection, "error", err)
		return ackResult("Unable to process update")
	}
	return ackResult("success")
}

// dbRemove deletes a record, softly as the site does
func (ws *Websocket) dbRemove(ctx context.Context, collection nsutils.Collection, id string, _ json.RawMessage) any {
	if err := ws.collections[collection].Delete(ctx, httputils.GetSiteContext(ctx), id, false); err != nil {
		return ackResult("Unable to process removal")
	}
	return ackResult("success")
}
