package engineio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client is an Engine.IO client over HTTP long-polling, which can upgrade to a websocket, or over a websocket from the start
type Client struct {
	baseURL      string
	sid          string
	url          string
	version      int
	http         *http.Client
	onMessage    func(data string)
	pingInterval time.Duration
	pingTimeout  time.Duration

	post     sync.Mutex // one POST at a time keeps packets in order
	pong     chan struct{}
	done     chan struct{}
	pollDone chan struct{} // closed when polling stops
	once     sync.Once
	mu       sync.Mutex
	ws       *WebSocket // set once the client is on a websocket
	cause    error
}

func newClient(rawURL string, version int, httpClient *http.Client, onMessage func(data string)) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: rawURL, version: version, http: httpClient, onMessage: onMessage, pong: make(chan struct{}, 1), done: make(chan struct{}), pollDone: make(chan struct{})}
}

// opened reads the session from the server's open packet
func (c *Client) opened(p Packet) error {
	var open openPayload
	if p.Type != Open || json.Unmarshal([]byte(p.Data), &open) != nil || open.SID == "" {
		return fmt.Errorf("%w: bad open packet", ErrBadPacket)
	}
	c.sid = open.SID
	c.pingInterval = time.Duration(open.PingInterval) * time.Millisecond
	c.pingTimeout = time.Duration(open.PingTimeout) * time.Millisecond
	return nil
}

// Dial makes the handshake with the server at rawURL (e.g. http://host/socket.io/) and starts polling.
// onMessage gets the server's messages in order, one at a time, on the polling goroutine.
func Dial(ctx context.Context, rawURL string, version int, httpClient *http.Client, onMessage func(data string)) (*Client, error) {
	c := newClient(rawURL, version, httpClient, onMessage)
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("EIO", strconv.Itoa(version))
	query.Set("transport", "polling")
	u.RawQuery = query.Encode()

	packets, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}
	if len(packets) == 0 {
		return nil, fmt.Errorf("%w: handshake did not open", ErrBadPacket)
	}
	if err := c.opened(packets[0]); err != nil {
		return nil, err
	}
	query.Set("sid", c.sid)
	u.RawQuery = query.Encode()
	c.url = u.String()

	// A v3 server can send messages with its open packet
	c.handle(packets[1:])
	go c.pollLoop()
	if version == V3 {
		go c.pingLoop()
	}
	return c, nil
}

// DialOverWebSocket opens a session over a websocket, without polling first.
// onMessage gets the server's messages in order, one at a time, on the reading goroutine.
func DialOverWebSocket(ctx context.Context, rawURL string, version int, onMessage func(data string)) (*Client, error) {
	c := newClient(rawURL, version, nil, onMessage)
	close(c.pollDone)
	ws, err := DialWebSocket(ctx, rawURL, version, "")
	if err != nil {
		return nil, err
	}
	p, err := readPacket(ctx, ws)
	if err == nil {
		err = c.opened(p)
	}
	if err != nil {
		ws.Close()
		return nil, err
	}
	c.ws = ws
	go c.readLoop(ws)
	if version == V3 {
		go c.pingLoop()
	}
	return c, nil
}

// Upgrade moves a polling client onto a websocket: it probes the websocket, then sends upgrade,
// and messages arrive over it once the waiting poll ends
func (c *Client) Upgrade(ctx context.Context) error {
	ws, err := DialWebSocket(ctx, c.baseURL, c.version, c.sid)
	if err != nil {
		return err
	}
	if err := ws.WritePacket(Packet{Type: Ping, Data: "probe"}); err != nil {
		ws.Close()
		return err
	}
	p, err := readPacket(ctx, ws)
	if err != nil || p.Type != Pong || p.Data != "probe" {
		ws.Close()
		return fmt.Errorf("%w: bad probe answer", ErrBadPacket)
	}

	// Holding post waits out a POST in flight, so packets stay in order
	c.post.Lock()
	c.mu.Lock()
	if c.cause != nil || c.ws != nil {
		c.mu.Unlock()
		c.post.Unlock()
		ws.Close()
		return ErrClosed
	}
	c.ws = ws
	c.mu.Unlock()
	err = ws.WritePacket(Packet{Type: Upgrade})
	c.post.Unlock()
	if err != nil {
		c.stop(err)
		return err
	}
	// Reading starts once polling stops, so messages are handled in order
	<-c.pollDone
	go c.readLoop(ws)
	return nil
}

// readPacket reads one packet, giving up at ctx's deadline
func readPacket(ctx context.Context, ws *WebSocket) (Packet, error) {
	if deadline, ok := ctx.Deadline(); ok {
		ws.conn.SetReadDeadline(deadline)
		defer ws.conn.SetReadDeadline(time.Time{})
	}
	return ws.ReadPacket()
}

func (c *Client) readLoop(ws *WebSocket) {
	for {
		p, err := ws.ReadPacket()
		if err != nil {
			c.stop(err)
			return
		}
		c.handle([]Packet{p})
	}
}

func (c *Client) webSocket() *WebSocket {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws
}

// Done is closed when the client closes; Err then says why
func (c *Client) Done() <-chan struct{} { return c.done }

func (c *Client) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cause
}

// Send sends a message to the server
func (c *Client) Send(data string) error {
	return c.send(Packet{Type: Message, Data: data})
}

// Close tells the server the client is going, then stops
func (c *Client) Close() error {
	err := c.send(Packet{Type: Close})
	c.stop(ErrClosed)
	return err
}

func (c *Client) send(packets ...Packet) error {
	select {
	case <-c.done:
		return ErrClosed
	default:
	}
	c.post.Lock()
	defer c.post.Unlock()
	if ws := c.webSocket(); ws != nil {
		for _, p := range packets {
			if err := ws.WritePacket(p); err != nil {
				c.stop(err)
				return err
			}
		}
		return nil
	}
	resp, err := c.http.Post(c.url, "text/plain;charset=UTF-8", strings.NewReader(EncodePayload(packets, c.version)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("engineio: post got %d: %s", resp.StatusCode, body)
		c.stop(err)
		return err
	}
	return nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]Packet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("engineio: poll got %d: %s", resp.StatusCode, body)
	}
	return DecodePayload(string(body), c.version)
}

// pollLoop polls until the client closes, or upgrades
func (c *Client) pollLoop() {
	defer close(c.pollDone)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c.done
		cancel()
	}()
	for {
		packets, err := c.get(ctx, c.url)
		if c.webSocket() != nil {
			// The upgrade ended this poll, or refused it if it came after
			if err == nil {
				c.handle(packets)
			}
			return
		}
		if err != nil {
			c.stop(err)
			return
		}
		c.handle(packets)
		select {
		case <-c.done:
			return
		default:
		}
	}
}

func (c *Client) handle(packets []Packet) {
	for _, p := range packets {
		switch p.Type {
		case Message:
			if c.onMessage != nil && p.Binary == nil {
				c.onMessage(p.Data)
			}
		case Ping:
			// v4: the server pings, the client answers; posting from a goroutine keeps polling
			go c.send(Packet{Type: Pong, Data: p.Data})
		case Pong:
			signal(c.pong)
		case Close:
			c.stop(ErrClosed)
		}
	}
}

// pingLoop is v3's heartbeat: the client pings every pingInterval and needs a pong within pingTimeout
func (c *Client) pingLoop() {
	for sleep(c.pingInterval, c.done) {
		if err := c.send(Packet{Type: Ping}); err != nil {
			c.stop(err)
			return
		}
		timer := time.NewTimer(c.pingTimeout)
		select {
		case <-c.pong:
			timer.Stop()
		case <-timer.C:
			c.stop(errors.New("engineio: ping timeout"))
			return
		case <-c.done:
			timer.Stop()
			return
		}
	}
}

func (c *Client) stop(cause error) {
	c.once.Do(func() {
		c.mu.Lock()
		c.cause = cause
		ws := c.ws
		c.mu.Unlock()
		close(c.done)
		if ws != nil {
			ws.Close()
		}
	})
}
