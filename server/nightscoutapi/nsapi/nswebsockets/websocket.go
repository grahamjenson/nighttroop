package nswebsockets

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
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
	// maxAuthorizeAttempts bounds the token checks one socket can make; clients authorize once, though the Nightscout
	// web app subscribes first with no credentials, then with its API secret, which may be mistyped
	maxAuthorizeAttempts = 3
	// maxHistoryHours is how far back the first dataUpdate goes, Nightscout's two days
	maxHistoryHours = 48
	// maxRecords is the most of each collection the first dataUpdate has, as many as a search returns
	maxRecords = 1000
	// maxDeviceStatuses keeps the first dataUpdate small: statuses are large, and clients only show the latest
	maxDeviceStatuses = 10
)

// authorizeWithin is how long a socket has to be given access to the site before it is closed. Clients authorize as
// they connect, so one that has not, or was given nothing, is only holding one of the site's SocketMaxSessionsPerSite
// sessions from its followers.
var authorizeWithin = 30 * time.Second

// refusedCloseAfter is how long a socket refused for good has for its answer and disconnect to reach its client before
// its session is closed
var refusedCloseAfter = time.Second

// identityRecheck is how often a socket's identity is read again, so one deleted, disabled or given new credentials
// stops being sent the site's data
var identityRecheck = time.Minute

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
	servers map[string]*siteSockets                  // one per site, so sites share no sessions or rooms
	storage map[string]map[string]*storageSubscriber // each site's /storage subscribers by socket ID
}

// IdentityFinder finds the identity a token belongs to, or errors when it belongs to none
type IdentityFinder func(config *config.ServerConfig, ctx context.Context, token string) (*controldb.SiteIdentity, error)

// NewWebsocket serves the Nightscout API's collections, finding identities as the API does
func NewWebsocket(controlDB *db.ControlDB, config *config.ServerConfig, collections map[nsutils.Collection]nscollections.CollectionStore, jwtIdentity, tokenIdentity IdentityFinder) *Websocket {
	return &Websocket{config: config, controlDB: controlDB, collections: collections, jwtIdentity: jwtIdentity, tokenIdentity: tokenIdentity, servers: map[string]*siteSockets{}, storage: map[string]map[string]*storageSubscriber{}}
}

func (ws *Websocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := ws.siteServer(httputils.GetSiteContext(r.Context()).SiteID, true)
	defer func() {
		ws.mu.Lock()
		server.serving--
		ws.mu.Unlock()
	}()
	server.ServeHTTP(w, r)
}

// siteSockets is a site's Socket.IO server, with its subscription to the site's events and how many requests it is
// serving, as one being served may be about to open a session
type siteSockets struct {
	*socketio.Server
	unsubscribe func()
	serving     int
}

// idleServerSweep is how often each site's server with no sessions, and no request being served, is freed with its
// subscription, so the servers of sites no longer connected to, or deleted or disabled, are not kept until restart
var idleServerSweep = 5 * time.Minute

// siteServer is the site's Socket.IO server, made on its first connection when create is set, which then counts as a
// request being served until ServeHTTP is done with it
func (ws *Websocket) siteServer(siteID string, create bool) *siteSockets {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if site := ws.servers[siteID]; site != nil || !create {
		if site != nil && create {
			site.serving++
		}
		return site
	}
	server := socketio.NewServer(ws.engineConfig())
	main := server.Of(socketio.DefaultNamespace)
	main.OnConnection(func(socket *socketio.Socket) { ws.onConnect(main, socket) })
	server.Of("/storage").OnConnection(ws.onStorageConnect)
	alarm := server.Of("/alarm")
	alarm.OnConnection(ws.onAlarmConnect)
	unsubscribe := ws.config.Events.Subscribe(siteID, func(event events.Event) {
		// Model changes are sent to the site's sockets, its announcements and an identity's alarms to its /alarm sockets
		if event.Model != nil {
			ws.broadcastChange(main, alarm, event)
			sendAnnouncement(alarm, event)
		}
		if checked, ok := event.Attributes.(alarms.Checked); ok && event.Name == events.Alarm {
			ws.sendAlarms(alarm, checked)
		}
		// A site deleted or disabled is no longer served, so its sessions are closed and the sweep frees its server
		if event.Name == events.SiteDeleted || event.Name == events.SiteDisabled {
			go server.Close()
		}
	})
	if len(ws.servers) == 0 {
		time.AfterFunc(idleServerSweep, ws.sweepServers)
	}
	site := &siteSockets{Server: server, unsubscribe: unsubscribe, serving: 1}
	ws.servers[siteID] = site
	return site
}

// sweepServers frees each site's server with no sessions and no request being served, then sweeps again while any
// are left. Requests get a site's server under ws.mu, so none can open a session on one being freed.
func (ws *Websocket) sweepServers() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for siteID, site := range ws.servers {
		if site.serving == 0 && site.Count() == 0 {
			site.unsubscribe()
			delete(ws.servers, siteID)
		}
	}
	if len(ws.servers) > 0 {
		time.AfterFunc(idleServerSweep, ws.sweepServers)
	}
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
		// One IP may hold only a few of the site's sessions until they are given access, leaving the rest to its followers
		ClientKey:       func(r *http.Request) string { return config.IPKey(httputils.GetSiteContext(r.Context()).IPAddress) },
		MaxUnauthorized: ws.config.SocketMaxUnauthPerIP,
	}
}

// acknowledge answers an event with reply, when the client asked for an answer
func acknowledge(ack func(...any), reply any) {
	if ack != nil {
		ack(reply)
	}
}

// onCredentials handles a socket's event that checks credentials, as authorize and the subscribes do. A client
// sends few, so past maxAuthorizeAttempts, or over its limit, the socket is refused; otherwise its first argument,
// decoded as M, is handled and the reply is the ack, unless handling it disconnected the socket. A socket that
// handling has not granted access within authorizeWithin is closed, and one that handling disconnected or refused
// maxAuthorizeAttempts times is refused.
func onCredentials[M any](ws *Websocket, socket *socketio.Socket, event string, handle func(message M) (reply any, granted bool)) {
	attempts := 0
	var granted atomic.Bool
	time.AfterFunc(authorizeWithin, func() {
		if !granted.Load() && socket.Connected() {
			socket.Close()
		}
	})
	// refuse disconnects the socket now, after any answer, and closes its session once both have reached the client
	refuse := func() {
		if socket.Connected() {
			socket.Disconnect()
		}
		time.AfterFunc(refusedCloseAfter, socket.Close)
	}
	socket.On(event, func(args []json.RawMessage, ack func(...any)) {
		attempts++
		if attempts > maxAuthorizeAttempts || !ws.withinAuthorizeLimit(socket, granted.Load()) {
			refuse()
			return
		}
		var message M
		if len(args) > 0 {
			json.Unmarshal(args[0], &message)
		}
		reply, ok := handle(message)
		if socket.Connected() {
			acknowledge(ack, reply)
		}
		if ok {
			granted.Store(true)
			socket.Authorize()
		} else if attempts == maxAuthorizeAttempts || !socket.Connected() {
			refuse()
		}
	})
}

// withinAuthorizeLimit counts an authorize against the IP's unauthenticated budget on the site until the socket is
// given access, as a request that fails auth is, so one IP cannot spend the site's own limit; after, against the
// site's GET limit, as it reads the recent data, which a client reconnecting in a loop could otherwise ask for without end
func (ws *Websocket) withinAuthorizeLimit(socket *socketio.Socket, granted bool) bool {
	if granted {
		return ws.withinReadLimit(socket.Request().Context())
	}
	allowed, _ := httputils.SiteUnauthWithinRateLimit(ws.config, socket.Request())
	return allowed
}

// withinReadLimit counts a socket's read against the site's GET limit, as a request would
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

// authorize finds the identity for a token, a JWT or the token given as the API secret, and what it may do. It is
// false when there is no such identity.
func (ws *Websocket) authorize(socket *socketio.Socket, token, secret string) (context.Context, authorization, bool) {
	ctx, nsIdentity := ws.identity(socket, token, secret)
	if nsIdentity == nil {
		return ctx, authorization{}, false
	}

	return ctx, authorization{
		Read:           nsutils.HasPermission(*nsIdentity, "api:*:read"),
		Write:          nsutils.HasPermission(*nsIdentity, "api:*:create"),
		WriteTreatment: nsutils.HasPermission(*nsIdentity, "api:treatments:create"),
	}, true
}

// identity is the Nightscout identity for a token, a JWT or the token given as the API secret, or nil for none. The
// socket is disconnected once that identity is revoked.
func (ws *Websocket) identity(socket *socketio.Socket, token, secret string) (context.Context, *nsutils.Identity) {
	r := socket.Request()
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
	nsIdentity, err := nsutils.NighttroopIdentityToNightscoutIdentity(identity.Name, consts.IdentityRole(identity.Role), identity.Client)
	if err != nil {
		return ctx, nil
	}
	httputils.RecordAuthClient(r, identity)
	ws.watchIdentity(socket, identity)
	return httputils.SetIdentityModelInSiteContext(ctx, identity), nsIdentity
}

// watchIdentity disconnects the socket once identity, which it authorized as, is revoked, checking every
// identityRecheck while it is connected. A socket is not authenticated again on each event, so would otherwise keep
// reading and writing the site's data after its identity was disabled, deleted or given new credentials.
func (ws *Websocket) watchIdentity(socket *socketio.Socket, identity *controldb.SiteIdentity) {
	var check func()
	check = func() {
		if !socket.Connected() {
			return
		}
		if httputils.IdentityRevoked(context.Background(), ws.controlDB, identity) {
			socket.Disconnect()
			return
		}
		time.AfterFunc(identityRecheck, check)
	}
	time.AfterFunc(identityRecheck, check)
}
