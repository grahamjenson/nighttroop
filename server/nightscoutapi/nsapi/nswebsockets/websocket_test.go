package nswebsockets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// testConfig allows 3 writes a site and identity, then 1 a minute; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_POST_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_POST_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_GET_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_GET_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_BURST", "2")
	os.Setenv("SOCKET_MAX_SESSIONS_PER_SITE", "2")
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
	controlDB, identity := siteIdentity(t)
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, identity.SiteID, "127.0.0.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	ctx = httputils.SetIdentityModelInSiteContext(ctx, identity)

	ws := &Websocket{config: cfg, controlDB: controlDB}
	writes := 0
	write := func(context.Context, nsutils.Collection, string, json.RawMessage) any {
		writes++
		return ackResult("success")
	}
	args := []json.RawMessage{json.RawMessage(`{"collection":"treatments","data":{}}`)}
	auth := &authorization{Read: true, Write: true, WriteTreatment: true}

	for range 3 {
		require.Equal(t, ackResult("success"), ws.handleDBWrite("dbAdd", args, auth, ctx, write))
	}
	require.Equal(t, ackResult("Too many requests"), ws.handleDBWrite("dbAdd", args, auth, ctx, write))
	require.Equal(t, 3, writes)
}

// testControlDB is the control DB of every test's sites, so each identity has its own ID and rate limits
var testControlDB = sync.OnceValues(func() (*db.ControlDB, error) { return db.NewControlDB(context.Background(), testConfig) })

// siteIdentity is a new site's read only device identity
func siteIdentity(t *testing.T) (*db.ControlDB, *controldb.SiteIdentity) {
	controlDB, err := testControlDB()
	require.NoError(t, err)
	site, err := controlDB.CreateSite(t.Context())
	require.NoError(t, err)
	identity, err := controlDB.Queries.GetSiteIdentity(t.Context(), controldb.GetSiteIdentityParams{SiteID: site.SiteDB.Subdomain, Name: consts.FIRST_READ_ONLY_NAME})
	require.NoError(t, err)
	return controlDB, identity
}

// A socket is only authenticated as it authorizes, so once its identity is revoked its writes are refused and it is
// disconnected, rather than it reading and writing the site's data until it next reconnects
func TestSocketOfARevokedIdentityIsRefusedAndDisconnected(t *testing.T) {
	recheck := identityRecheck
	identityRecheck = 20 * time.Millisecond
	t.Cleanup(func() { identityRecheck = recheck })
	revocations := map[string]func(*db.ControlDB, *controldb.SiteIdentity) error{
		"disabled": func(controlDB *db.ControlDB, identity *controldb.SiteIdentity) error {
			return controlDB.Queries.UpdateSiteIdentityEnabled(t.Context(), controldb.UpdateSiteIdentityEnabledParams{Enabled: false, SiteID: identity.SiteID, ID: identity.ID})
		},
		"given a new password": func(controlDB *db.ControlDB, identity *controldb.SiteIdentity) error {
			_, err := controlDB.ResetSiteIdentityPassword(t.Context(), identity, "admin")
			return err
		},
		"deleted": func(controlDB *db.ControlDB, identity *controldb.SiteIdentity) error {
			return controlDB.Queries.DeleteSiteIdentity(t.Context(), controldb.DeleteSiteIdentityParams{SiteID: identity.SiteID, ID: identity.ID})
		},
	}
	for name, revoke := range revocations {
		t.Run(name, func(t *testing.T) {
			controlDB, identity := siteIdentity(t)
			siteCtx := httputils.CreateSiteContext(context.Background(), testConfig, controlDB, nil, nil, identity.SiteID, "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			found := func(*config.ServerConfig, context.Context, string) (*controldb.SiteIdentity, error) {
				return identity, nil
			}
			ws := NewWebsocket(controlDB, testConfig, nil, found, found)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ws.ServeHTTP(w, r.WithContext(siteCtx)) }))
			t.Cleanup(func() {
				ws.siteServer(identity.SiteID, false).Close()
				ts.Close()
			})

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			client, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
			require.NoError(t, err)
			// The default namespace's authorize reads the recent data, which this test's site has no DB for
			require.NoError(t, client.Connect(ctx, "/alarm", nil))
			reply, err := client.EmitWithAck(ctx, "/alarm", "subscribe", map[string]any{"accessToken": "token"})
			require.NoError(t, err)
			require.JSONEq(t, `{"success":true,"message":"Subscribed for alarms"}`, string(reply[0]))
			time.Sleep(3 * identityRecheck)
			require.True(t, client.Connected("/alarm"), "a socket whose identity is not revoked stays connected")
			require.NoError(t, revoke(controlDB, identity))

			args := []json.RawMessage{json.RawMessage(`{"collection":"treatments","data":{}}`)}
			write := func(context.Context, nsutils.Collection, string, json.RawMessage) any { return ackResult("success") }
			writeCtx := httputils.SetIdentityModelInSiteContext(httputils.CreateSiteContext(context.Background(), testConfig, controlDB, nil, nil, identity.SiteID, "127.0.0.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON), identity)
			require.Equal(t, ackResult("Not authorized"), ws.handleDBWrite("dbAdd", args, &authorization{Write: true, WriteTreatment: true}, writeCtx, write))
			require.Eventually(t, func() bool { return !client.Connected("/alarm") }, 2*time.Second, 10*time.Millisecond)
		})
	}
}

// noIdentity finds no identity, as the Nightscout API does for these tests' empty tokens
func noIdentity(*config.ServerConfig, context.Context, string) (*controldb.SiteIdentity, error) {
	return nil, errors.New("no identity")
}

// siteSocketServer serves a site's Socket.IO server, as the router would, with every request in its site context
func siteSocketServer(t *testing.T, cfg *config.ServerConfig, siteCtx context.Context) *httptest.Server {
	ws := NewWebsocket(nil, cfg, nil, noIdentity, noIdentity)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeHTTP(w, r.WithContext(siteCtx))
	}))
	t.Cleanup(func() {
		ws.siteServer(httputils.GetSiteContext(siteCtx).SiteID, false).Close()
		ts.Close()
	})
	return ts
}

// The session limits come from the server config: here two sessions a site
func TestSocketLimitsComeFromConfig(t *testing.T) {
	cfg := testConfig
	require.Equal(t, engineio.Config{PingInterval: 25 * time.Second, PingTimeout: 20 * time.Second, MaxPayload: 100000, MaxSessions: 2, MaxBuffered: 1000, MaxBufferedBytes: 2000000, MaxPacketRate: 20}, (&Websocket{config: cfg}).engineConfig())

	run := runs.Add(1)
	siteCtx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	ts := siteSocketServer(t, cfg, siteCtx)
	for _, status := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		resp, err := http.Get(ts.URL + "/socket.io/?EIO=4&transport=polling")
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, status, resp.StatusCode)
	}
}

// An authorize checks a token then reads the recent data, so the site's GET limit counts them, however many sockets
// a client spreads them over; the per-socket attempt limit alone starts again on each reconnect
func TestSocketAuthorizeIsRateLimited(t *testing.T) {
	cfg := testConfig
	run := runs.Add(1)
	siteCtx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)

	ts := siteSocketServer(t, cfg, siteCtx)

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

// As in Nightscout, an authorize with a token that is no one's is disconnected unanswered, while one with a wrong
// API secret, or none, is answered that it can do nothing
func TestSocketAuthorizeWithABadTokenDisconnects(t *testing.T) {
	run := runs.Add(1)
	siteCtx := httputils.CreateSiteContext(context.Background(), testConfig, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	ts := siteSocketServer(t, testConfig, siteCtx)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, message := range []map[string]any{{"secret": "not the secret"}, {"token": "not-a-real-token"}} {
		client, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
		require.NoError(t, err)
		require.NoError(t, client.Connect(ctx, socketio.DefaultNamespace, nil))
		short, cancelShort := context.WithTimeout(ctx, 500*time.Millisecond)
		reply, err := client.EmitWithAck(short, socketio.DefaultNamespace, "authorize", message)
		cancelShort()
		if message["token"] == nil {
			require.NoError(t, err)
			require.JSONEq(t, `{"read":false,"write":false,"write_treatment":false}`, string(reply[0]))
			require.True(t, client.Connected(socketio.DefaultNamespace))
		} else {
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.False(t, client.Connected(socketio.DefaultNamespace))
		}
		client.Close()
	}
}

// A site has SocketMaxSessionsPerSite sessions, so a socket not given access within authorizeWithin is closed, rather
// than a client opening them all and never authorizing, or authorizing with nothing, to keep the site's followers out
func TestSocketNotGivenAccessIsClosed(t *testing.T) {
	within := authorizeWithin
	authorizeWithin = 50 * time.Millisecond
	t.Cleanup(func() { authorizeWithin = within })
	tests := []struct {
		name      string
		namespace string
		// event and message are what the client sends as it connects, if anything
		event   string
		message map[string]any
		closed  bool
	}{
		{name: "never authorizing", namespace: socketio.DefaultNamespace, closed: true},
		{name: "authorizing with a wrong API secret", namespace: socketio.DefaultNamespace, event: "authorize", message: map[string]any{"secret": "not the secret"}, closed: true},
		{name: "never subscribing to /storage", namespace: "/storage", closed: true},
		{name: "never subscribing to /alarm", namespace: "/alarm", closed: true},
		{name: "subscribing to /alarm", namespace: "/alarm", event: "subscribe", message: map[string]any{"accessToken": "token"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controlDB, identity := siteIdentity(t)
			siteCtx := httputils.CreateSiteContext(context.Background(), testConfig, controlDB, nil, nil, identity.SiteID, "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			found := func(_ *config.ServerConfig, _ context.Context, token string) (*controldb.SiteIdentity, error) {
				if token == "token" {
					return identity, nil
				}
				return nil, errors.New("no identity")
			}
			ws := NewWebsocket(controlDB, testConfig, nil, found, found)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ws.ServeHTTP(w, r.WithContext(siteCtx)) }))
			t.Cleanup(func() {
				ws.siteServer(identity.SiteID, false).Close()
				ts.Close()
			})

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			client, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
			require.NoError(t, err)
			t.Cleanup(func() { client.Close() })
			require.NoError(t, client.Connect(ctx, test.namespace, nil))
			if test.event != "" {
				_, err := client.EmitWithAck(ctx, test.namespace, test.event, test.message)
				require.NoError(t, err)
			}
			select {
			case <-client.Done():
				require.True(t, test.closed, "the socket's session was closed")
			case <-time.After(6 * authorizeWithin):
				require.False(t, test.closed, "the socket's session stayed open")
				require.True(t, client.Connected(test.namespace))
			}
		})
	}
}

// A site's Socket.IO server and its event subscription are freed once it has no sessions, or once the site is deleted
// or disabled, rather than kept for every site ever connected to until restart, and a client connecting after gets a
// new server that works
func TestSocketServerIsFreedOnceItHasNoSessions(t *testing.T) {
	sweep := idleServerSweep
	idleServerSweep = 20 * time.Millisecond
	t.Cleanup(func() { idleServerSweep = sweep })
	tests := map[string]func(siteID string, client *socketio.Client){
		"its last session closing": func(_ string, client *socketio.Client) { client.Close() },
		"its site deleted": func(siteID string, _ *socketio.Client) {
			testConfig.Events.Publish(siteID, events.Event{Name: events.SiteDeleted, SiteID: siteID})
		},
		"its site disabled": func(siteID string, _ *socketio.Client) {
			testConfig.Events.Publish(siteID, events.Event{Name: events.SiteDisabled, SiteID: siteID})
		},
	}
	for name, end := range tests {
		t.Run(name, func(t *testing.T) {
			controlDB, identity := siteIdentity(t)
			siteCtx := httputils.CreateSiteContext(context.Background(), testConfig, controlDB, nil, nil, identity.SiteID, "127.0.0.1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			found := func(*config.ServerConfig, context.Context, string) (*controldb.SiteIdentity, error) { return identity, nil }
			ws := NewWebsocket(controlDB, testConfig, nil, found, found)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ws.ServeHTTP(w, r.WithContext(siteCtx)) }))
			t.Cleanup(ts.Close)

			// subscribe connects a client to /alarm, where it is sent the site's announcements
			subscribe := func() (*socketio.Client, chan struct{}) {
				ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
				defer cancel()
				client, err := socketio.Dial(ctx, ts.URL+"/socket.io/", engineio.V4, nil)
				require.NoError(t, err)
				t.Cleanup(func() { client.Close() })
				announced := make(chan struct{}, 10)
				client.On("/alarm", "announcement", func([]json.RawMessage) { announced <- struct{}{} })
				require.NoError(t, client.Connect(ctx, "/alarm", nil))
				_, err = client.EmitWithAck(ctx, "/alarm", "subscribe", map[string]any{"accessToken": "token"})
				require.NoError(t, err)
				return client, announced
			}
			announce := func(announced chan struct{}) {
				testConfig.Events.Publish(identity.SiteID, events.Event{SiteID: identity.SiteID, Action: events.Created,
					Model: &sitedb.EventGroupsBundle{NotificationEvent: &sitedb.NotificationEvent{NotificationType: "announcement", Title: "Hello"}}})
				select {
				case <-announced:
				case <-time.After(2 * time.Second):
					require.Fail(t, "the client was not sent the announcement")
				}
			}

			client, announced := subscribe()
			announce(announced)
			ws.mu.Lock()
			first := ws.servers[identity.SiteID]
			var unsubscribed atomic.Bool
			unsubscribe := first.unsubscribe
			first.unsubscribe = func() { unsubscribe(); unsubscribed.Store(true) }
			ws.mu.Unlock()
			time.Sleep(5 * idleServerSweep)
			require.Same(t, first, ws.siteServer(identity.SiteID, false), "a server with a session is kept")

			end(identity.SiteID, client)
			select {
			case <-client.Done():
			case <-time.After(2 * time.Second):
				require.Fail(t, "the client's session was not closed")
			}
			require.Eventually(t, func() bool { return ws.siteServer(identity.SiteID, false) == nil && unsubscribed.Load() }, 2*time.Second, 10*time.Millisecond,
				"the server is freed and unsubscribed")

			_, announced = subscribe()
			require.NotSame(t, first, ws.siteServer(identity.SiteID, false), "a new client gets a new server")
			announce(announced)
		})
	}
}
