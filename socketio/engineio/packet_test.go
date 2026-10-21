package engineio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayloadRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		version int
		packets []Packet
		payload string
	}{
		{"v4 one message", V4, []Packet{{Type: Message, Data: "hello"}}, "4hello"},
		{"v4 messages and ping", V4, []Packet{{Type: Message, Data: "hello"}, {Type: Ping}, {Type: Message, Data: "world"}}, "4hello\x1e2\x1e4world"},
		{"v4 binary", V4, []Packet{{Type: Message, Data: "hello"}, {Type: Message, Binary: []byte{1, 2, 3, 4}}}, "4hello\x1ebAQIDBA=="},
		{"v4 open", V4, []Packet{{Type: Open, Data: `{"sid":"a"}`}}, `0{"sid":"a"}`},
		{"v3 messages", V3, []Packet{{Type: Message, Data: "hello"}, {Type: Message, Data: "world"}}, "6:4hello6:4world"},
		{"v3 counts characters not bytes", V3, []Packet{{Type: Message, Data: "€"}}, "2:4€"},
		{"v3 counts an emoji as two, as JavaScript does", V3, []Packet{{Type: Message, Data: "😀"}}, "3:4😀"},
		{"v3 binary keeps its type", V3, []Packet{{Type: Message, Data: "€"}, {Type: Message, Binary: []byte{1, 2, 3, 4}}}, "2:4€10:b4AQIDBA=="},
		{"v3 ping and pong", V3, []Packet{{Type: Ping, Data: "probe"}, {Type: Pong}}, "6:2probe1:3"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.payload, EncodePayload(test.packets, test.version))
			packets, err := DecodePayload(test.payload, test.version)
			require.NoError(t, err)
			require.Equal(t, test.packets, packets)
		})
	}
}

func TestDecodePayloadRejectsBadPayloads(t *testing.T) {
	tests := []struct {
		name    string
		version int
		payload string
	}{
		{"v4 empty", V4, ""},
		{"v4 empty packet between separators", V4, "4a\x1e\x1e4b"},
		{"v4 unknown type", V4, "9hello"},
		{"v4 bad base64", V4, "b!!!"},
		{"v3 no length", V3, "4hello"},
		{"v3 length too long", V3, "9:4hello"},
		{"v3 length not a number", V3, "x:4hello"},
		{"v3 zero length", V3, "0:"},
		{"v3 negative length", V3, "-1:4"},
		{"v3 length splits a character", V3, "2:4😀"},
		{"v3 binary without type", V3, "1:b"},
		{"v3 unknown type", V3, "2:9a"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodePayload(test.payload, test.version)
			require.ErrorIs(t, err, ErrBadPacket)
		})
	}
}
