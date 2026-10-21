package socketio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"nighttroop.t1d.nz/socketio/engineio"
)

// Client is a Socket.IO client over long-polling or a websocket, used to test the server
type Client struct {
	engine *engineio.Client

	mu        sync.Mutex
	version   int
	handlers  map[string]map[string]func(args []json.RawMessage)
	connected map[string]bool
	waiting   map[string]chan error
	acks      map[int64]chan []json.RawMessage
	nextAck   int64
}

// Dial opens a connection to the server at rawURL (e.g. http://host/socket.io/); call Connect before emitting.
// version is the Engine.IO version: engineio.V4 for Socket.IO v5, engineio.V3 for Socket.IO v2 clients.
func Dial(ctx context.Context, rawURL string, version int, httpClient *http.Client) (*Client, error) {
	c := NewClient(version)
	if err := c.Open(ctx, rawURL, httpClient); err != nil {
		return nil, err
	}
	return c, nil
}

// Open is Dial for a client from NewClient, so On can be set before anything arrives:
// a Socket.IO v2 server connects the default namespace, and can emit, before Dial returns
func (c *Client) Open(ctx context.Context, rawURL string, httpClient *http.Client) error {
	engine, err := engineio.Dial(ctx, rawURL, c.version, httpClient, c.onMessage)
	if err != nil {
		return err
	}
	c.engine = engine
	return nil
}

// DialOverWebSocket is Dial over a websocket from the start
func DialOverWebSocket(ctx context.Context, rawURL string, version int) (*Client, error) {
	c := NewClient(version)
	engine, err := engineio.DialOverWebSocket(ctx, rawURL, version, c.onMessage)
	if err != nil {
		return nil, err
	}
	c.engine = engine
	return c, nil
}

// NewClient is a client that is not yet connected; set its handlers with On, then Open it
func NewClient(version int) *Client {
	return &Client{
		version:   version,
		handlers:  map[string]map[string]func([]json.RawMessage){},
		connected: map[string]bool{},
		waiting:   map[string]chan error{},
		acks:      map[int64]chan []json.RawMessage{},
	}
}

// Upgrade moves a polling client onto a websocket, as browsers do after connecting
func (c *Client) Upgrade(ctx context.Context) error { return c.engine.Upgrade(ctx) }

// Connect joins a namespace, sending auth (if any) with v5, and waits for the server's answer.
// A Socket.IO v2 server connects the default namespace itself, so Connect only waits for it.
func (c *Client) Connect(ctx context.Context, namespace string, auth any) error {
	c.mu.Lock()
	if c.connected[namespace] {
		c.mu.Unlock()
		return nil
	}
	wait := make(chan error, 1)
	c.waiting[namespace] = wait
	c.mu.Unlock()

	if c.version != engineio.V3 || namespace != DefaultNamespace {
		p := Packet{Type: Connect, Namespace: namespace}
		if auth != nil && c.version == engineio.V4 {
			data, err := json.Marshal(auth)
			if err != nil {
				return err
			}
			p.Data = data
		}
		if err := c.engine.Send(p.Encode()); err != nil {
			return err
		}
	}
	select {
	case err := <-wait:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-c.engine.Done():
		return engineio.ErrClosed
	}
}

// On sets the handler for an event from the server in a namespace
func (c *Client) On(namespace, event string, handler func(args []json.RawMessage)) {
	c.mu.Lock()
	if c.handlers[namespace] == nil {
		c.handlers[namespace] = map[string]func([]json.RawMessage){}
	}
	c.handlers[namespace][event] = handler
	c.mu.Unlock()
}

// Emit sends an event to the server
func (c *Client) Emit(namespace, event string, args ...any) error {
	data, err := EventData(event, args...)
	if err != nil {
		return err
	}
	return c.engine.Send(Packet{Type: Event, Namespace: namespace, Data: data}.Encode())
}

// EmitWithAck sends an event and waits for the server's acknowledgement
func (c *Client) EmitWithAck(ctx context.Context, namespace, event string, args ...any) ([]json.RawMessage, error) {
	data, err := EventData(event, args...)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	id := c.nextAck
	c.nextAck++
	reply := make(chan []json.RawMessage, 1)
	c.acks[id] = reply
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.acks, id)
		c.mu.Unlock()
	}()

	if err := c.engine.Send(Packet{Type: Event, Namespace: namespace, ID: &id, Data: data}.Encode()); err != nil {
		return nil, err
	}
	select {
	case args := <-reply:
		return args, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.engine.Done():
		return nil, engineio.ErrClosed
	}
}

// Connected is whether the client is in a namespace
func (c *Client) Connected(namespace string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected[namespace]
}

// Disconnect leaves a namespace
func (c *Client) Disconnect(namespace string) error {
	c.mu.Lock()
	delete(c.connected, namespace)
	c.mu.Unlock()
	return c.engine.Send(Packet{Type: Disconnect, Namespace: namespace}.Encode())
}

// Close ends the connection
func (c *Client) Close() error { return c.engine.Close() }

// Done is closed when the connection closes
func (c *Client) Done() <-chan struct{} { return c.engine.Done() }

func (c *Client) onMessage(data string) {
	p, err := DecodePacket(data)
	if err != nil {
		return
	}
	switch p.Type {
	case Connect:
		c.mu.Lock()
		c.connected[p.Namespace] = true
		c.mu.Unlock()
		c.answer(p.Namespace, nil)
	case ConnectError:
		c.answer(p.Namespace, connectErrorMessage(p.Data))
	case Disconnect:
		c.mu.Lock()
		delete(c.connected, p.Namespace)
		c.mu.Unlock()
	case Event:
		event, args, err := ParseEvent(p.Data)
		if err != nil {
			return
		}
		c.mu.Lock()
		handler := c.handlers[p.Namespace][event]
		c.mu.Unlock()
		if handler != nil {
			handler(args)
		}
	case Ack:
		var args []json.RawMessage
		if p.ID == nil || json.Unmarshal(p.Data, &args) != nil {
			return
		}
		c.mu.Lock()
		reply := c.acks[*p.ID]
		c.mu.Unlock()
		if reply != nil {
			reply <- args
		}
	}
}

func (c *Client) answer(namespace string, err error) {
	c.mu.Lock()
	wait := c.waiting[namespace]
	delete(c.waiting, namespace)
	c.mu.Unlock()
	if wait != nil {
		wait <- err
	}
}

// connectErrorMessage reads v5's {"message": ...} or v2's bare string
func connectErrorMessage(data json.RawMessage) error {
	var v5 struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &v5) == nil && v5.Message != "" {
		return errors.New(v5.Message)
	}
	var v2 string
	if json.Unmarshal(data, &v2) == nil && v2 != "" {
		return errors.New(v2)
	}
	return errors.New("connect refused")
}
