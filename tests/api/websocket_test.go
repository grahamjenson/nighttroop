package api_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/testutils"
)

// websocketClient is a Socket.IO client of the site's /socket.io/, recording the counts of connections Nightscout
// sends as clients
type websocketClient struct {
	*socketio.Client
	clientCounts chan json.RawMessage
}

func dialWebsocket(t *testing.T, ts *httptest.Server, subdomain string) *websocketClient {
	client := socketio.NewClient(engineio.V4)
	c := &websocketClient{Client: client, clientCounts: make(chan json.RawMessage, 10)}
	client.On(socketio.DefaultNamespace, "clients", func(args []json.RawMessage) {
		arg := json.RawMessage("null")
		if len(args) > 0 {
			arg = args[0]
		}
		// A server sending more than the test reads must not block the client's reads, which would hang the test
		select {
		case c.clientCounts <- arg:
		default:
		}
	})
	require.NoError(t, client.Open(wsTimeout(t), "http://"+subdomain+".ns.localhost/socket.io/", testutils.HostClient(ts)))
	t.Cleanup(func() { client.Close() })
	require.NoError(t, client.Connect(wsTimeout(t), socketio.DefaultNamespace, nil))
	return c
}

func wsTimeout(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// clients is the next count of connections
func (c *websocketClient) clients(t *testing.T) float64 {
	select {
	case arg := <-c.clientCounts:
		var count float64
		require.NoError(t, json.Unmarshal(arg, &count))
		return count
	case <-time.After(5 * time.Second):
		t.Fatal("no clients event")
		return 0
	}
}

// TestNightscoutWebsocket checks what needs more than the one connection at a time a .sio file in tests/socket has:
// the authorize answers, the data each token gets and what other sites get are tested there
func TestNightscoutWebsocket(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	ts := httptest.NewServer(server.GetMainHandler(controlDB, config))
	t.Cleanup(ts.Close)
	subdomain := site.SiteDB.Subdomain

	t.Run("clients counts the site's connections", func(t *testing.T) {
		first := dialWebsocket(t, ts, subdomain)
		count := first.clients(t)
		second := dialWebsocket(t, ts, subdomain)
		require.Equal(t, count+1, first.clients(t))
		require.NoError(t, second.Close())
		require.Equal(t, count, first.clients(t))
	})

	t.Run("too many authorize attempts disconnect", func(t *testing.T) {
		client := dialWebsocket(t, ts, subdomain)
		for range 3 {
			_, err := client.EmitWithAck(wsTimeout(t), socketio.DefaultNamespace, "authorize", map[string]any{"secret": "guess"})
			require.NoError(t, err)
		}
		require.NoError(t, client.Emit(socketio.DefaultNamespace, "authorize", map[string]any{"secret": "guess"}))
		require.Eventually(t, func() bool { return !client.Connected(socketio.DefaultNamespace) }, 5*time.Second, 10*time.Millisecond)
	})
}
