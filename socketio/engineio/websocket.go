package engineio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket is Engine.IO's websocket transport: each packet is one websocket message,
// text packets in text frames and binary packets in binary frames
type WebSocket struct {
	conn    *websocket.Conn
	version int

	write sync.Mutex // gorilla allows one writer at a time
}

// writeTimeout stops a client that never reads from blocking a write forever
const writeTimeout = 10 * time.Second

// UpgradeWebSocket turns an Engine.IO websocket request into a WebSocket for its EIO version.
// A message longer than maxPayload bytes closes the connection; checkOrigin, if not nil, decides which origins may connect.
func UpgradeWebSocket(w http.ResponseWriter, r *http.Request, maxPayload int, checkOrigin func(*http.Request) bool) (*WebSocket, error) {
	version, err := strconv.Atoi(r.URL.Query().Get("EIO"))
	if err != nil || (version != V3 && version != V4) {
		return nil, fmt.Errorf("%w: unsupported EIO version", ErrBadPacket)
	}
	upgrader := websocket.Upgrader{CheckOrigin: checkOrigin}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(int64(maxPayload))
	return &WebSocket{conn: conn, version: version}, nil
}

// DialWebSocket connects to the Engine.IO server at rawURL (e.g. http://host/socket.io/) over websocket.
// sid, when set, joins an existing polling session to upgrade it; empty opens a new session.
func DialWebSocket(ctx context.Context, rawURL string, version int, sid string) (*WebSocket, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)
	query := u.Query()
	query.Set("EIO", strconv.Itoa(version))
	query.Set("transport", "websocket")
	if sid != "" {
		query.Set("sid", sid)
	}
	u.RawQuery = query.Encode()
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("engineio: websocket dial got %d: %w", resp.StatusCode, err)
		}
		return nil, err
	}
	return &WebSocket{conn: conn, version: version}, nil
}

// Version is the Engine.IO protocol version spoken over the connection
func (ws *WebSocket) Version() int { return ws.version }

// ReadPacket waits for the next packet
func (ws *WebSocket) ReadPacket() (Packet, error) {
	messageType, data, err := ws.conn.ReadMessage()
	if err != nil {
		return Packet{}, err
	}
	if messageType == websocket.TextMessage {
		if len(data) == 0 {
			return Packet{}, ErrBadPacket
		}
		p := Packet{Type: PacketType(data[0]), Data: string(data[1:])}
		return p, validType(p.Type)
	}
	// v4 binary frames are message data; v3 ones start with the packet type as a byte
	if ws.version == V3 {
		if len(data) == 0 {
			return Packet{}, ErrBadPacket
		}
		p := Packet{Type: PacketType('0' + data[0]), Binary: data[1:]}
		return p, validType(p.Type)
	}
	return Packet{Type: Message, Binary: data}, nil
}

// WritePacket sends one packet
func (ws *WebSocket) WritePacket(p Packet) error {
	ws.write.Lock()
	defer ws.write.Unlock()
	ws.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if p.Binary == nil {
		return ws.conn.WriteMessage(websocket.TextMessage, []byte(string(p.Type)+p.Data))
	}
	data := p.Binary
	if ws.version == V3 {
		data = append([]byte{byte(p.Type - '0')}, p.Binary...)
	}
	return ws.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close tells the other side the connection is closing, if it is still there, and closes it
func (ws *WebSocket) Close() error {
	ws.write.Lock()
	ws.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
	ws.write.Unlock()
	return ws.conn.Close()
}
