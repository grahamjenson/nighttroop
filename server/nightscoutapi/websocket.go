package nightscoutapi

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nsv1"
	"nighttroop.t1d.nz/server/nightscoutapi/nscollections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	// dataReceivers is the room of sockets that can read, which get every dataUpdate
	dataReceivers = "DataReceivers"
	// maxAuthorizeAttempts bounds the token checks one socket can make; clients authorize once
	maxAuthorizeAttempts = 3
	// maxHistoryHours is how far back the first dataUpdate goes, Nightscout's two days
	maxHistoryHours = 48
	// maxRecords is the most of each collection the first dataUpdate has, as many as a search returns
	maxRecords = 1000
	// maxDeviceStatuses keeps the first dataUpdate small: statuses are large, and clients only show the latest
	maxDeviceStatuses = 10
)

// Websocket is Nightscout's live updates over Socket.IO at /socket.io/. A client emits authorize with its token
// or API secret; one that can read joins DataReceivers, gets a dataUpdate with recent data, then one for each
// change to the site's models, however it was written.
type Websocket struct {
	config      *config.ServerConfig
	controlDB   *db.ControlDB
	collections map[nsutils.Collection]nscollections.CollectionStore

	mu      sync.Mutex
	servers map[string]*socketio.Server              // one per site, so sites share no sessions or rooms
	storage map[string]map[string]*storageSubscriber // each site's /storage subscribers by socket ID
}

func NewWebsocket(controlDB *db.ControlDB, config *config.ServerConfig) *Websocket {
	return &Websocket{config: config, controlDB: controlDB, collections: AllCollections(), servers: map[string]*socketio.Server{}, storage: map[string]map[string]*storageSubscriber{}}
}

func (ws *Websocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.server(httputils.GetSiteContext(r.Context()).SiteID, true).ServeHTTP(w, r)
}

// server is the site's Socket.IO server, made on its first connection when create is set
func (ws *Websocket) server(subdomain string, create bool) *socketio.Server {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if server := ws.servers[subdomain]; server != nil || !create {
		return server
	}
	server := socketio.NewServer(ws.engineConfig())
	main := server.Of(socketio.DefaultNamespace)
	main.OnConnection(func(socket *socketio.Socket) { ws.connected(main, socket) })
	server.Of("/storage").OnConnection(ws.storageConnected)
	server.Of("/alarm").OnConnection(ws.alarmConnected)
	ws.config.Events.Subscribe(subdomain, func(event events.Event) {
		// Only model changes are sent to the site's sockets
		if event.Model != nil {
			ws.changed(main, event)
		}
	})
	ws.servers[subdomain] = server
	return server
}

// engineConfig is the server config's limits on one site's sessions
func (ws *Websocket) engineConfig() engineio.Config {
	return engineio.Config{
		PingInterval:     time.Duration(ws.config.SocketPingIntervalInSeconds) * time.Second,
		PingTimeout:      time.Duration(ws.config.SocketPingTimeoutInSeconds) * time.Second,
		MaxPayload:       ws.config.SocketMaxPayloadBytes,
		MaxSessions:      ws.config.SocketMaxSessionsPerSite,
		MaxBuffered:      ws.config.SocketMaxBufferedPackets,
		MaxBufferedBytes: ws.config.SocketMaxBufferedBytes,
		MaxPacketRate:    ws.config.SocketMaxPacketsPerSecond,
	}
}

// authorization is what authorize answers, as Nightscout does
type authorization struct {
	Read           bool `json:"read"`
	Write          bool `json:"write"`
	WriteTreatment bool `json:"write_treatment"`
}

func (ws *Websocket) connected(main *socketio.Namespace, socket *socketio.Socket) {
	main.Emit("clients", main.Count())
	socket.OnDisconnect(func(string) { main.Emit("clients", main.Count()) })

	attempts := 0
	var auth *authorization // from the latest authorize, which writes need
	var authCtx context.Context
	socket.On("authorize", func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts || !ws.withinReadLimit(socket.Request().Context()) {
			socket.Disconnect()
			return
		}
		var message struct {
			Secret  string   `json:"secret"`
			Token   string   `json:"token"`
			History *float64 `json:"history"`
			Status  bool     `json:"status"`
		}
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		ctx, authorized := ws.authorize(socket.Request().Context(), message.Token, message.Secret)
		auth, authCtx = &authorized, ctx

		socket.Emit("connected")
		if auth.Read {
			socket.Join(dataReceivers)
			hours := float64(maxHistoryHours)
			if message.History != nil && *message.History > 0 && *message.History < hours {
				hours = *message.History
			}
			data := ws.recentData(ctx, time.Duration(hours*float64(time.Hour)))
			if message.Status {
				data["status"] = nsv1.Status()
			}
			socket.Emit("dataUpdate", data)
		}
		if ack != nil {
			ack(auth)
		}
	})

	for event, write := range map[string]dbWriter{"dbAdd": ws.dbAdd, "dbUpdate": ws.dbUpdate, "dbUpdateUnset": ws.dbUpdateUnset, "dbRemove": ws.dbRemove} {
		socket.On(event, func(args []json.RawMessage, ack func(...any)) {
			reply := ws.dbWrite(event, args, auth, authCtx, write)
			if ack != nil {
				ack(reply)
			}
		})
	}

	// loadRetro is the web app asking for the device statuses its retro view shows, which come as a retroUpdate
	socket.On("loadRetro", func(args []json.RawMessage, ack func(...any)) {
		reply := result("success")
		if auth == nil || !auth.Read {
			reply = result("Not permitted")
		} else if !ws.withinReadLimit(authCtx) {
			reply = result("Too many requests")
		}
		if ack != nil {
			ack(reply)
		}
		if reply["result"] == "success" {
			socket.Emit("retroUpdate", map[string]any{"devicestatus": ws.retroDeviceStatuses(authCtx)})
		}
	})
}

// withinReadLimit counts an authorize against the site's GET limit, as a request would: it checks a token
// then reads the recent data, which a client reconnecting in a loop could otherwise ask for without end
func (ws *Websocket) withinReadLimit(ctx context.Context) bool {
	allowed, _ := httputils.WithinRateLimitAll(ws.config.RateLimiters.SiteGETLimiter(httputils.GetSiteContext(ctx).SiteID))
	return allowed
}

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

func result(message string) map[string]string { return map[string]string{"result": message} }

// dbWrite checks a write as Nightscout does, answering why it is refused, then writes it through the collection
// like the REST API, so it is validated, deduplicated and sent on as a dataUpdate the same way
func (ws *Websocket) dbWrite(event string, args []json.RawMessage, auth *authorization, ctx context.Context, write dbWriter) any {
	var message dbMessage
	if len(args) == 0 || json.Unmarshal(args[0], &message) != nil {
		return result("Invalid request")
	}
	collection := nsutils.Collection(message.Collection)
	if !slices.Contains(dbCollections, collection) {
		return result("Wrong collection")
	}
	if auth == nil {
		return result("Not authorized")
	}
	if collection == nsutils.Treatments && !auth.WriteTreatment || collection != nsutils.Treatments && !auth.Write {
		return result("Not permitted")
	}
	// A websocket's events skip the HTTP rate limits, so each write counts against them here
	siteContext := httputils.GetSiteContext(ctx)
	limiters := ws.config.RateLimiters
	if allowed, _ := httputils.WithinRateLimitAll(limiters.SitePOSTLimiter(siteContext.SiteID), limiters.IdentityPOSTLimiter(siteContext.IdentityModel.ID)); !allowed {
		return result("Too many requests")
	}
	var id string
	if event != "dbAdd" && (json.Unmarshal(message.ID, &id) != nil || id == "") {
		return result("Missing or invalid _id")
	}
	return write(ctx, collection, id, message.Data)
}

// dbAdd creates a record, or each of a list, and answers with those stored (the existing one for a duplicate)
func (ws *Websocket) dbAdd(ctx context.Context, collection nsutils.Collection, _ string, data json.RawMessage) any {
	var documents []map[string]any
	if json.Unmarshal(data, &documents) != nil {
		var document map[string]any
		if json.Unmarshal(data, &document) != nil || document == nil {
			return result("Invalid data")
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
		return result("Invalid data")
	}
	if record, err := ws.collections[collection].Patch(ctx, httputils.GetSiteContext(ctx), id, fields, false); err != nil || record == nil {
		slog.Info("Websocket dbUpdate refused", "collection", collection, "error", err)
		return result("Unable to process update")
	}
	return result("success")
}

// dbUpdateUnset removes the given fields from a record, by replacing it with what it is without them
func (ws *Websocket) dbUpdateUnset(ctx context.Context, collection nsutils.Collection, id string, data json.RawMessage) any {
	var fields map[string]any
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return result("Invalid data")
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
		return result("Unable to process update")
	}
	return result("success")
}

// dbRemove deletes a record, softly as the site does
func (ws *Websocket) dbRemove(ctx context.Context, collection nsutils.Collection, id string, _ json.RawMessage) any {
	if err := ws.collections[collection].Delete(ctx, httputils.GetSiteContext(ctx), id, false); err != nil {
		return result("Unable to process removal")
	}
	return result("success")
}

// authorize finds the identity for a token, a JWT or the token given as the API secret, and what it may do
func (ws *Websocket) authorize(ctx context.Context, token, secret string) (context.Context, authorization) {
	ctx, nsIdentity := ws.identity(ctx, token, secret)
	if nsIdentity == nil {
		return ctx, authorization{}
	}

	return ctx, authorization{
		Read:           nsutils.HasPermission(*nsIdentity, "api:*:read"),
		Write:          nsutils.HasPermission(*nsIdentity, "api:*:create"),
		WriteTreatment: nsutils.HasPermission(*nsIdentity, "api:treatments:create"),
	}
}

// identity is the Nightscout identity for a token, a JWT or the token given as the API secret, or nil for none
func (ws *Websocket) identity(ctx context.Context, token, secret string) (context.Context, *nsutils.Identity) {
	identity, err := jwtIdentity(ws.config, ctx, token)
	if err != nil {
		identity, err = legacyTokenIdentity(ws.config, ctx, token)
	}
	if err != nil {
		identity, err = legacyTokenIdentity(ws.config, ctx, secret)
	}
	if err != nil {
		slog.Info("Websocket authorization failed", "error", err)
		return ctx, nil
	}
	nsIdentity, err := nsutils.NighttroopIdentityToNightscoutIdentity(identity.Name, consts.IdentityRole(identity.Role))
	if err != nil {
		return ctx, nil
	}
	return httputils.SetIdentityModelInSiteContext(ctx, identity), nsIdentity
}

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

// changed sends a changed model of the site to its DataReceivers as a delta dataUpdate, when it is a Nightscout
// record. As in Nightscout, a treatment says whether it was updated or removed, and other removals are not sent.
func (ws *Websocket) changed(main *socketio.Namespace, event events.Event) {
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
			ws.send(main, collection, record, event.Action)
			ws.sendToStorage(siteDB, siteModel, event, collection, r)
			return
		}
	}
}

func (ws *Websocket) send(main *socketio.Namespace, collection nsutils.Collection, record nsutils.Record, action events.Action) {
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

// storageCollections are those API v3 serves, which a /storage socket can subscribe to
var storageCollections = []nsutils.Collection{nsutils.Devicestatus, nsutils.Entries, nsutils.Food, nsutils.Profile, nsutils.Settings, nsutils.Treatments}

// storageConnected is API v3's live updates on /storage: a socket subscribes with its access token to the
// collections it wants, and gets each create, update and delete in them as API v3 shows its records
func (ws *Websocket) storageConnected(socket *socketio.Socket) {
	siteID := httputils.GetSiteContext(socket.Request().Context()).SiteID
	socket.OnDisconnect(func(string) {
		ws.mu.Lock()
		delete(ws.storage[siteID], socket.ID())
		ws.mu.Unlock()
	})
	attempts := 0
	socket.On("subscribe", func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts || !ws.withinReadLimit(socket.Request().Context()) {
			socket.Disconnect()
			return
		}
		var message struct {
			AccessToken string               `json:"accessToken"`
			Collections []nsutils.Collection `json:"collections"`
		}
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		reply := ws.subscribe(socket, message.AccessToken, message.Collections)
		if ack != nil {
			ack(reply)
		}
	})
}

// storageSubscriber is a /storage socket, its identity, whose settings it sees records in, and its collections
type storageSubscriber struct {
	socket      *socketio.Socket
	identity    *controldb.SiteIdentity
	collections []nsutils.Collection
}

// subscribe adds each collection the socket may read (every one when it names none) to those it is sent,
// keeping those it had as Nightscout's rooms do, answering as Nightscout does
func (ws *Websocket) subscribe(socket *socketio.Socket, accessToken string, collections []nsutils.Collection) map[string]any {
	ctx, identity := ws.identity(socket.Request().Context(), accessToken, "")
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
	ws.mu.Lock()
	subscribers := []storageSubscriber{}
	for _, subscriber := range ws.storage[event.SiteID] {
		if slices.Contains(subscriber.collections, collection) {
			subscribers = append(subscribers, *subscriber)
		}
	}
	ws.mu.Unlock()
	for _, subscriber := range subscribers {
		ctx := httputils.CreateSiteContext(context.Background(), ws.config, ws.controlDB, siteDB, siteModel, event.SiteID, "", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		record, _ := r.Record(event.Model, httputils.GetSiteContext(httputils.SetIdentityModelInSiteContext(ctx, subscriber.identity)))
		message := map[string]any{"colName": collection, "doc": record.ToMap(true)}
		if event.Action == events.Deleted {
			message = map[string]any{"colName": collection, "identifier": record.Identifier()}
		}
		subscriber.socket.Emit(storageEvents[event.Action], message)
	}
}

// alarmReceivers is the room of /alarm sockets that can read, which the site's alarms will go to
const alarmReceivers = "AlarmReceivers"

// alarmConnected is API v3's alarms on /alarm: a socket subscribes with its access token, or as the web app does
// with its API secret or JWT, and joins AlarmReceivers when it can read. The site has no alarms to send yet.
func (ws *Websocket) alarmConnected(socket *socketio.Socket) {
	attempts := 0
	socket.On("subscribe", func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts || !ws.withinReadLimit(socket.Request().Context()) {
			socket.Disconnect()
			return
		}
		var message struct {
			AccessToken string `json:"accessToken"`
			JWTToken    string `json:"jwtToken"`
			Secret      string `json:"secret"`
		}
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		reply := ws.subscribeToAlarms(socket, message.AccessToken, message.JWTToken, message.Secret)
		if ack != nil {
			ack(reply)
		}
	})
}

// subscribeToAlarms answers as Nightscout does: any credential given must be good, and the web app, which may
// have none yet, is told what it may do. Acks are ignored until the site has alarms.
func (ws *Websocket) subscribeToAlarms(socket *socketio.Socket, accessToken, jwtToken, secret string) map[string]any {
	token := cmp.Or(accessToken, jwtToken)
	if secret == "null" {
		secret = ""
	}
	_, identity := ws.identity(socket.Request().Context(), token, secret)
	if identity == nil && token+secret != "" {
		return map[string]any{"success": false, "message": "Missing or bad accessToken"}
	}
	read := identity != nil && nsutils.HasPermission(*identity, "api:*:read")
	if read {
		socket.Join(alarmReceivers)
	} else {
		socket.Leave(alarmReceivers)
	}
	reply := map[string]any{"success": true, "message": "Subscribed for alarms"}
	if accessToken == "" {
		reply["read"], reply["ack"] = read, false
	}
	return reply
}
