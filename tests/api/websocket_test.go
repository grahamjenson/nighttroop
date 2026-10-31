package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// websocketClient is a Socket.IO client of the site's /socket.io/, recording the events Nightscout sends
type websocketClient struct {
	*socketio.Client
	events map[string]chan json.RawMessage
}

func dialWebsocket(t *testing.T, ts *httptest.Server, subdomain string, version int) *websocketClient {
	client := socketio.NewClient(version)
	c := &websocketClient{Client: client, events: map[string]chan json.RawMessage{}}
	for _, event := range []string{"connected", "dataUpdate", "clients"} {
		events := make(chan json.RawMessage, 10)
		c.events[event] = events
		client.On(socketio.DefaultNamespace, event, func(args []json.RawMessage) {
			arg := json.RawMessage("null")
			if len(args) > 0 {
				arg = args[0]
			}
			// A server sending more than the test reads must not block the client's reads, which would hang the test
			select {
			case events <- arg:
			default:
			}
		})
	}
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

// authorize emits authorize as xDrip+ and AAPS do and returns the answer
func (c *websocketClient) authorize(t *testing.T, message map[string]any) map[string]any {
	reply, err := c.EmitWithAck(wsTimeout(t), socketio.DefaultNamespace, "authorize", message)
	require.NoError(t, err)
	var auth map[string]any
	require.NoError(t, json.Unmarshal(reply[0], &auth))
	return auth
}

// next is the event's next argument, failing after a few seconds
func (c *websocketClient) next(t *testing.T, event string) map[string]any {
	select {
	case arg := <-c.events[event]:
		var out map[string]any
		json.Unmarshal(arg, &out)
		return out
	case <-time.After(5 * time.Second):
		t.Fatalf("no %s event", event)
		return nil
	}
}

// clients is the next count of connections
func (c *websocketClient) clients(t *testing.T) float64 {
	select {
	case arg := <-c.events["clients"]:
		var count float64
		require.NoError(t, json.Unmarshal(arg, &count))
		return count
	case <-time.After(5 * time.Second):
		t.Fatal("no clients event")
		return 0
	}
}

// none checks the event does not come
func (c *websocketClient) none(t *testing.T, event string) {
	select {
	case arg := <-c.events[event]:
		t.Fatalf("unexpected %s event: %s", event, arg)
	case <-time.After(300 * time.Millisecond):
	}
}

func TestNightscoutWebsocket(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	router := server.GetMainHandler(controlDB, config)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	subdomain := site.SiteDB.Subdomain
	entries := "http://" + subdomain + ".ns.localhost/api/v1/entries?token=" + site.PwdNSToken

	postEntryAt := func(sgv int, at time.Time) {
		status, _, _ := testutils.JSONRequest[any](t, router, http.MethodPost, entries, []map[string]any{
			{"type": "sgv", "sgv": sgv, "date": at.UnixMilli(), "device": "websocket test", "direction": "Flat"},
		})
		require.Equal(t, http.StatusOK, status)
	}
	postEntry := func(sgv int) { postEntryAt(sgv, time.Now()) }
	// Older than the hour of history the clients ask for
	postEntryAt(90, time.Now().Add(-3*time.Hour))
	postEntry(100)

	for name, version := range map[string]int{"socket.io v2": engineio.V3, "socket.io v5": engineio.V4} {
		t.Run(name, func(t *testing.T) {
			client := dialWebsocket(t, ts, subdomain, version)
			require.Positive(t, client.clients(t))

			auth := client.authorize(t, map[string]any{"client": "test", "secret": site.PwdNSToken, "history": 1})
			require.Equal(t, map[string]any{"read": true, "write": true, "write_treatment": true}, auth)
			client.next(t, "connected")

			// The first dataUpdate has the recent data, sgvs as the Nightscout client holds them
			first := client.next(t, "dataUpdate")
			require.NotContains(t, first, "delta")
			mgdls := map[float64]map[string]any{}
			for _, sgv := range first["sgvs"].([]any) {
				mgdls[sgv.(map[string]any)["mgdl"].(float64)] = sgv.(map[string]any)
			}
			require.Contains(t, mgdls, 100.0)
			require.NotContains(t, mgdls, 90.0, "older than the history")
			require.Equal(t, "Flat", mgdls[100]["direction"])
			require.NotEmpty(t, mgdls[100]["mills"])
			for _, key := range []string{"mbgs", "cals", "treatments", "devicestatus", "profiles", "food"} {
				require.Contains(t, first, key)
			}

			// Each write after that comes as a delta
			postEntry(120)
			delta := client.next(t, "dataUpdate")
			require.Equal(t, true, delta["delta"])
			require.EqualValues(t, 120, delta["sgvs"].([]any)[0].(map[string]any)["mgdl"])
		})
	}

	t.Run("read only token can read but not write", func(t *testing.T) {
		client := dialWebsocket(t, ts, subdomain, engineio.V4)
		auth := client.authorize(t, map[string]any{"token": site.ReadOnlyNSToken})
		require.Equal(t, map[string]any{"read": true, "write": false, "write_treatment": false}, auth)
	})

	t.Run("jwt token", func(t *testing.T) {
		jwt, _, err := config.Security.GenerateUserJWTToken(subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
		require.NoError(t, err)
		client := dialWebsocket(t, ts, subdomain, engineio.V4)
		require.Equal(t, true, client.authorize(t, map[string]any{"token": jwt})["read"])
	})

	t.Run("bad token gets no data", func(t *testing.T) {
		client := dialWebsocket(t, ts, subdomain, engineio.V4)
		auth := client.authorize(t, map[string]any{"secret": "not a token"})
		require.Equal(t, map[string]any{"read": false, "write": false, "write_treatment": false}, auth)
		client.next(t, "connected")
		postEntry(130)
		client.none(t, "dataUpdate")
	})

	t.Run("another site's writes are not sent", func(t *testing.T) {
		other, err := controlDB.CreateSite(t.Context())
		require.NoError(t, err)
		client := dialWebsocket(t, ts, other.SiteDB.Subdomain, engineio.V4)
		client.authorize(t, map[string]any{"secret": other.PwdNSToken})
		client.next(t, "dataUpdate")
		postEntry(140)
		client.none(t, "dataUpdate")

		// Nor does another site's token work here
		client = dialWebsocket(t, ts, other.SiteDB.Subdomain, engineio.V4)
		require.Equal(t, false, client.authorize(t, map[string]any{"secret": site.PwdNSToken})["read"])
	})

	t.Run("clients counts the site's connections", func(t *testing.T) {
		first := dialWebsocket(t, ts, subdomain, engineio.V4)
		count := first.clients(t)
		second := dialWebsocket(t, ts, subdomain, engineio.V4)
		require.Equal(t, count+1, first.clients(t))
		require.NoError(t, second.Close())
		require.Equal(t, count, first.clients(t))
	})

	t.Run("too many authorize attempts disconnect", func(t *testing.T) {
		client := dialWebsocket(t, ts, subdomain, engineio.V4)
		for range 3 {
			client.authorize(t, map[string]any{"secret": "guess"})
		}
		require.NoError(t, client.Emit(socketio.DefaultNamespace, "authorize", map[string]any{"secret": "guess"}))
		require.Eventually(t, func() bool { return !client.Connected(socketio.DefaultNamespace) }, 5*time.Second, 10*time.Millisecond)
	})
}
