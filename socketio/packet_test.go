package socketio

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func id(n int64) *int64 { return &n }

// The examples are from the Socket.IO v5 protocol document
func TestPacketRoundTrip(t *testing.T) {
	tests := []struct {
		encoded string
		packet  Packet
	}{
		{"0", Packet{Type: Connect, Namespace: "/"}},
		{`0{"token":"123"}`, Packet{Type: Connect, Namespace: "/", Data: json.RawMessage(`{"token":"123"}`)}},
		{`0/admin,{"sid":"oSO0OpakMV_3jnilAAAA"}`, Packet{Type: Connect, Namespace: "/admin", Data: json.RawMessage(`{"sid":"oSO0OpakMV_3jnilAAAA"}`)}},
		{`4{"message":"Not authorized"}`, Packet{Type: ConnectError, Namespace: "/", Data: json.RawMessage(`{"message":"Not authorized"}`)}},
		{`2["foo"]`, Packet{Type: Event, Namespace: "/", Data: json.RawMessage(`["foo"]`)}},
		{`2/admin,["bar"]`, Packet{Type: Event, Namespace: "/admin", Data: json.RawMessage(`["bar"]`)}},
		{`212["foo"]`, Packet{Type: Event, Namespace: "/", ID: id(12), Data: json.RawMessage(`["foo"]`)}},
		{`3/admin,13["bar"]`, Packet{Type: Ack, Namespace: "/admin", ID: id(13), Data: json.RawMessage(`["bar"]`)}},
		{"31[]", Packet{Type: Ack, Namespace: "/", ID: id(1), Data: json.RawMessage(`[]`)}},
		{"1", Packet{Type: Disconnect, Namespace: "/"}},
		{"1/admin,", Packet{Type: Disconnect, Namespace: "/admin"}},
	}
	for _, test := range tests {
		t.Run(test.encoded, func(t *testing.T) {
			packet, err := DecodePacket(test.encoded)
			require.NoError(t, err)
			require.Equal(t, test.packet, packet)
			require.Equal(t, test.encoded, packet.Encode())
		})
	}
}

func TestDecodePacketWithoutNamespaceComma(t *testing.T) {
	packet, err := DecodePacket("1/admin")
	require.NoError(t, err)
	require.Equal(t, Packet{Type: Disconnect, Namespace: "/admin"}, packet)
}

func TestDecodePacketRejectsBadPackets(t *testing.T) {
	tests := []struct {
		encoded string
		err     error
	}{
		{"", ErrBadPacket},
		{"9", ErrBadPacket},
		{"x", ErrBadPacket},
		{`2["foo"`, ErrBadPacket},
		{"2not json", ErrBadPacket},
		{"299999999999999999999[]", ErrBadPacket},
		{`51-["baz",{"_placeholder":true,"num":0}]`, ErrBinary},
		{`61-15["bar",{"_placeholder":true,"num":0}]`, ErrBinary},
	}
	for _, test := range tests {
		t.Run(test.encoded, func(t *testing.T) {
			_, err := DecodePacket(test.encoded)
			require.ErrorIs(t, err, test.err)
		})
	}
}

func TestEventData(t *testing.T) {
	data, err := EventData("dataUpdate", map[string]int{"sgvs": 1}, "x")
	require.NoError(t, err)
	require.JSONEq(t, `["dataUpdate",{"sgvs":1},"x"]`, string(data))

	event, args, err := ParseEvent(data)
	require.NoError(t, err)
	require.Equal(t, "dataUpdate", event)
	require.Len(t, args, 2)
	require.JSONEq(t, `{"sgvs":1}`, string(args[0]))

	for _, bad := range []string{`[]`, `{}`, `[1]`, `"x"`} {
		_, _, err := ParseEvent(json.RawMessage(bad))
		require.ErrorIs(t, err, ErrBadPacket, bad)
	}
}
