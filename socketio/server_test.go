package socketio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/socketio/engineio"
)

var versions = map[string]int{"socket.io v5": engineio.V4, "socket.io v2": engineio.V3}

var transports = []string{"polling", "websocket", "upgrade"}

func newTestServer(t *testing.T) (*Server, string) {
	server := NewServer(engineio.Config{PingInterval: 100 * time.Millisecond, PingTimeout: 100 * time.Millisecond})
	ts := httptest.NewServer(server)
	t.Cleanup(func() {
		server.Close()
		ts.Close()
	})
	return server, ts.URL + "/socket.io/"
}

func dial(t *testing.T, url string, version int) *Client {
	return dialOver(t, url, version, "polling")
}

// dialOver connects over polling, over a websocket, or over polling then upgraded to a websocket
func dialOver(t *testing.T, url string, version int, transport string) *Client {
	var client *Client
	var err error
	if transport == "websocket" {
		client, err = DialOverWebSocket(context.Background(), url, version)
	} else {
		client, err = Dial(context.Background(), url, version, nil)
	}
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	if transport == "upgrade" {
		require.NoError(t, client.Upgrade(timeout(t)))
	}
	return client
}

// overEveryTransport runs test for each Socket.IO version over each transport
func overEveryTransport(t *testing.T, test func(t *testing.T, version int, transport string)) {
	for name, version := range versions {
		for _, transport := range transports {
			t.Run(name+" "+transport, func(t *testing.T) { test(t, version, transport) })
		}
	}
}

func timeout(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func receive[T any](t *testing.T, c chan T) T {
	select {
	case v := <-c:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting")
		var zero T
		return zero
	}
}

func TestEventsAndAcks(t *testing.T) {
	overEveryTransport(t, func(t *testing.T, version int, transport string) {
		server, url := newTestServer(t)
		server.Of("/").OnConnection(func(s *Socket) {
			// Echo every hello back, and answer authorize like Nightscout does
			s.On("hello", func(args []json.RawMessage, ack func(...any)) {
				s.Emit("hello back", args[0])
			})
			s.On("authorize", func(args []json.RawMessage, ack func(...any)) {
				ack(map[string]bool{"read": true, "write": false})
			})
		})

		client := dialOver(t, url, version, transport)
		require.NoError(t, client.Connect(timeout(t), "/", nil))
		require.True(t, client.Connected("/"))

		got := make(chan string, 1)
		client.On("/", "hello back", func(args []json.RawMessage) { got <- string(args[0]) })
		require.NoError(t, client.Emit("/", "hello", map[string]int{"sgv": 101}))
		require.JSONEq(t, `{"sgv":101}`, receive(t, got))

		reply, err := client.EmitWithAck(timeout(t), "/", "authorize", map[string]string{"token": "abc"})
		require.NoError(t, err)
		require.Len(t, reply, 1)
		require.JSONEq(t, `{"read":true,"write":false}`, string(reply[0]))
	})
}

func TestEventsOnConnectionArriveWithHandlersSetBeforeOpen(t *testing.T) {
	server, url := newTestServer(t)
	// A Socket.IO v2 server connects the default namespace itself, so it can emit before Dial returns
	server.Of("/").OnConnection(func(s *Socket) { s.Emit("clients", 1) })

	client := NewClient(engineio.V3)
	got := make(chan string, 1)
	client.On("/", "clients", func(args []json.RawMessage) { got <- string(args[0]) })
	require.NoError(t, client.Open(context.Background(), url, nil))
	t.Cleanup(func() { client.Close() })
	require.Equal(t, "1", receive(t, got))
}

func TestNamespaceAuthorization(t *testing.T) {
	server, url := newTestServer(t)
	admin := server.Of("/admin")
	admin.Use(func(s *Socket, auth json.RawMessage) error {
		var a struct{ Token string }
		if json.Unmarshal(auth, &a) != nil || a.Token != "secret" {
			return errors.New("Not authorized")
		}
		return nil
	})
	connected := make(chan string, 1)
	admin.OnConnection(func(s *Socket) { connected <- s.ID() })

	client := dial(t, url, engineio.V4)
	require.EqualError(t, client.Connect(timeout(t), "/admin", map[string]string{"token": "wrong"}), "Not authorized")
	require.Equal(t, 0, admin.Count())

	require.NoError(t, client.Connect(timeout(t), "/admin", map[string]string{"token": "secret"}))
	require.NotEmpty(t, receive(t, connected))
	require.Equal(t, 1, admin.Count())

	require.EqualError(t, client.Connect(timeout(t), "/nope", nil), "Invalid namespace")
}

func TestV2ConnectErrorIsAString(t *testing.T) {
	server, url := newTestServer(t)
	server.Of("/").Use(func(s *Socket, auth json.RawMessage) error {
		// v2 clients send no auth payload; their credentials come in the handshake query
		if s.Request().URL.Query().Get("token") != "secret" {
			return errors.New("Not authorized")
		}
		return nil
	})

	refused := dial(t, url, engineio.V3)
	require.EqualError(t, refused.Connect(timeout(t), "/", nil), "Not authorized")

	allowed := dial(t, url+"?token=secret", engineio.V3)
	require.NoError(t, allowed.Connect(timeout(t), "/", nil))
}

// A Socket.IO v2 server answers the default namespace as soon as the client opens, so a refusal can arrive before
// Connect waits for it; Connect still returns it rather than waiting for its timeout
func TestV2ConnectErrorBeforeConnect(t *testing.T) {
	server, url := newTestServer(t)
	server.Of("/").Use(func(s *Socket, auth json.RawMessage) error { return errors.New("Not authorized") })

	refused := dial(t, url, engineio.V3)
	time.Sleep(200 * time.Millisecond) // the refusal arrives
	require.EqualError(t, refused.Connect(timeout(t), "/", nil), "Not authorized")
}

func TestRoomsAndBroadcast(t *testing.T) {
	overEveryTransport(t, func(t *testing.T, version int, transport string) {
		server, url := newTestServer(t)
		main := server.Of("/")
		main.OnConnection(func(s *Socket) {
			s.On("join", func(args []json.RawMessage, ack func(...any)) {
				s.Join("DataReceivers")
				ack()
			})
		})

		reader, watcher := dialOver(t, url, version, transport), dialOver(t, url, version, transport)
		readerGot, watcherGot := make(chan string, 2), make(chan string, 2)
		for client, got := range map[*Client]chan string{reader: readerGot, watcher: watcherGot} {
			require.NoError(t, client.Connect(timeout(t), "/", nil))
			client.On("/", "dataUpdate", func(args []json.RawMessage) { got <- "dataUpdate" })
			client.On("/", "clients", func(args []json.RawMessage) { got <- "clients " + string(args[0]) })
		}
		_, err := reader.EmitWithAck(timeout(t), "/", "join")
		require.NoError(t, err)

		// Only the room gets dataUpdate; everyone gets clients
		require.NoError(t, main.EmitTo("DataReceivers", "dataUpdate", map[string]any{"delta": true}))
		require.NoError(t, main.Emit("clients", 2))
		require.Equal(t, "dataUpdate", receive(t, readerGot))
		require.Equal(t, "clients 2", receive(t, readerGot))
		require.Equal(t, "clients 2", receive(t, watcherGot))
	})
}

func TestDisconnects(t *testing.T) {
	server, url := newTestServer(t)
	reasons := make(chan string, 3)
	var serverSocket *Socket
	sockets := make(chan *Socket, 3)
	server.Of("/").OnConnection(func(s *Socket) {
		s.OnDisconnect(func(reason string) { reasons <- reason })
		sockets <- s
	})

	// The client leaves the namespace
	client := dial(t, url, engineio.V4)
	require.NoError(t, client.Connect(timeout(t), "/", nil))
	receive(t, sockets)
	require.NoError(t, client.Disconnect("/"))
	require.Equal(t, "client namespace disconnect", receive(t, reasons))

	// The server disconnects the socket, and the client sees it
	require.NoError(t, client.Connect(timeout(t), "/", nil))
	serverSocket = receive(t, sockets)
	serverSocket.Disconnect()
	require.Equal(t, "server namespace disconnect", receive(t, reasons))
	require.Eventually(t, func() bool { return !client.Connected("/") }, 2*time.Second, 10*time.Millisecond)
	require.ErrorIs(t, serverSocket.Emit("gone"), ErrDisconnected)

	// Closing the connection disconnects its sockets
	require.NoError(t, client.Connect(timeout(t), "/", nil))
	receive(t, sockets)
	require.NoError(t, client.Close())
	require.Equal(t, "client close", receive(t, reasons))
	require.Equal(t, 0, server.Of("/").Count())
}

func TestProtocolViolationsCloseTheConnection(t *testing.T) {
	tests := map[string]string{
		"bad packet":           "9",
		"binary event":         `51-["baz",{"_placeholder":true,"num":0}]`,
		"event before connect": `2["hello"]`,
		"event not an array":   `0`, // connects first, then sends a bad event
	}
	for name, first := range tests {
		t.Run(name, func(t *testing.T) {
			_, url := newTestServer(t)
			engine, err := engineio.Dial(context.Background(), url, engineio.V4, nil, nil)
			require.NoError(t, err)
			require.NoError(t, engine.Send(first))
			if first == "0" {
				require.NoError(t, engine.Send(`2{"not":"an array"}`))
			}
			select {
			case <-engine.Done():
			case <-time.After(2 * time.Second):
				t.Fatal("connection stayed open")
			}
		})
	}
}

// A session holds one of the server's MaxSessions, so one that connects to no namespace within connectTimeout is
// closed, and a socket the server closes closes its session only when it is the session's last
func TestSessionsThatHoldNoSocketAreClosed(t *testing.T) {
	timeout := connectTimeout
	connectTimeout = 50 * time.Millisecond
	t.Cleanup(func() { connectTimeout = timeout })
	tests := []struct {
		name string
		// namespaces are those the client connects to, and close is the one whose socket the server closes
		namespaces []string
		close      string
		closed     bool
	}{
		{name: "a session that connects to no namespace is closed", closed: true},
		{name: "a session with a socket stays open", namespaces: []string{"/"}},
		{name: "closing a session's only socket closes it", namespaces: []string{"/"}, close: "/", closed: true},
		{name: "closing one of a session's sockets leaves the other", namespaces: []string{"/", "/alarm"}, close: "/alarm"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, url := newTestServer(t)
			sockets := make(chan *Socket, 2)
			for _, namespace := range []string{"/", "/alarm"} {
				server.Of(namespace).OnConnection(func(s *Socket) { sockets <- s })
			}
			engine, err := engineio.Dial(context.Background(), url, engineio.V4, nil, nil)
			require.NoError(t, err)
			t.Cleanup(func() { engine.Close() })
			for _, namespace := range test.namespaces {
				require.NoError(t, engine.Send(Packet{Type: Connect, Namespace: namespace}.Encode()))
				socket := receive(t, sockets)
				if socket.Namespace().name == test.close {
					socket.Close()
				}
			}
			select {
			case <-engine.Done():
				require.True(t, test.closed, "the session was closed")
			case <-time.After(4 * connectTimeout):
				require.False(t, test.closed, "the session stayed open")
			}
		})
	}
}
