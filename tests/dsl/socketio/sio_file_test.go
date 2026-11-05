package socketio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/tests/dsl/script"
)

func TestParseFile(t *testing.T) {
	file, err := parse(`
# a comment
NAME connect
CONNECT http://{{ site }}/socket.io/?EIO=3
T 2s

NAME authorize
EMIT authorize
B {
  "secret": "{{ token }}"
}
ACK
ASSERT_EQUALS true "@read"
SET read {{ .authorize.read }}

RUN_BACKGROUND_JOBS

NAME updates
COLLECT dataUpdate
ASSERT_LENGTH 1 "@"

NAME first
WAIT dataUpdate

NAME post
POST http://{{ site }}/api/v1/entries
H api-secret: {{ token }}
B []
ASSERT_STATUS 200

NAME bye
DISCONNECT
`)
	require.NoError(t, err)
	require.Len(t, file.Steps, 7)

	connect, authorize, jobs := file.Steps[0], file.Steps[1], file.Steps[2]
	require.Equal(t, "CONNECT", connect.Verb)
	require.Equal(t, "http://{{ site }}/socket.io/?EIO=3", connect.Target)
	require.Equal(t, 2*time.Second, connect.Timeout)
	require.Equal(t, "EMIT", authorize.Verb)
	require.Equal(t, "authorize", authorize.Target)
	require.True(t, authorize.Ack)
	require.Equal(t, "{\n\"secret\": \"{{ token }}\"\n}", *authorize.Body)
	require.Len(t, authorize.Asserts, 1)
	require.Len(t, authorize.Sets, 1)
	require.Equal(t, "RUN_BACKGROUND_JOBS", jobs.Verb)
	require.Equal(t, "{{ token }}", file.Steps[5].Headers.Get("api-secret"))
	require.Equal(t, "DISCONNECT", file.Steps[6].Verb)
	require.Equal(t, []string{"dataUpdate"}, file.Events())
}

func TestParseFileErrors(t *testing.T) {
	tests := map[string]string{
		"line before NAME":         "CONNECT http://x/socket.io/",
		"step without a verb":      "NAME empty\nT 1s",
		"two verbs":                "NAME two\nWAIT a\nWAIT b",
		"ACK without EMIT":         "NAME wait\nWAIT dataUpdate\nACK",
		"bad T":                    "NAME wait\nWAIT dataUpdate\nT soon",
		"negative T":               "NAME wait\nWAIT dataUpdate\nT -1s",
		"WAIT without an event":    "NAME wait\nWAIT",
		"DISCONNECT with a target": "NAME bye\nDISCONNECT now",
		"bad assert":               "NAME wait\nWAIT dataUpdate\nASSERT_EQUALS true @read",
		"unknown line":             "NAME wait\nWAIT dataUpdate\nLISTEN dataUpdate",
		"empty NAME":               "NAME",
		"bad header":               "NAME get\nGET http://x/\nH no colon",
	}
	for name, contents := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parse(contents)
			require.ErrorContains(t, err, "test.sio")
		})
	}
}

// parse is the steps of a .sio file's contents
func parse(contents string) (*File, error) {
	lines, err := script.SplitLines([]byte(contents), "test.sio", nil)
	if err != nil {
		return nil, err
	}
	return ParseLines(lines, "test.sio")
}
