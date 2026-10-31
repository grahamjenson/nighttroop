package nswebsockets

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nscollections"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// The websocket is split by what it serves: the default namespace's handlers in websocket_main.go, its db writes in
// websocket_db.go, the data it sends in websocket_data.go, and the /storage and /alarm namespaces in their own files.

const (
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
	// jwtIdentity and tokenIdentity find the identity of a JWT and of an access token, as the Nightscout API does
	jwtIdentity, tokenIdentity IdentityFinder

	mu      sync.Mutex
	servers map[string]*socketio.Server              // one per site, so sites share no sessions or rooms
	storage map[string]map[string]*storageSubscriber // each site's /storage subscribers by socket ID
}

// IdentityFinder finds the identity a token belongs to, or errors when it belongs to none
type IdentityFinder func(config *config.ServerConfig, ctx context.Context, token string) (*controldb.SiteIdentity, error)

// NewWebsocket serves the Nightscout API's collections, finding identities as the API does
func NewWebsocket(controlDB *db.ControlDB, config *config.ServerConfig, collections map[nsutils.Collection]nscollections.CollectionStore, jwtIdentity, tokenIdentity IdentityFinder) *Websocket {
	return &Websocket{config: config, controlDB: controlDB, collections: collections, jwtIdentity: jwtIdentity, tokenIdentity: tokenIdentity, servers: map[string]*socketio.Server{}, storage: map[string]map[string]*storageSubscriber{}}
}

func (ws *Websocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.siteServer(httputils.GetSiteContext(r.Context()).SiteID, true).ServeHTTP(w, r)
}

// siteServer is the site's Socket.IO server, made on its first connection when create is set
func (ws *Websocket) siteServer(siteID string, create bool) *socketio.Server {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if server := ws.servers[siteID]; server != nil || !create {
		return server
	}
	server := socketio.NewServer(ws.engineConfig())
	main := server.Of(socketio.DefaultNamespace)
	main.OnConnection(func(socket *socketio.Socket) { ws.onConnect(main, socket) })
	server.Of("/storage").OnConnection(ws.onStorageConnect)
	alarm := server.Of("/alarm")
	alarm.OnConnection(ws.onAlarmConnect)
	ws.config.Events.Subscribe(siteID, func(event events.Event) {
		// Model changes are sent to the site's sockets, and an identity's alarms to its /alarm sockets
		if event.Model != nil {
			ws.broadcastChange(main, event)
		}
		if checked, ok := event.Attributes.(alarms.Checked); ok && event.Name == events.Alarm {
			ws.sendAlarms(alarm, checked)
		}
	})
	ws.servers[siteID] = server
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

// acknowledge answers an event with reply, when the client asked for an answer
func acknowledge(ack func(...any), reply any) {
	if ack != nil {
		ack(reply)
	}
}

// onCredentials handles a socket's event that checks credentials, as authorize and the subscribes do. A client
// sends one, so past maxAuthorizeAttempts, or over the site's read limit, the socket is disconnected; otherwise its
// first argument, decoded as M, is handled and the reply is the ack.
func onCredentials[M any](ws *Websocket, socket *socketio.Socket, event string, handle func(message M) any) {
	attempts := 0
	socket.On(event, func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts || !ws.withinReadLimit(socket.Request().Context()) {
			socket.Disconnect()
			return
		}
		var message M
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		acknowledge(ack, handle(message))
	})
}

// withinReadLimit counts an authorize against the site's GET limit, as a request would: it checks a token
// then reads the recent data, which a client reconnecting in a loop could otherwise ask for without end
func (ws *Websocket) withinReadLimit(ctx context.Context) bool {
	allowed, _ := httputils.WithinRateLimitAll(ws.config.RateLimiters.SiteGETLimiter(httputils.GetSiteContext(ctx).SiteID))
	return allowed
}

// authorization is what authorize answers, as Nightscout does
type authorization struct {
	Read           bool `json:"read"`
	Write          bool `json:"write"`
	WriteTreatment bool `json:"write_treatment"`
}

// authorize finds the identity for a token, a JWT or the token given as the API secret, and what it may do
func (ws *Websocket) authorize(r *http.Request, token, secret string) (context.Context, authorization) {
	ctx, nsIdentity := ws.identity(r, token, secret)
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
func (ws *Websocket) identity(r *http.Request, token, secret string) (context.Context, *nsutils.Identity) {
	ctx := r.Context()
	identity, err := ws.jwtIdentity(ws.config, ctx, token)
	if err != nil {
		identity, err = ws.tokenIdentity(ws.config, ctx, token)
	}
	if err != nil {
		identity, err = ws.tokenIdentity(ws.config, ctx, secret)
	}
	if err != nil {
		slog.InfoContext(r.Context(), "Websocket authorization failed", "error", err)
		return ctx, nil
	}
	nsIdentity, err := nsutils.NighttroopIdentityToNightscoutIdentity(identity.Name, consts.IdentityRole(identity.Role))
	if err != nil {
		return ctx, nil
	}
	httputils.RecordAuthClient(r, identity)
	return httputils.SetIdentityModelInSiteContext(ctx, identity), nsIdentity
}
