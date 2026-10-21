package nightscoutapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// testConfig allows 3 writes a site and identity, then 1 a minute; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_POST_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_POST_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_GET_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_GET_PER_MINUTE_BURST", "3")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

var runs atomic.Int64

// A websocket's events skip the HTTP rate limits, so writes over one must count against them too
func TestSocketWritesAreRateLimited(t *testing.T) {
	cfg := testConfig
	// A new site and identity each run, so -count does not share their limits
	run := runs.Add(1)
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{ID: run, Name: "phone"})

	ws := &Websocket{config: cfg}
	writes := 0
	write := func(context.Context, nsutils.Collection, string, json.RawMessage) any {
		writes++
		return result("success")
	}
	args := []json.RawMessage{json.RawMessage(`{"collection":"treatments","data":{}}`)}
	auth := &authorization{Read: true, Write: true, WriteTreatment: true}

	for range 3 {
		require.Equal(t, result("success"), ws.dbWrite("dbAdd", args, auth, ctx, write))
	}
	require.Equal(t, result("Too many requests"), ws.dbWrite("dbAdd", args, auth, ctx, write))
	require.Equal(t, 3, writes)
}

// An authorize checks a token then reads the recent data, so the site's GET limit counts them, however many sockets
// a client spreads them over; the per-socket attempt limit alone starts again on each reconnect
func TestSocketAuthorizeIsRateLimited(t *testing.T) {
	cfg := testConfig
	run := runs.Add(1)
	siteCtx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)

	ws := &Websocket{config: cfg}
	server := socketio.NewServer(engineio.Config{})
	main := server.Of(socketio.DefaultNamespace)
	main.OnConnection(func(socket *socketio.Socket) { ws.connected(main, socket) })
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r.WithContext(siteCtx))
	}))
	t.Cleanup(func() {
		server.Close()
		ts.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
	require.NoError(t, err)
	second, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
	require.NoError(t, err)
	for _, client := range []*socketio.Client{first, second} {
		require.NoError(t, client.Connect(ctx, socketio.DefaultNamespace, nil))
	}
	// Two authorizes each are within both the per-socket attempts and the site's burst of 3, until the fourth
	for _, client := range []*socketio.Client{first, second, first} {
		reply, err := client.EmitWithAck(ctx, socketio.DefaultNamespace, "authorize", map[string]any{})
		require.NoError(t, err)
		require.JSONEq(t, `{"read":false,"write":false,"write_treatment":false}`, string(reply[0]))
	}
	short, cancelShort := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancelShort()
	_, err = second.EmitWithAck(short, socketio.DefaultNamespace, "authorize", map[string]any{})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.False(t, second.Connected(socketio.DefaultNamespace))
	require.True(t, first.Connected(socketio.DefaultNamespace))
}
