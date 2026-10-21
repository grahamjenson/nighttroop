package engineio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// websocketServer upgrades each request and sends every packet it reads back, until the connection fails
func websocketServer(t *testing.T, maxPayload int) (*httptest.Server, chan Packet) {
	received := make(chan Packet, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := UpgradeWebSocket(w, r, maxPayload, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer ws.Close()
		for {
			p, err := ws.ReadPacket()
			if err != nil {
				return
			}
			received <- p
			if ws.WritePacket(p) != nil {
				return
			}
		}
	}))
	t.Cleanup(ts.Close)
	return ts, received
}

func TestWebSocketRoundTrip(t *testing.T) {
	packets := []Packet{
		{Type: Ping, Data: "probe"},
		{Type: Upgrade},
		{Type: Message, Data: "hello € 😀"},
		{Type: Message, Binary: []byte{1, 2, 3, 4}},
		{Type: Close},
	}
	for _, version := range []int{V3, V4} {
		t.Run("v"+string(rune('0'+version)), func(t *testing.T) {
			ts, received := websocketServer(t, 1000)
			ws, err := DialWebSocket(context.Background(), ts.URL+"/socket.io/", version, "")
			require.NoError(t, err)
			defer ws.Close()
			for _, p := range packets {
				require.NoError(t, ws.WritePacket(p))
				require.Equal(t, p, <-received)
				echoed, err := ws.ReadPacket()
				require.NoError(t, err)
				require.Equal(t, p, echoed)
			}
		})
	}
}

// The frames on the wire match the protocol, not just each other
func TestWebSocketFrames(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		messageType int
		frame       []byte
		packet      Packet
	}{
		{"v4 text", "4", websocket.TextMessage, []byte("4hello"), Packet{Type: Message, Data: "hello"}},
		{"v4 binary is message data", "4", websocket.BinaryMessage, []byte{1, 2, 3}, Packet{Type: Message, Binary: []byte{1, 2, 3}}},
		{"v3 text", "3", websocket.TextMessage, []byte("2probe"), Packet{Type: Ping, Data: "probe"}},
		{"v3 binary starts with its type", "3", websocket.BinaryMessage, []byte{4, 1, 2, 3}, Packet{Type: Message, Binary: []byte{1, 2, 3}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts, received := websocketServer(t, 1000)
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+"/?transport=websocket&EIO="+test.version, nil)
			require.NoError(t, err)
			defer conn.Close()

			require.NoError(t, conn.WriteMessage(test.messageType, test.frame))
			require.Equal(t, test.packet, <-received)
			messageType, frame, err := conn.ReadMessage()
			require.NoError(t, err)
			require.Equal(t, test.messageType, messageType)
			require.Equal(t, test.frame, frame)
		})
	}
}

func TestWebSocketRejectsBadInput(t *testing.T) {
	ts, received := websocketServer(t, 20)

	// An unsupported version is refused before upgrading
	_, err := DialWebSocket(context.Background(), ts.URL, 5, "")
	require.ErrorContains(t, err, "400")

	for name, frame := range map[string]string{
		"empty frame":                "",
		"unknown type":               "9hello",
		"over max payload":           "4" + strings.Repeat("x", 100),
		"v3 binary without its type": "",
	} {
		t.Run(name, func(t *testing.T) {
			ws, err := DialWebSocket(context.Background(), ts.URL, V3, "")
			require.NoError(t, err)
			defer ws.Close()
			messageType := websocket.TextMessage
			if name == "v3 binary without its type" {
				messageType = websocket.BinaryMessage
			}
			require.NoError(t, ws.conn.WriteMessage(messageType, []byte(frame)))
			// The server stops reading and closes, so nothing is echoed
			_, err = ws.ReadPacket()
			require.Error(t, err)
			require.Empty(t, received)
		})
	}
}
