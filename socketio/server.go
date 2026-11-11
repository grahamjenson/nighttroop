package socketio

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/socketio/engineio"
)

// EventHandler handles an event's arguments; ack is nil unless the client asked for an acknowledgement,
// and answers it at most once
type EventHandler func(args []json.RawMessage, ack func(args ...any))

// Server is a Socket.IO server, mounted at a path like /socket.io/
type Server struct {
	engine *engineio.Server

	mu         sync.RWMutex
	namespaces map[string]*Namespace
}

// NewServer starts a server with the default namespace "/"
func NewServer(config engineio.Config) *Server {
	s := &Server{namespaces: map[string]*Namespace{}}
	s.engine = engineio.NewServer(config, s.onEngineConnection)
	s.Of(DefaultNamespace)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.engine.ServeHTTP(w, r) }

// Close disconnects every client
func (s *Server) Close() { s.engine.Close() }

// Count is how many Engine.IO sessions are open, connected to a namespace or not
func (s *Server) Count() int { return s.engine.Count() }

// Of is the namespace with this name, made on first use; clients can only join namespaces made here
func (s *Server) Of(name string) *Namespace {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n, ok := s.namespaces[name]; ok {
		return n
	}
	n := &Namespace{name: name, sockets: map[string]*Socket{}, rooms: map[string]map[string]*Socket{}}
	s.namespaces[name] = n
	return n
}

func (s *Server) namespace(name string) *Namespace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.namespaces[name]
}

// Namespace is a channel clients connect to, with its own sockets, rooms and handlers
type Namespace struct {
	name string

	mu           sync.RWMutex
	sockets      map[string]*Socket
	rooms        map[string]map[string]*Socket
	onConnection func(*Socket)
	authorize    func(s *Socket, auth json.RawMessage) error
}

// OnConnection sets the handler for each socket that connects, where it adds its event handlers
func (n *Namespace) OnConnection(handler func(*Socket)) {
	n.mu.Lock()
	n.onConnection = handler
	n.mu.Unlock()
}

// Use sets the check each socket must pass to connect; its error is sent to the client.
// auth is the client's CONNECT payload, which Socket.IO v2 clients do not send.
func (n *Namespace) Use(authorize func(s *Socket, auth json.RawMessage) error) {
	n.mu.Lock()
	n.authorize = authorize
	n.mu.Unlock()
}

// Emit sends an event to every socket in the namespace
func (n *Namespace) Emit(event string, args ...any) error {
	return n.emit(n.allSockets(), event, args)
}

// EmitTo sends an event to every socket in a room
func (n *Namespace) EmitTo(room string, event string, args ...any) error {
	n.mu.RLock()
	sockets := make([]*Socket, 0, len(n.rooms[room]))
	for _, socket := range n.rooms[room] {
		sockets = append(sockets, socket)
	}
	n.mu.RUnlock()
	return n.emit(sockets, event, args)
}

// Count is how many sockets are connected
func (n *Namespace) Count() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.sockets)
}

func (n *Namespace) allSockets() []*Socket {
	n.mu.RLock()
	defer n.mu.RUnlock()
	sockets := make([]*Socket, 0, len(n.sockets))
	for _, socket := range n.sockets {
		sockets = append(sockets, socket)
	}
	return sockets
}

// emit encodes the event once, and sends it to sockets that are still open
func (n *Namespace) emit(sockets []*Socket, event string, args []any) error {
	data, err := EventData(event, args...)
	if err != nil {
		return err
	}
	encoded := Packet{Type: Event, Namespace: n.name, Data: data}.Encode()
	for _, socket := range sockets {
		socket.conn.session.Send(encoded)
	}
	return nil
}

func (n *Namespace) add(socket *Socket) {
	n.mu.Lock()
	n.sockets[socket.id] = socket
	n.mu.Unlock()
}

func (n *Namespace) remove(socket *Socket) {
	n.mu.Lock()
	delete(n.sockets, socket.id)
	for room, members := range n.rooms {
		delete(members, socket.id)
		if len(members) == 0 {
			delete(n.rooms, room)
		}
	}
	n.mu.Unlock()
}

// Socket is one client connected to one namespace
type Socket struct {
	id        string
	namespace *Namespace
	conn      *conn

	mu           sync.Mutex
	handlers     map[string]EventHandler
	onDisconnect func(reason string)
	connected    bool
}

func (s *Socket) ID() string { return s.id }

func (s *Socket) Namespace() *Namespace { return s.namespace }

// Request is the Engine.IO handshake request, for its headers, query, remote address and context values
func (s *Socket) Request() *http.Request { return s.conn.session.Request() }

// On sets the handler for an event from the client; handlers run in order, one at a time
func (s *Socket) On(event string, handler EventHandler) {
	s.mu.Lock()
	s.handlers[event] = handler
	s.mu.Unlock()
}

// OnDisconnect sets the handler called once when the socket disconnects
func (s *Socket) OnDisconnect(handler func(reason string)) {
	s.mu.Lock()
	s.onDisconnect = handler
	s.mu.Unlock()
}

var ErrDisconnected = errors.New("socketio: socket disconnected")

// Connected is whether the socket is still connected to its namespace
func (s *Socket) Connected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

// Emit sends an event to this socket
func (s *Socket) Emit(event string, args ...any) error {
	if !s.Connected() {
		return ErrDisconnected
	}
	data, err := EventData(event, args...)
	if err != nil {
		return err
	}
	return s.send(Packet{Type: Event, Data: data})
}

// Join adds the socket to a room in its namespace
func (s *Socket) Join(room string) {
	n := s.namespace
	n.mu.Lock()
	if _, ok := n.sockets[s.id]; ok {
		if n.rooms[room] == nil {
			n.rooms[room] = map[string]*Socket{}
		}
		n.rooms[room][s.id] = s
	}
	n.mu.Unlock()
}

// Leave takes the socket out of a room
func (s *Socket) Leave(room string) {
	n := s.namespace
	n.mu.Lock()
	delete(n.rooms[room], s.id)
	if len(n.rooms[room]) == 0 {
		delete(n.rooms, room)
	}
	n.mu.Unlock()
}

// Disconnect tells the client the socket is closed and removes it from its namespace
func (s *Socket) Disconnect() {
	s.send(Packet{Type: Disconnect})
	s.conn.disconnect(s, "server namespace disconnect")
}

// Close disconnects the socket and, when its client has no other socket, closes its session, so it holds none of the
// server's sessions
func (s *Socket) Close() {
	s.Disconnect()
	s.conn.closeIfEmpty("server close")
}

func (s *Socket) send(p Packet) error {
	p.Namespace = s.namespace.name
	return s.conn.session.Send(p.Encode())
}

// conn is one Engine.IO session and the sockets it has in each namespace
type conn struct {
	server  *Server
	session *engineio.Session

	mu      sync.Mutex
	sockets map[string]*Socket
}

// connectTimeout is how long a client has to connect to a namespace before its session is closed, Socket.IO's
// connectTimeout, so a client cannot hold one of the server's sessions without using it
var connectTimeout = 45 * time.Second

func (s *Server) onEngineConnection(session *engineio.Session) {
	c := &conn{server: s, session: session, sockets: map[string]*Socket{}}
	session.OnMessage(c.onMessage)
	session.OnClose(c.onClose)
	// Socket.IO v2 clients are connected to the default namespace without asking
	if session.Version() == engineio.V3 {
		c.connect(DefaultNamespace, nil)
	}
	time.AfterFunc(connectTimeout, func() { c.closeIfEmpty("connect timeout") })
}

// closeIfEmpty closes the session when it has no socket in any namespace
func (c *conn) closeIfEmpty(reason string) {
	c.mu.Lock()
	empty := len(c.sockets) == 0
	c.mu.Unlock()
	if empty {
		c.session.Close(reason)
	}
}

func (c *conn) onMessage(data string) {
	p, err := DecodePacket(data)
	if err != nil {
		c.session.Close("bad packet")
		return
	}
	c.mu.Lock()
	socket := c.sockets[p.Namespace]
	none := len(c.sockets) == 0
	c.mu.Unlock()

	switch p.Type {
	case Connect:
		if socket == nil {
			c.connect(p.Namespace, p.Data)
		}
	case Disconnect:
		if socket != nil {
			c.disconnect(socket, "client namespace disconnect")
		}
	case Event:
		if socket != nil {
			socket.handle(p)
		} else if none {
			// A client must connect to a namespace before anything else
			c.session.Close("event before connect")
		}
	}
}

func (c *conn) onClose(reason string) {
	c.mu.Lock()
	sockets := make([]*Socket, 0, len(c.sockets))
	for _, socket := range c.sockets {
		sockets = append(sockets, socket)
	}
	c.mu.Unlock()
	for _, socket := range sockets {
		c.disconnect(socket, reason)
	}
}

var errInvalidNamespace = errors.New("Invalid namespace")

func (c *conn) connect(name string, auth json.RawMessage) {
	n := c.server.namespace(name)
	if n == nil {
		c.connectError(name, errInvalidNamespace)
		return
	}
	id := make([]byte, 15)
	rand.Read(id)
	socket := &Socket{id: base64.RawURLEncoding.EncodeToString(id), namespace: n, conn: c, handlers: map[string]EventHandler{}, connected: true}

	n.mu.RLock()
	authorize, onConnection := n.authorize, n.onConnection
	n.mu.RUnlock()
	if authorize != nil {
		if err := authorize(socket, auth); err != nil {
			c.connectError(name, err)
			return
		}
	}

	c.mu.Lock()
	c.sockets[name] = socket
	c.mu.Unlock()
	n.add(socket)

	// v5 clients get the socket's id; v2 clients get a bare CONNECT
	var data json.RawMessage
	if c.session.Version() == engineio.V4 {
		data, _ = json.Marshal(map[string]string{"sid": socket.id})
	}
	socket.send(Packet{Type: Connect, Data: data})
	if onConnection != nil {
		onConnection(socket)
	}
}

// connectError refuses a namespace: v5 sends {"message": ...}, v2 a bare string
func (c *conn) connectError(name string, err error) {
	var data json.RawMessage
	if c.session.Version() == engineio.V4 {
		data, _ = json.Marshal(map[string]string{"message": err.Error()})
	} else {
		data, _ = json.Marshal(err.Error())
	}
	c.session.Send(Packet{Type: ConnectError, Namespace: name, Data: data}.Encode())
}

func (c *conn) disconnect(socket *Socket, reason string) {
	c.mu.Lock()
	if c.sockets[socket.namespace.name] == socket {
		delete(c.sockets, socket.namespace.name)
	}
	c.mu.Unlock()
	socket.namespace.remove(socket)

	socket.mu.Lock()
	wasConnected := socket.connected
	socket.connected = false
	onDisconnect := socket.onDisconnect
	socket.mu.Unlock()
	if wasConnected && onDisconnect != nil {
		onDisconnect(reason)
	}
}

func (s *Socket) handle(p Packet) {
	event, args, err := ParseEvent(p.Data)
	if err != nil {
		s.conn.session.Close("bad event")
		return
	}
	s.mu.Lock()
	handler := s.handlers[event]
	s.mu.Unlock()
	if handler == nil {
		return
	}
	var ack func(args ...any)
	if p.ID != nil {
		var once sync.Once
		id := *p.ID
		ack = func(args ...any) {
			once.Do(func() {
				data, err := json.Marshal(append([]any{}, args...))
				if err == nil {
					s.send(Packet{Type: Ack, ID: &id, Data: data})
				}
			})
		}
	}
	handler(args, ack)
}
