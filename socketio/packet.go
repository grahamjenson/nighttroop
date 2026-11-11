// Package socketio is a Socket.IO server and client, built on Engine.IO long-polling, so
// Nightscout clients (the web app, xDrip+, AAPS) can get live updates from NightTroop.
// It speaks Socket.IO v5 (EIO=4) and v4 (EIO=3, Socket.IO v2 clients). Binary packets are not supported.
// https://github.com/socketio/socket.io/blob/main/docs/socket.io-protocol/v5-current.md
package socketio

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// PacketType is the digit that starts every Socket.IO packet
type PacketType byte

const (
	Connect      PacketType = '0'
	Disconnect   PacketType = '1'
	Event        PacketType = '2'
	Ack          PacketType = '3'
	ConnectError PacketType = '4'
	BinaryEvent  PacketType = '5'
	BinaryAck    PacketType = '6'
)

// DefaultNamespace is the namespace a packet is in when it names none
const DefaultNamespace = "/"

var (
	ErrBadPacket = errors.New("socketio: bad packet")
	ErrBinary    = errors.New("socketio: binary packets are not supported")
)

// Packet is one Socket.IO packet, sent as the data of an Engine.IO message
type Packet struct {
	Type      PacketType
	Namespace string
	// ID asks for, or answers, an acknowledgement when set
	ID   *int64
	Data json.RawMessage
}

// Encode is the packet as <type>[<namespace>,][<ack id>][<json>]
func (p Packet) Encode() string {
	var b strings.Builder
	b.WriteByte(byte(p.Type))
	if p.Namespace != "" && p.Namespace != DefaultNamespace {
		b.WriteString(p.Namespace)
		b.WriteByte(',')
	}
	if p.ID != nil {
		b.WriteString(strconv.FormatInt(*p.ID, 10))
	}
	b.Write(p.Data)
	return b.String()
}

// DecodePacket parses one packet; its data must be JSON when there is any
func DecodePacket(s string) (Packet, error) {
	if s == "" {
		return Packet{}, ErrBadPacket
	}
	p := Packet{Type: PacketType(s[0]), Namespace: DefaultNamespace}
	if p.Type == BinaryEvent || p.Type == BinaryAck {
		return Packet{}, ErrBinary
	}
	if p.Type < Connect || p.Type > ConnectError {
		return Packet{}, ErrBadPacket
	}
	rest := s[1:]
	if strings.HasPrefix(rest, "/") {
		namespace, after, found := strings.Cut(rest, ",")
		if !found {
			// A namespace alone, e.g. "1/admin", is allowed without its comma
			namespace, after = rest, ""
		}
		p.Namespace, rest = namespace, after
	}
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits > 0 {
		id, err := strconv.ParseInt(rest[:digits], 10, 64)
		if err != nil {
			return Packet{}, ErrBadPacket
		}
		p.ID, rest = &id, rest[digits:]
	}
	if rest != "" {
		if !json.Valid([]byte(rest)) {
			return Packet{}, ErrBadPacket
		}
		p.Data = json.RawMessage(rest)
	}
	return p, nil
}

// EventData is an event packet's data: its name then its arguments
func EventData(event string, args ...any) (json.RawMessage, error) {
	return json.Marshal(append([]any{event}, args...))
}

// ParseEvent splits an event packet's data into its name and arguments
func ParseEvent(data json.RawMessage) (string, []json.RawMessage, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(data, &parts); err != nil || len(parts) == 0 {
		return "", nil, ErrBadPacket
	}
	var event string
	if err := json.Unmarshal(parts[0], &event); err != nil {
		return "", nil, ErrBadPacket
	}
	return event, parts[1:], nil
}
