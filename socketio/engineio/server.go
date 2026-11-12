package engineio

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Config is how a Server runs its sessions; zero values take the defaults
type Config struct {
	PingInterval     time.Duration // how often the server (v4) or client (v3) pings, 25s by default
	PingTimeout      time.Duration // how long to wait for the answer, 20s by default
	MaxPayload       int           // the largest POST body in bytes, 100KB by default
	MaxSessions      int           // how many sessions can be open at once, 10,000 by default
	MaxBuffered      int           // how many packets can wait for a session's next poll before it is closed, 1,000 by default
	MaxBufferedBytes int           // how many bytes of them can wait before it is closed, 2MB by default
	MaxPacketRate    float64       // how many packets a second a client may send, averaged, 20 by default; it may burst 5 seconds' worth
	// CheckOrigin decides which browser origins may open a websocket; nil allows only the server's own origin, and clients that send none
	CheckOrigin func(*http.Request) bool
	// ClientKey is who a handshake is from, such as its IP. A client may have at most MaxUnauthorized sessions open that
	// are not yet authorized (see Session.Authorize), MaxSessions by default, so it cannot take every session.
	ClientKey       func(*http.Request) string
	MaxUnauthorized int
}

func (c Config) withDefaults() Config {
	if c.PingInterval <= 0 {
		c.PingInterval = 25 * time.Second
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = 20 * time.Second
	}
	if c.MaxPayload <= 0 {
		c.MaxPayload = 100_000
	}
	if c.MaxSessions <= 0 {
		c.MaxSessions = 10_000
	}
	if c.MaxUnauthorized <= 0 {
		c.MaxUnauthorized = c.MaxSessions
	}
	if c.MaxBuffered <= 0 {
		c.MaxBuffered = 1_000
	}
	if c.MaxBufferedBytes <= 0 {
		c.MaxBufferedBytes = 2_000_000
	}
	if c.MaxPacketRate <= 0 {
		c.MaxPacketRate = 20
	}
	return c
}

// Server is an Engine.IO server over HTTP long-polling and websockets, mounted at a path like /socket.io/.
// A client can connect over either, and upgrade a polling session to a websocket.
type Server struct {
	config       Config
	onConnection func(*Session)

	mu       sync.Mutex
	sessions map[string]*Session
}

// NewServer calls onConnection with each new session; anything it sends goes after the open packet
func NewServer(config Config, onConnection func(*Session)) *Server {
	return &Server{config: config.withDefaults(), onConnection: onConnection, sessions: map[string]*Session{}}
}

// openPayload is the open packet's data in the handshake
type openPayload struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int64    `json:"pingInterval"`
	PingTimeout  int64    `json:"pingTimeout"`
	MaxPayload   int      `json:"maxPayload,omitempty"`
}

// Engine.IO error codes, as the JavaScript server sends them
const (
	errTransportUnknown   = 0
	errSessionUnknown     = 1
	errBadHandshakeMethod = 2
	errBadRequest         = 3
	errUnsupportedVersion = 5
)

var errorMessages = map[int]string{
	errTransportUnknown:   "Transport unknown",
	errSessionUnknown:     "Session ID unknown",
	errBadHandshakeMethod: "Bad handshake method",
	errBadRequest:         "Bad request",
	errUnsupportedVersion: "Unsupported protocol version",
}

func writeError(w http.ResponseWriter, status int, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"code": code, "message": errorMessages[code]})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	version, err := strconv.Atoi(query.Get("EIO"))
	if err != nil || (version != V3 && version != V4) {
		writeError(w, http.StatusBadRequest, errUnsupportedVersion)
		return
	}
	transport := query.Get("transport")
	if transport != "polling" && transport != "websocket" {
		writeError(w, http.StatusBadRequest, errTransportUnknown)
		return
	}
	sid := query.Get("sid")
	if transport == "websocket" {
		s.serveWebSocket(w, r, version, sid)
		return
	}
	if sid == "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusBadRequest, errBadHandshakeMethod)
			return
		}
		s.handshake(w, r, version)
		return
	}

	s.mu.Lock()
	session := s.sessions[sid]
	s.mu.Unlock()
	if session == nil || session.version != version {
		writeError(w, http.StatusBadRequest, errSessionUnknown)
		return
	}
	switch r.Method {
	case http.MethodGet:
		session.poll(w, r)
	case http.MethodPost:
		session.receive(w, r)
	default:
		writeError(w, http.StatusBadRequest, errBadRequest)
	}
}

// open starts a session and makes its open packet; a polling session can upgrade to a websocket
func (s *Server) open(r *http.Request, version int, upgrades []string) (*Session, Packet, bool) {
	client := ""
	if s.config.ClientKey != nil {
		client = s.config.ClientKey(r)
	}
	s.mu.Lock()
	unauthorized := 0
	for _, session := range s.sessions {
		if client != "" && session.client == client && !session.authorized.Load() {
			unauthorized++
		}
	}
	if len(s.sessions) >= s.config.MaxSessions || unauthorized >= s.config.MaxUnauthorized {
		s.mu.Unlock()
		return nil, Packet{}, false
	}
	session := newSession(s, r, version)
	session.client = client
	s.sessions[session.id] = session
	s.mu.Unlock()

	open := openPayload{
		SID:          session.id,
		Upgrades:     upgrades,
		PingInterval: s.config.PingInterval.Milliseconds(),
		PingTimeout:  s.config.PingTimeout.Milliseconds(),
	}
	if version == V4 {
		open.MaxPayload = s.config.MaxPayload
	}
	data, _ := json.Marshal(open)
	return session, Packet{Type: Open, Data: string(data)}, true
}

// started runs once a session's open packet is sent
func (s *Server) started(session *Session) {
	if s.onConnection != nil {
		s.onConnection(session)
	}
	go session.heartbeat()
}

func (s *Server) handshake(w http.ResponseWriter, r *http.Request, version int) {
	session, open, ok := s.open(r, version, []string{"websocket"})
	if !ok {
		http.Error(w, "Too many sessions", http.StatusTooManyRequests)
		return
	}
	writePayload(w, []Packet{open}, version)
	s.started(session)
}

// serveWebSocket opens a session over a websocket, or with a sid upgrades that polling session to one
func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request, version int, sid string) {
	var session *Session
	if sid != "" {
		s.mu.Lock()
		session = s.sessions[sid]
		s.mu.Unlock()
		if session == nil || session.version != version || session.webSocket() != nil {
			writeError(w, http.StatusBadRequest, errSessionUnknown)
			return
		}
	} else if s.Count() >= s.config.MaxSessions {
		http.Error(w, "Too many sessions", http.StatusTooManyRequests)
		return
	}

	ws, err := UpgradeWebSocket(w, r, s.config.MaxPayload, s.config.CheckOrigin)
	if err != nil {
		// The upgrader has already answered the request
		return
	}
	if session != nil {
		if !session.probe(ws) {
			ws.Close()
			return
		}
		session.runWebSocket(ws)
		return
	}

	session, open, ok := s.open(r, version, []string{})
	if !ok || ws.WritePacket(open) != nil {
		ws.Close()
		if session != nil {
			session.Close("transport error")
		}
		return
	}
	s.started(session)
	session.runWebSocket(ws)
}

// Close closes every session
func (s *Server) Close() {
	s.mu.Lock()
	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	s.mu.Unlock()
	for _, session := range sessions {
		session.Close("server shutting down")
	}
}

// Count is how many sessions are open
func (s *Server) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

func (s *Server) remove(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func writePayload(w http.ResponseWriter, packets []Packet, version int) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	io.WriteString(w, EncodePayload(packets, version))
}

var ErrClosed = errors.New("engineio: session closed")

// Session is one client's connection
type Session struct {
	id      string
	version int
	server  *Server
	request *http.Request
	client  string // the Config's ClientKey of the handshake
	// authorized is set once the session is authorized, so it no longer counts against its client
	authorized atomic.Bool

	limiter  *rate.Limiter // paces the client's packets
	handling sync.Mutex    // handles one packet at a time, even from a POST in flight as the session upgrades

	mu        sync.Mutex
	ws        *WebSocket // set once the session is on a websocket
	queue     []Packet
	queued    int // bytes in queue
	polling   bool
	probing   bool // the client is upgrading, so a poll is ended at once, as it waits for its poll to end to upgrade
	posting   bool
	closed    bool
	onMessage func(data string)
	onClose   func(reason string)
	notify    chan struct{} // wakes the waiting poll, or the websocket writer
	upgrade   chan struct{} // closed once the session is on a websocket
	alive     chan struct{} // a pong (v4) or ping (v3) came
	done      chan struct{}
}

func newSession(server *Server, r *http.Request, version int) *Session {
	id := make([]byte, 15)
	rand.Read(id)
	return &Session{
		id:      base64.RawURLEncoding.EncodeToString(id),
		version: version,
		server:  server,
		request: r.Clone(context.WithoutCancel(r.Context())),
		limiter: rate.NewLimiter(rate.Limit(server.config.MaxPacketRate), max(1, int(5*server.config.MaxPacketRate))),
		notify:  make(chan struct{}, 1),
		upgrade: make(chan struct{}),
		alive:   make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
}

func (s *Session) ID() string { return s.id }

// Version is the Engine.IO protocol version the client speaks
func (s *Session) Version() int { return s.version }

// Request is the handshake request, for its headers, query, remote address and context values (without its cancel)
func (s *Session) Request() *http.Request { return s.request }

// Authorize marks the session as given access, so it no longer counts against its client's Config.MaxUnauthorized
func (s *Session) Authorize() { s.authorized.Store(true) }

// OnMessage sets the handler for the client's messages; it runs in order, one at a time, so must not block
func (s *Session) OnMessage(handler func(data string)) {
	s.mu.Lock()
	s.onMessage = handler
	s.mu.Unlock()
}

// OnClose sets the handler called once when the session closes
func (s *Session) OnClose(handler func(reason string)) {
	s.mu.Lock()
	s.onClose = handler
	s.mu.Unlock()
}

// Done is closed when the session closes
func (s *Session) Done() <-chan struct{} { return s.done }

// Send queues a message for the client; a client too slow to collect them, in packets or bytes, is closed
func (s *Session) Send(data string) error {
	return s.enqueue(Packet{Type: Message, Data: data})
}

func (s *Session) enqueue(p Packet) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}
	if len(s.queue) >= s.server.config.MaxBuffered {
		s.mu.Unlock()
		s.Close("too many packets buffered")
		return ErrClosed
	}
	if s.queued += len(p.Data) + len(p.Binary); s.queued > s.server.config.MaxBufferedBytes {
		s.mu.Unlock()
		s.Close("too many bytes buffered")
		return ErrClosed
	}
	s.queue = append(s.queue, p)
	s.mu.Unlock()
	signal(s.notify)
	return nil
}

// Close ends the session; a waiting poll gets a close packet
func (s *Session) Close(reason string) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	onClose, ws := s.onClose, s.ws
	s.mu.Unlock()

	s.server.remove(s.id)
	close(s.done)
	if ws != nil {
		ws.Close()
	}
	if onClose != nil {
		onClose(reason)
	}
}

func signal(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default:
	}
}

// heartbeat closes the session when the client stops answering: in v4 the server pings,
// in v3 the client must ping within pingInterval + pingTimeout
func (s *Session) heartbeat() {
	interval, timeout := s.server.config.PingInterval, s.server.config.PingTimeout
	for {
		wait := interval + timeout
		if s.version == V4 {
			if !sleep(interval, s.done) {
				return
			}
			if s.enqueue(Packet{Type: Ping}) != nil {
				return
			}
			wait = timeout
		}
		timer := time.NewTimer(wait)
		select {
		case <-s.alive:
			timer.Stop()
		case <-timer.C:
			s.Close("ping timeout")
			return
		case <-s.done:
			timer.Stop()
			return
		}
	}
}

// sleep waits for d, or is false if done closes first
func sleep(d time.Duration, done <-chan struct{}) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-done:
		return false
	}
}

// poll answers a GET with the queued packets, waiting up to pingInterval for some
func (s *Session) poll(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.ws != nil {
		s.mu.Unlock()
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	if s.polling {
		s.mu.Unlock()
		s.Close("concurrent poll")
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	s.polling = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.polling = false
		s.mu.Unlock()
	}()

	// The wait can outlast the HTTP server's write timeout
	http.NewResponseController(w).SetWriteDeadline(time.Now().Add(s.server.config.PingInterval + writeTimeout))
	timer := time.NewTimer(s.server.config.PingInterval)
	defer timer.Stop()
	for {
		s.mu.Lock()
		upgraded := s.ws != nil
		var packets []Packet
		if !upgraded {
			packets, s.queue, s.queued = s.queue, nil, 0
		}
		closed, probing := s.closed, s.probing
		s.mu.Unlock()
		switch {
		case upgraded:
			// The upgrade ends the waiting poll; the websocket sends from now on, so pass on any wake this poll took
			signal(s.notify)
			writePayload(w, []Packet{{Type: Noop}}, s.version)
			return
		case len(packets) > 0:
			writePayload(w, packets, s.version)
			return
		case closed:
			writePayload(w, []Packet{{Type: Close}}, s.version)
			return
		case probing:
			writePayload(w, []Packet{{Type: Noop}}, s.version)
			return
		}
		select {
		case <-s.notify:
		case <-s.upgrade:
		case <-s.done:
		case <-timer.C:
			writePayload(w, []Packet{{Type: Noop}}, s.version)
			return
		case <-r.Context().Done():
			return
		}
	}
}

// receive handles a POST of the client's packets
func (s *Session) receive(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	// Once upgraded the client sends over the websocket, as the JavaScript server refuses a transport mismatch
	if s.ws != nil {
		s.mu.Unlock()
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	if s.posting {
		s.mu.Unlock()
		s.Close("concurrent post")
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	s.posting = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.posting = false
		s.mu.Unlock()
	}()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, int64(s.server.config.MaxPayload)))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.Close("payload too large")
			http.Error(w, "Payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	packets, err := DecodePayload(string(body), s.version)
	if err != nil {
		s.Close("bad payload")
		writeError(w, http.StatusBadRequest, errBadRequest)
		return
	}
	for _, p := range packets {
		s.handle(p)
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	io.WriteString(w, "ok")
}

// handle acts on one of the client's packets; a client sending them faster than MaxPacketRate is closed
func (s *Session) handle(p Packet) {
	s.handling.Lock()
	defer s.handling.Unlock()
	if !s.limiter.Allow() {
		s.Close("too many packets")
		return
	}
	switch p.Type {
	case Message:
		s.mu.Lock()
		onMessage := s.onMessage
		s.mu.Unlock()
		// Binary messages are dropped: nothing above this layer uses them
		if onMessage != nil && p.Binary == nil {
			onMessage(p.Data)
		}
	case Ping:
		if s.version == V3 {
			signal(s.alive)
			s.enqueue(Packet{Type: Pong, Data: p.Data})
		}
	case Pong:
		if s.version == V4 {
			signal(s.alive)
		}
	case Close:
		s.Close("client close")
	}
}

func (s *Session) webSocket() *WebSocket {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ws
}

// probe is the start of an upgrade: the client pings "probe" over the websocket, the server answers,
// then the client sends upgrade. It is false if the client sends anything else, or takes longer than pingTimeout.
func (s *Session) probe(ws *WebSocket) bool {
	ws.conn.SetReadDeadline(time.Now().Add(s.server.config.PingTimeout))
	defer ws.conn.SetReadDeadline(time.Time{})
	p, err := ws.ReadPacket()
	if err != nil || p.Type != Ping || p.Data != "probe" || ws.WritePacket(Packet{Type: Pong, Data: "probe"}) != nil {
		return false
	}
	// The client pauses polling, waiting for its poll to end, before it sends upgrade, so the waiting poll is ended
	// with a noop, as the JavaScript server does
	s.setProbing(true)
	defer s.setProbing(false)
	p, err = ws.ReadPacket()
	return err == nil && p.Type == Upgrade
}

func (s *Session) setProbing(probing bool) {
	s.mu.Lock()
	s.probing = probing
	s.mu.Unlock()
	if probing {
		signal(s.notify)
	}
}

// runWebSocket moves the session onto the websocket: a waiting poll is ended with a noop, queued packets are sent
// over the websocket from now on, and the client's packets are read from it until it closes
func (s *Session) runWebSocket(ws *WebSocket) {
	s.mu.Lock()
	if s.closed || s.ws != nil {
		s.mu.Unlock()
		ws.Close()
		return
	}
	s.ws = ws
	close(s.upgrade)
	s.mu.Unlock()

	go s.writeWebSocket(ws)
	for {
		p, err := ws.ReadPacket()
		if err != nil {
			s.Close("transport close")
			return
		}
		s.handle(p)
	}
}

// writeWebSocket sends queued packets as they come, until the session closes
func (s *Session) writeWebSocket(ws *WebSocket) {
	for {
		s.mu.Lock()
		packets := s.queue
		s.queue, s.queued = nil, 0
		s.mu.Unlock()
		for _, p := range packets {
			if ws.WritePacket(p) != nil {
				s.Close("transport error")
				return
			}
		}
		select {
		case <-s.notify:
		case <-s.done:
			return
		}
	}
}
