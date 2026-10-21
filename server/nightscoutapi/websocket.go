package nightscoutapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"sync"
	"time"

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
	// maxSocketsPerSite stops one site holding every connection
	maxSocketsPerSite = 100
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
// or API secret; one that can read joins DataReceivers, gets a dataUpdate with recent data, then one for each write.
type Websocket struct {
	config *config.ServerConfig
	store  *NightTroopAPIStore

	mu      sync.Mutex
	servers map[string]*socketio.Server // one per site, so sites share no sessions or rooms
}

func NewWebsocket(config *config.ServerConfig, store *NightTroopAPIStore) *Websocket {
	ws := &Websocket{config: config, store: store, servers: map[string]*socketio.Server{}}
	store.onChange = ws.changed
	return ws
}

func (ws *Websocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.server(httputils.GetSiteContext(r.Context()).Subdomain, true).ServeHTTP(w, r)
}

// server is the site's Socket.IO server, made on its first connection when create is set
func (ws *Websocket) server(subdomain string, create bool) *socketio.Server {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if server := ws.servers[subdomain]; server != nil || !create {
		return server
	}
	server := socketio.NewServer(engineio.Config{MaxSessions: maxSocketsPerSite})
	main := server.Of(socketio.DefaultNamespace)
	main.OnConnection(func(socket *socketio.Socket) { ws.connected(main, socket) })
	ws.servers[subdomain] = server
	return server
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
	socket.On("authorize", func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts {
			socket.Disconnect()
			return
		}
		var message struct {
			Secret  string   `json:"secret"`
			Token   string   `json:"token"`
			History *float64 `json:"history"`
		}
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		ctx, auth := ws.authorize(socket.Request().Context(), message.Token, message.Secret)

		socket.Emit("connected")
		if auth.Read {
			socket.Join(dataReceivers)
			hours := float64(maxHistoryHours)
			if message.History != nil && *message.History > 0 && *message.History < hours {
				hours = *message.History
			}
			socket.Emit("dataUpdate", ws.recentData(ctx, time.Duration(hours*float64(time.Hour))))
		}
		if ack != nil {
			ack(auth)
		}
	})
}

// authorize finds the identity for a token, a JWT or the token given as the API secret, and what it may do
func (ws *Websocket) authorize(ctx context.Context, token, secret string) (context.Context, authorization) {
	identity, err := jwtIdentity(ws.config, ctx, token)
	if err != nil {
		identity, err = legacyTokenIdentity(ws.config, ctx, token)
	}
	if err != nil {
		identity, err = legacyTokenIdentity(ws.config, ctx, secret)
	}
	if err != nil {
		slog.Info("Websocket authorization failed", "error", err)
		return ctx, authorization{}
	}
	nsIdentity, err := NighttroopIdentityToNightscoutIdentity(identity.Name, consts.IdentityRole(identity.Role))
	if err != nil {
		return ctx, authorization{}
	}
	auth := &NightTroopAPIAuth{}
	return httputils.SetIdentityModelInSiteContext(ctx, identity), authorization{
		Read:           auth.HasPermission(*nsIdentity, "api:*:read"),
		Write:          auth.HasPermission(*nsIdentity, "api:*:create"),
		WriteTreatment: auth.HasPermission(*nsIdentity, "api:treatments:create"),
	}
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
		records, err := ws.store.Search(ctx, collection, query)
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

// changed sends a written record to the site's DataReceivers as a delta dataUpdate
func (ws *Websocket) changed(ctx context.Context, collection nsutils.Collection, record nsutils.Record) {
	server := ws.server(httputils.GetSiteContext(ctx).Subdomain, false)
	key, value, ok := clientData(collection, record)
	if server == nil || !ok {
		return
	}
	server.Of(socketio.DefaultNamespace).EmitTo(dataReceivers, "dataUpdate", map[string]any{
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
