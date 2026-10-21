package engineio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fast heartbeats so tests see timeouts quickly
var testConfig = Config{PingInterval: 50 * time.Millisecond, PingTimeout: 50 * time.Millisecond}

// echoServer sends every message back, and records why each session closed
func echoServer(t *testing.T, config Config) (*Server, *httptest.Server, chan string) {
	closed := make(chan string, 10)
	server := NewServer(config, func(s *Session) {
		s.OnMessage(func(data string) { s.Send(data) })
		s.OnClose(func(reason string) { closed <- reason })
	})
	ts := httptest.NewServer(server)
	t.Cleanup(func() {
		server.Close()
		ts.Close()
	})
	return server, ts, closed
}

func get(t *testing.T, url string) (int, string) {
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func post(t *testing.T, url, body string) (int, string) {
	resp, err := http.Post(url, "text/plain", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

// handshake opens a session by hand and returns its sid
func handshake(t *testing.T, base string, version string) string {
	status, body := get(t, base+"/?EIO="+version+"&transport=polling")
	require.Equal(t, http.StatusOK, status)
	if version == "3" {
		body = body[strings.Index(body, ":")+1:]
	}
	require.Equal(t, byte(Open), body[0])
	var open openPayload
	require.NoError(t, json.Unmarshal([]byte(body[1:]), &open))
	return open.SID
}

func TestHandshake(t *testing.T) {
	_, ts, _ := echoServer(t, Config{})

	status, body := get(t, ts.URL+"/?EIO=4&transport=polling")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, byte(Open), body[0])
	var open map[string]any
	require.NoError(t, json.Unmarshal([]byte(body[1:]), &open))
	require.NotEmpty(t, open["sid"])
	// A polling session can upgrade to a websocket
	require.Equal(t, []any{"websocket"}, open["upgrades"])
	require.EqualValues(t, 25000, open["pingInterval"])
	require.EqualValues(t, 20000, open["pingTimeout"])
	require.EqualValues(t, 100000, open["maxPayload"])

	// v3 puts the packet's length first, and has no maxPayload
	status, body = get(t, ts.URL+"/?EIO=3&transport=polling")
	require.Equal(t, http.StatusOK, status)
	length, packet, found := strings.Cut(body, ":")
	require.True(t, found)
	require.Equal(t, length, strings.TrimSpace(length))
	require.NotContains(t, packet, "maxPayload")
}

var transports = []string{"polling", "websocket", "upgrade"}

// dial connects a client over polling, over a websocket, or over polling then upgraded to a websocket
func dial(t *testing.T, url string, version int, transport string, onMessage func(string)) *Client {
	if transport == "websocket" {
		client, err := DialOverWebSocket(context.Background(), url, version, onMessage)
		require.NoError(t, err)
		return client
	}
	client, err := Dial(context.Background(), url, version, nil, onMessage)
	require.NoError(t, err)
	if transport == "upgrade" {
		require.NoError(t, client.Upgrade(context.Background()))
	}
	return client
}

func TestClientRoundTrip(t *testing.T) {
	for _, version := range []int{V3, V4} {
		for _, transport := range transports {
			t.Run("v"+string(rune('0'+version))+" "+transport, func(t *testing.T) {
				testRoundTrip(t, version, transport)
			})
		}
	}
}

func testRoundTrip(t *testing.T, version int, transport string) {
	_, ts, closed := echoServer(t, testConfig)
	received := make(chan string, 10)
	client := dial(t, ts.URL+"/socket.io/", version, transport, func(data string) { received <- data })

	for _, msg := range []string{"hello", "€ and 😀", `{"json":true}`} {
		require.NoError(t, client.Send(msg))
		require.Equal(t, msg, <-received)
	}

	// The heartbeat keeps the session open well past several ping intervals
	time.Sleep(300 * time.Millisecond)
	require.NoError(t, client.Send("still here"))
	require.Equal(t, "still here", <-received)

	require.NoError(t, client.Close())
	require.Equal(t, "client close", <-closed)
	<-client.Done()
}

func TestSessionClosesWithoutHeartbeat(t *testing.T) {
	for _, version := range []string{"3", "4"} {
		t.Run("v"+version, func(t *testing.T) {
			server, ts, closed := echoServer(t, testConfig)
			// A client that never pings (v3) or answers pings (v4)
			sid := handshake(t, ts.URL, version)
			require.Equal(t, "ping timeout", <-closed)
			require.Equal(t, 0, server.Count())

			status, body := get(t, ts.URL+"/?EIO="+version+"&transport=polling&sid="+sid)
			require.Equal(t, http.StatusBadRequest, status)
			require.JSONEq(t, `{"code":1,"message":"Session ID unknown"}`, body)
		})
	}
}

func TestPollWaitsThenSendsNoop(t *testing.T) {
	_, ts, _ := echoServer(t, Config{PingInterval: 100 * time.Millisecond, PingTimeout: time.Second})
	sid := handshake(t, ts.URL, "3")
	start := time.Now()
	status, body := get(t, ts.URL+"/?EIO=3&transport=polling&sid="+sid)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "1:6", body)
	require.GreaterOrEqual(t, time.Since(start), 100*time.Millisecond)
}

func TestRequestErrors(t *testing.T) {
	_, ts, _ := echoServer(t, testConfig)
	sid := handshake(t, ts.URL, "4")
	tests := []struct {
		name   string
		method string
		query  string
		status int
		code   int
	}{
		{"no version", "GET", "transport=polling", 400, errUnsupportedVersion},
		{"unknown version", "GET", "EIO=5&transport=polling", 400, errUnsupportedVersion},
		{"unknown transport", "GET", "EIO=4&transport=flashsocket", 400, errTransportUnknown},
		{"upgrade with unknown sid", "GET", "EIO=4&transport=websocket&sid=nope", 400, errSessionUnknown},
		{"upgrade with wrong version for sid", "GET", "EIO=3&transport=websocket&sid=" + sid, 400, errSessionUnknown},
		{"post handshake", "POST", "EIO=4&transport=polling", 400, errBadHandshakeMethod},
		{"unknown sid", "GET", "EIO=4&transport=polling&sid=nope", 400, errSessionUnknown},
		{"wrong version for sid", "GET", "EIO=3&transport=polling&sid=" + sid, 400, errSessionUnknown},
		{"put", "PUT", "EIO=4&transport=polling&sid=" + sid, 400, errBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, ts.URL+"/?"+test.query, nil)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			require.Equal(t, test.status, resp.StatusCode)
			var got map[string]any
			require.NoError(t, json.Unmarshal(body, &got))
			require.EqualValues(t, test.code, got["code"])
		})
	}
}

func TestBadPostsCloseTheSession(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
		reason string
	}{
		{"bad payload", "9nonsense", http.StatusBadRequest, "bad payload"},
		{"too large", "4" + strings.Repeat("x", 200), http.StatusRequestEntityTooLarge, "payload too large"},
		{"close packet", "1", http.StatusOK, "client close"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, ts, closed := echoServer(t, Config{MaxPayload: 100})
			sid := handshake(t, ts.URL, "4")
			status, _ := post(t, ts.URL+"/?EIO=4&transport=polling&sid="+sid, test.body)
			require.Equal(t, test.status, status)
			require.Equal(t, test.reason, <-closed)
		})
	}
}

func TestConcurrentPollClosesTheSession(t *testing.T) {
	_, ts, closed := echoServer(t, Config{})
	sid := handshake(t, ts.URL, "4")
	url := ts.URL + "/?EIO=4&transport=polling&sid=" + sid

	var wg sync.WaitGroup
	wg.Add(1)
	var firstBody string
	go func() {
		defer wg.Done()
		_, firstBody = get(t, url)
	}()
	time.Sleep(50 * time.Millisecond)
	status, _ := get(t, url)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "concurrent poll", <-closed)
	wg.Wait()
	// The waiting poll is told the session closed
	require.Equal(t, "1", firstBody)
}

func TestMaxSessions(t *testing.T) {
	_, ts, _ := echoServer(t, Config{MaxSessions: 2})
	handshake(t, ts.URL, "4")
	handshake(t, ts.URL, "4")
	status, _ := get(t, ts.URL+"/?EIO=4&transport=polling")
	require.Equal(t, http.StatusTooManyRequests, status)
}

func TestSlowClientIsClosedWhenItsBufferFills(t *testing.T) {
	var session *Session
	server := NewServer(Config{MaxBuffered: 3}, func(s *Session) { session = s })
	ts := httptest.NewServer(server)
	defer ts.Close()
	handshake(t, ts.URL, "4")

	for range 3 {
		require.NoError(t, session.Send("x"))
	}
	require.ErrorIs(t, session.Send("x"), ErrClosed)
	<-session.Done()
	require.Equal(t, 0, server.Count())
}

func TestServerCloseEndsClients(t *testing.T) {
	for _, transport := range transports {
		t.Run(transport, func(t *testing.T) {
			server, ts, _ := echoServer(t, testConfig)
			client := dial(t, ts.URL, V4, transport, nil)
			server.Close()
			<-client.Done()
			require.ErrorIs(t, client.Send("x"), ErrClosed)
		})
	}
}

// upgrade probes and upgrades the session sid by hand
func upgrade(t *testing.T, base string, version int, sid string) *WebSocket {
	ws, err := DialWebSocket(context.Background(), base, version, sid)
	require.NoError(t, err)
	t.Cleanup(func() { ws.Close() })
	require.NoError(t, ws.WritePacket(Packet{Type: Ping, Data: "probe"}))
	p, err := ws.ReadPacket()
	require.NoError(t, err)
	require.Equal(t, Packet{Type: Pong, Data: "probe"}, p)
	require.NoError(t, ws.WritePacket(Packet{Type: Upgrade}))
	return ws
}

func TestUpgradeEndsTheWaitingPoll(t *testing.T) {
	_, ts, _ := echoServer(t, Config{})
	sid := handshake(t, ts.URL, "4")
	url := ts.URL + "/?EIO=4&transport=polling&sid=" + sid

	polled := make(chan string)
	go func() {
		_, body := get(t, url)
		polled <- body
	}()
	time.Sleep(50 * time.Millisecond)
	start := time.Now()
	ws := upgrade(t, ts.URL, V4, sid)
	// The waiting poll gets a noop at once, not after pingInterval
	require.Equal(t, "6", <-polled)
	require.Less(t, time.Since(start), time.Second)

	// Polling is over, and the websocket carries the messages
	status, _ := get(t, url)
	require.Equal(t, http.StatusBadRequest, status)
	require.NoError(t, ws.WritePacket(Packet{Type: Message, Data: "hello"}))
	p, err := ws.ReadPacket()
	require.NoError(t, err)
	require.Equal(t, Packet{Type: Message, Data: "hello"}, p)

	// A session upgrades once
	_, err = DialWebSocket(context.Background(), ts.URL, V4, sid)
	require.ErrorContains(t, err, "400")
}

func TestBadProbeKeepsPolling(t *testing.T) {
	_, ts, _ := echoServer(t, Config{})
	sid := handshake(t, ts.URL, "4")
	ws, err := DialWebSocket(context.Background(), ts.URL, V4, sid)
	require.NoError(t, err)
	defer ws.Close()
	require.NoError(t, ws.WritePacket(Packet{Type: Message, Data: "not a probe"}))
	_, err = ws.ReadPacket()
	require.Error(t, err)

	url := ts.URL + "/?EIO=4&transport=polling&sid=" + sid
	status, _ := post(t, url, "4still polling")
	require.Equal(t, http.StatusOK, status)
	_, body := get(t, url)
	require.Equal(t, "4still polling", body)
}

func TestWebSocketSessionClosesWithoutHeartbeat(t *testing.T) {
	for _, version := range []int{V3, V4} {
		t.Run("v"+string(rune('0'+version)), func(t *testing.T) {
			server, ts, closed := echoServer(t, testConfig)
			// A client that opens a session and never pings (v3) or answers pings (v4)
			ws, err := DialWebSocket(context.Background(), ts.URL, version, "")
			require.NoError(t, err)
			defer ws.Close()
			p, err := ws.ReadPacket()
			require.NoError(t, err)
			require.Equal(t, Open, p.Type)
			require.Equal(t, "ping timeout", <-closed)
			require.Equal(t, 0, server.Count())
		})
	}
}

func TestMaxSessionsOverWebSocket(t *testing.T) {
	_, ts, _ := echoServer(t, Config{MaxSessions: 1})
	handshake(t, ts.URL, "4")
	_, err := DialWebSocket(context.Background(), ts.URL, V4, "")
	require.ErrorContains(t, err, "429")
}
