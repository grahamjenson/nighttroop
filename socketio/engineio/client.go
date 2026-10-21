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

// Client is an Engine.IO client over HTTP long-polling
type Client struct {
	url          string
	version      int
	http         *http.Client
	onMessage    func(data string)
	pingInterval time.Duration
	pingTimeout  time.Duration

	post  sync.Mutex // one POST at a time keeps packets in order
	pong  chan struct{}
	done  chan struct{}
	once  sync.Once
	mu    sync.Mutex
	cause error
}

// Dial makes the handshake with the server at rawURL (e.g. http://host/socket.io/) and starts polling.
// onMessage gets the server's messages in order, one at a time, on the polling goroutine.
func Dial(ctx context.Context, rawURL string, version int, httpClient *http.Client, onMessage func(data string)) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{version: version, http: httpClient, onMessage: onMessage, pong: make(chan struct{}, 1), done: make(chan struct{})}
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
	if len(packets) == 0 || packets[0].Type != Open {
		return nil, fmt.Errorf("%w: handshake did not open", ErrBadPacket)
	}
	var open openPayload
	if err := json.Unmarshal([]byte(packets[0].Data), &open); err != nil || open.SID == "" {
		return nil, fmt.Errorf("%w: bad open packet", ErrBadPacket)
	}
	query.Set("sid", open.SID)
	u.RawQuery = query.Encode()
	c.url = u.String()
	c.pingInterval = time.Duration(open.PingInterval) * time.Millisecond
	c.pingTimeout = time.Duration(open.PingTimeout) * time.Millisecond

	// A v3 server can send messages with its open packet
	c.handle(packets[1:])
	go c.pollLoop()
	if version == V3 {
		go c.pingLoop()
	}
	return c, nil
}

// Done is closed when the client closes; Err then says why
func (c *Client) Done() <-chan struct{} { return c.done }

func (c *Client) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cause
}

// Send posts a message to the server
func (c *Client) Send(data string) error {
	return c.send(Packet{Type: Message, Data: data})
}

// Close tells the server the client is going, then stops polling
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

func (c *Client) pollLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c.done
		cancel()
	}()
	for {
		packets, err := c.get(ctx, c.url)
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
		c.mu.Unlock()
		close(c.done)
	})
}
