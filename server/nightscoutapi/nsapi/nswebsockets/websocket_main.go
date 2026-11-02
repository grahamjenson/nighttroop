package nswebsockets

import (
	"context"
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi/nsapi/nsv1"
	"nighttroop.t1d.nz/socketio"
)

// dataReceivers is the room of sockets that can read, which get every dataUpdate
const dataReceivers = "DataReceivers"

// mainSocket is a socket on the default namespace and what its latest authorize allowed, which writes and
// loadRetro need
type mainSocket struct {
	socket *socketio.Socket
	auth   *authorization
	ctx    context.Context // the site context with the authorized identity
}

// authorizeMessage is what a client authorizes with: its API secret or token, how many hours of history its first
// dataUpdate has, and whether it has Nightscout's status
type authorizeMessage struct {
	Secret  string   `json:"secret"`
	Token   string   `json:"token"`
	History *float64 `json:"history"`
	Status  bool     `json:"status"`
}

// onConnect tells the site's sockets how many are connected, and handles a new socket's authorize, db writes and
// loadRetro
func (ws *Websocket) onConnect(main *socketio.Namespace, socket *socketio.Socket) {
	main.Emit("clients", main.Count())
	socket.OnDisconnect(func(string) { main.Emit("clients", main.Count()) })

	s := &mainSocket{socket: socket}
	onCredentials(ws, socket, "authorize", func(message authorizeMessage) any { return ws.handleAuthorize(s, message) })
	for event, write := range map[string]dbWriter{"dbAdd": ws.dbAdd, "dbUpdate": ws.dbUpdate, "dbUpdateUnset": ws.dbUpdateUnset, "dbRemove": ws.dbRemove} {
		socket.On(event, func(args []json.RawMessage, ack func(...any)) {
			acknowledge(ack, ws.handleDBWrite(event, args, s.auth, s.ctx, write))
		})
	}
	socket.On("loadRetro", func(_ []json.RawMessage, ack func(...any)) { ws.handleLoadRetro(s, ack) })
}

// handleAuthorize keeps what the socket may do, and sends a socket that can read the recent data
func (ws *Websocket) handleAuthorize(s *mainSocket, message authorizeMessage) any {
	ctx, authorized := ws.authorize(s.socket, message.Token, message.Secret)
	s.auth, s.ctx = &authorized, ctx

	s.socket.Emit("connected")
	if authorized.Read {
		s.socket.Join(dataReceivers)
		s.socket.Emit("dataUpdate", ws.firstDataUpdate(ctx, message))
	}
	return s.auth
}

// firstDataUpdate is the recent data over the history the client asked for, at most maxHistoryHours, with
// Nightscout's status when it asked for it
func (ws *Websocket) firstDataUpdate(ctx context.Context, message authorizeMessage) map[string]any {
	hours := float64(maxHistoryHours)
	if message.History != nil && *message.History > 0 && *message.History < hours {
		hours = *message.History
	}
	data := ws.recentData(ctx, time.Duration(hours*float64(time.Hour)))
	if message.Status {
		data["status"] = nsv1.Status()
	}
	return data
}

// handleLoadRetro is the web app asking for the device statuses its retro view shows, which come as a retroUpdate
func (ws *Websocket) handleLoadRetro(s *mainSocket, ack func(...any)) {
	reply := ackResult("success")
	if s.auth == nil || !s.auth.Read {
		reply = ackResult("Not permitted")
	} else if !ws.withinReadLimit(s.ctx) {
		reply = ackResult("Too many requests")
	}
	acknowledge(ack, reply)
	if reply["result"] == "success" {
		s.socket.Emit("retroUpdate", map[string]any{"devicestatus": ws.retroDeviceStatuses(s.ctx)})
	}
}
