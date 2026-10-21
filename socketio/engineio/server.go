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
	"time"
)

// Config is how a Server runs its sessions; zero values take the defaults
type Config struct {
	PingInterval time.Duration // how often the server (v4) or client (v3) pings, 25s by default
	PingTimeout  time.Duration // how long to wait for the answer, 20s by default
	MaxPayload   int           // the largest POST body in bytes, 100KB by default
	MaxSessions  int           // how many sessions can be open at once, 10,000 by default
	MaxBuffered  int           // how many packets can wait for a session's next poll before it is closed, 1,000 by default
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
	if c.MaxBuffered <= 0 {
		c.MaxBuffered = 1_000
	}
	return c
}

// Server is an Engine.IO server over HTTP long-polling, mounted at a path like /socket.io/
type Server struct {
	config       Config
	onConnection func(*Session)

	mu       sync.Mutex
	sessions map[string]*Session
}

// NewServer calls onConnection with each new session, before its handshake is answered
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
	if query.Get("transport") != "polling" {
		writeError(w, http.StatusBadRequest, errTransportUnknown)
		return
	}
	sid := query.Get("sid")
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

func (s *Server) handshake(w http.ResponseWriter, r *http.Request, version int) {
	s.mu.Lock()
	if len(s.sessions) >= s.config.MaxSessions {
		s.mu.Unlock()
		http.Error(w, "Too many sessions", http.StatusTooManyRequests)
		return
	}
	session := newSession(s, r, version)
	s.sessions[session.id] = session
	s.mu.Unlock()

	open := openPayload{
		SID:          session.id,
		Upgrades:     []string{},
		PingInterval: s.config.PingInterval.Milliseconds(),
		PingTimeout:  s.config.PingTimeout.Milliseconds(),
	}
	if version == V4 {
		open.MaxPayload = s.config.MaxPayload
	}
	data, _ := json.Marshal(open)
	if s.onConnection != nil {
		s.onConnection(session)
	}
	go session.heartbeat()
	writePayload(w, []Packet{{Type: Open, Data: string(data)}}, version)
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

	mu        sync.Mutex
	queue     []Packet
	polling   bool
	posting   bool
	closed    bool
	onMessage func(data string)
	onClose   func(reason string)
	notify    chan struct{} // wakes the waiting poll
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
		request: r.Clone(context.Background()),
		notify:  make(chan struct{}, 1),
		alive:   make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
}

func (s *Session) ID() string { return s.id }

// Version is the Engine.IO protocol version the client speaks
func (s *Session) Version() int { return s.version }

// Request is the handshake request, for its headers, query and remote address
func (s *Session) Request() *http.Request { return s.request }

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

// Send queues a message for the client's next poll; a client too slow to collect them is closed
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
	onClose := s.onClose
	s.mu.Unlock()

	s.server.remove(s.id)
	close(s.done)
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

	timer := time.NewTimer(s.server.config.PingInterval)
	defer timer.Stop()
	for {
		s.mu.Lock()
		packets := s.queue
		s.queue = nil
		closed := s.closed
		s.mu.Unlock()
		switch {
		case len(packets) > 0:
			writePayload(w, packets, s.version)
			return
		case closed:
			writePayload(w, []Packet{{Type: Close}}, s.version)
			return
		}
		select {
		case <-s.notify:
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

func (s *Session) handle(p Packet) {
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
