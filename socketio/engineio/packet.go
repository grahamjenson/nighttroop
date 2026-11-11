// Package engineio is the Engine.IO protocol under Socket.IO, over HTTP long-polling only.
// It speaks v4 (EIO=4, Socket.IO v3+ clients) and v3 (EIO=3, Socket.IO v2 clients like xDrip+ and AAPS).
// https://github.com/socketio/engine.io-protocol
package engineio

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

// PacketType is the single digit that starts every Engine.IO packet
type PacketType byte

const (
	Open    PacketType = '0'
	Close   PacketType = '1'
	Ping    PacketType = '2'
	Pong    PacketType = '3'
	Message PacketType = '4'
	Upgrade PacketType = '5'
	Noop    PacketType = '6'
)

// Protocol versions, as sent in the EIO query parameter
const (
	V3 = 3
	V4 = 4
)

// recordSeparator splits packets in a v4 polling payload
const recordSeparator = "\x1e"

var ErrBadPacket = errors.New("engineio: bad packet")

// Packet is one Engine.IO packet. Binary packets carry Binary instead of Data.
type Packet struct {
	Type   PacketType
	Data   string
	Binary []byte
}

// encode is the packet as text; polling sends binary as base64 after a "b"
func (p Packet) encode(version int) string {
	if p.Binary == nil {
		return string(p.Type) + p.Data
	}
	if version == V3 {
		// v3 keeps the packet type after the b
		return "b" + string(p.Type) + base64.StdEncoding.EncodeToString(p.Binary)
	}
	return "b" + base64.StdEncoding.EncodeToString(p.Binary)
}

func decodePacket(s string, version int) (Packet, error) {
	if s == "" {
		return Packet{}, ErrBadPacket
	}
	if s[0] == 'b' {
		packetType := Message
		data := s[1:]
		if version == V3 {
			if data == "" {
				return Packet{}, ErrBadPacket
			}
			packetType, data = PacketType(data[0]), data[1:]
		}
		binary, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return Packet{}, fmt.Errorf("%w: %v", ErrBadPacket, err)
		}
		return Packet{Type: packetType, Binary: binary}, validType(packetType)
	}
	packetType := PacketType(s[0])
	return Packet{Type: packetType, Data: s[1:]}, validType(packetType)
}

func validType(t PacketType) error {
	if t < Open || t > Noop {
		return ErrBadPacket
	}
	return nil
}

// EncodePayload joins packets into one polling body: v4 separates them with \x1e,
// v3 puts each packet's length (in JavaScript string length) and a colon before it
func EncodePayload(packets []Packet, version int) string {
	var b strings.Builder
	for i, p := range packets {
		encoded := p.encode(version)
		if version == V3 {
			b.WriteString(strconv.Itoa(jsLength(encoded)))
			b.WriteByte(':')
		} else if i > 0 {
			b.WriteString(recordSeparator)
		}
		b.WriteString(encoded)
	}
	return b.String()
}

// DecodePayload splits one polling body into its packets
func DecodePayload(payload string, version int) ([]Packet, error) {
	if version != V3 {
		packets := []Packet{}
		for _, s := range strings.Split(payload, recordSeparator) {
			p, err := decodePacket(s, version)
			if err != nil {
				return nil, err
			}
			packets = append(packets, p)
		}
		return packets, nil
	}

	packets := []Packet{}
	for payload != "" {
		colon := strings.IndexByte(payload, ':')
		if colon < 1 {
			return nil, ErrBadPacket
		}
		length, err := strconv.Atoi(payload[:colon])
		if err != nil || length < 1 {
			return nil, ErrBadPacket
		}
		s, rest, ok := cutJSLength(payload[colon+1:], length)
		if !ok {
			return nil, ErrBadPacket
		}
		p, err := decodePacket(s, version)
		if err != nil {
			return nil, err
		}
		packets = append(packets, p)
		payload = rest
	}
	return packets, nil
}

// jsLength is a string's length as JavaScript counts it, in UTF-16 code units, which v3 lengths use
func jsLength(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

// cutJSLength splits s after length UTF-16 code units, or is not ok if s is shorter or a rune would be split
func cutJSLength(s string, length int) (string, string, bool) {
	n := 0
	for i, r := range s {
		if n == length {
			return s[:i], s[i:], true
		}
		n += utf16.RuneLen(r)
	}
	return s, "", n == length
}
