package nswebsockets

import (
	"context"
	"slices"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nscollections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/utils/httputils"
)

// storageCollections are those API v3 serves, which a /storage socket can subscribe to
var storageCollections = []nsutils.Collection{nsutils.Devicestatus, nsutils.Entries, nsutils.Food, nsutils.Profile, nsutils.Settings, nsutils.Treatments}

// storageSubscribeMessage is a /storage subscribe: the socket's access token and the collections it wants
type storageSubscribeMessage struct {
	AccessToken string               `json:"accessToken"`
	Collections []nsutils.Collection `json:"collections"`
}

// onStorageConnect is API v3's live updates on /storage: a socket subscribes with its access token to the
// collections it wants, and gets each create, update and delete in them as API v3 shows its records
func (ws *Websocket) onStorageConnect(socket *socketio.Socket) {
	siteID := httputils.GetSiteContext(socket.Request().Context()).SiteID
	socket.OnDisconnect(func(string) {
		ws.mu.Lock()
		delete(ws.storage[siteID], socket.ID())
		ws.mu.Unlock()
	})
	onCredentials(ws, socket, "subscribe", func(message storageSubscribeMessage) (any, bool) {
		reply := ws.subscribeToStorage(socket, message.AccessToken, message.Collections)
		return reply, reply["success"] == true
	})
}

// storageSubscriber is a /storage socket, its identity, whose settings it sees records in, and its collections
type storageSubscriber struct {
	socket      *socketio.Socket
	identity    *controldb.SiteIdentity
	collections []nsutils.Collection
}

// subscribeToStorage adds each collection the socket may read (every one when it names none) to those it is sent,
// keeping those it had as Nightscout's rooms do, answering as Nightscout does
func (ws *Websocket) subscribeToStorage(socket *socketio.Socket, accessToken string, collections []nsutils.Collection) map[string]any {
	ctx, identity := ws.identity(socket, accessToken, "")
	if identity == nil {
		return map[string]any{"success": false, "message": "Missing or bad accessToken"}
	}
	if collections == nil {
		collections = storageCollections
	}
	subscribed := []nsutils.Collection{}
	for _, collection := range collections {
		permission := "api:" + string(collection) + ":read"
		if collection == nsutils.Settings {
			permission = "api:settings:admin"
		}
		if slices.Contains(storageCollections, collection) && !slices.Contains(subscribed, collection) && nsutils.HasPermission(*identity, permission) {
			subscribed = append(subscribed, collection)
		}
	}
	if len(subscribed) == 0 {
		return map[string]any{"success": false, "message": "Unauthorized to receive any collection"}
	}
	siteContext := httputils.GetSiteContext(ctx)
	ws.mu.Lock()
	if ws.storage[siteContext.SiteID] == nil {
		ws.storage[siteContext.SiteID] = map[string]*storageSubscriber{}
	}
	subscriber := ws.storage[siteContext.SiteID][socket.ID()]
	if subscriber == nil {
		subscriber = &storageSubscriber{socket: socket}
		ws.storage[siteContext.SiteID][socket.ID()] = subscriber
	}
	subscriber.identity = siteContext.IdentityModel
	for _, collection := range subscribed {
		if !slices.Contains(subscriber.collections, collection) {
			subscriber.collections = append(subscriber.collections, collection)
		}
	}
	ws.mu.Unlock()
	return map[string]any{"success": true, "collections": subscribed}
}

// storageEvents are what API v3 calls each change on /storage
var storageEvents = map[events.Action]string{events.Created: "create", events.Updated: "update", events.Deleted: "delete"}

// sendToStorage sends a change to the /storage sockets subscribed to its collection: a created or updated
// record whole, as API v3 shows it to the socket's identity, and a deleted one by its identifier
func (ws *Websocket) sendToStorage(siteDB *db.SiteDB, siteModel *controldb.Site, event events.Event, collection nsutils.Collection, r recorder) {
	fingerStick := collection == nsutils.Entries && fingerStick(siteDB, event.Model)
	ws.mu.Lock()
	subscribers := []storageSubscriber{}
	for _, subscriber := range ws.storage[event.SiteID] {
		if slices.Contains(subscriber.collections, collection) || (fingerStick && slices.Contains(subscriber.collections, nsutils.Treatments)) {
			subscribers = append(subscribers, *subscriber)
		}
	}
	ws.mu.Unlock()
	for _, subscriber := range subscribers {
		ctx := httputils.SetMgdlUnits(httputils.CreateSiteContext(context.Background(), ws.config, ws.controlDB, siteDB, siteModel, event.SiteID, "", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON))
		siteContext := httputils.GetSiteContext(httputils.SetIdentityModelInSiteContext(ctx, subscriber.identity))
		record, ok := r.Record(event.Model, siteContext)
		if !ok {
			continue
		}
		if slices.Contains(subscriber.collections, collection) {
			emitStorage(subscriber.socket, event.Action, collection, record)
		}
		// A meter reading in no treatment goes to the treatments as a BG Check, as API v3 gives it
		if bgCheck, ok := nscollections.BGCheck(record, siteContext); ok && fingerStick && slices.Contains(subscriber.collections, nsutils.Treatments) {
			emitStorage(subscriber.socket, event.Action, nsutils.Treatments, bgCheck)
		}
	}
}

// emitStorage sends a record's change: a created or updated record whole, and a deleted one by its identifier
func emitStorage(socket *socketio.Socket, action events.Action, collection nsutils.Collection, record nsutils.Document) {
	message := map[string]any{"colName": collection, "doc": record.Fields(true)}
	if action == events.Deleted {
		message = map[string]any{"colName": collection, "identifier": record.Identifier()}
	}
	socket.Emit(storageEvents[action], message)
}

// fingerStick is whether the model is a meter reading in no treatment, which API v3 gives as a BG Check
func fingerStick(siteDB *db.SiteDB, model any) bool {
	reading, ok := model.(*sitedb.GlucoseReading)
	if !ok || !nscollections.FingerStick(reading) {
		return false
	}
	treated, err := siteDB.Queries().GlucoseReadingInTreatment(context.Background(), sitedb.GlucoseReadingInTreatmentParams{ID: &reading.ID})
	return err == nil && !treated
}
