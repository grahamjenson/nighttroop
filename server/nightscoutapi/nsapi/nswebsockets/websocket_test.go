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
	"nighttroop.t1d.nz/data/sql/controldb"
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
	os.Setenv("SITE_UNAUTH_PER_MINUTE_BURST", "4")
	os.Setenv("SOCKET_MAX_SESSIONS_PER_SITE", "4")
	os.Setenv("SOCKET_MAX_UNAUTH_SESSIONS_PER_IP", "2")
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

// The session limits come from the server config: here four sessions a site, two of them an IP's before they are
// given access, the IP keyed as its limits are
func TestSocketLimitsComeFromConfig(t *testing.T) {
	cfg := testConfig
	engineConfig := (&Websocket{config: cfg}).engineConfig()
	ipv6 := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, "site", "2001:db8::1", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	require.Equal(t, "2001:db8::/64", engineConfig.ClientKey(httptest.NewRequestWithContext(ipv6, http.MethodGet, "/", nil)))
	engineConfig.ClientKey = nil
	require.Equal(t, engineio.Config{PingInterval: 25 * time.Second, PingTimeout: 20 * time.Second, MaxPayload: 100000, MaxSessions: 4, MaxBuffered: 1000, MaxBufferedBytes: 2000000, MaxPacketRate: 20, MaxUnauthorized: 2}, engineConfig)

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

// ipSocketServer serves a site's Socket.IO server, as the router would, with each request in the site's context from
// the IP in its ip query, so a test can connect from many IPs; found finds each token's identity
func ipSocketServer(t *testing.T, controlDB *db.ControlDB, siteID string, found IdentityFinder) *httptest.Server {
	ws := NewWebsocket(controlDB, testConfig, nil, found, found)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := httputils.CreateSiteContext(r.Context(), testConfig, controlDB, nil, nil, siteID, r.URL.Query().Get("ip"), httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		ws.ServeHTTP(w, r.WithContext(ctx))
	}))
	t.Cleanup(func() {
		ws.siteServer(siteID, false).Close()
		ts.Close()
	})
	return ts
}

// An authorize before a socket is given access spends its IP's budget on the site, as a request that fails auth does,
// not the site's GET limit, so one IP's failed authorizes, however many sockets it spreads them over, cannot leave the
// site's own pages and followers with 429s
func TestSocketFailedAuthorizesSpendTheIPsBudgetNotTheSites(t *testing.T) {
	siteID := fmt.Sprint("site", runs.Add(1))
	ts := ipSocketServer(t, nil, siteID, noIdentity)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	authorize := func(ip string) error {
		client, err := socketio.Dial(ctx, ts.URL+"/socket.io/?ip="+ip, engineio.V4, nil)
		require.NoError(t, err)
		defer client.Close()
		require.NoError(t, client.Connect(ctx, socketio.DefaultNamespace, nil))
		short, cancelShort := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancelShort()
		reply, err := client.EmitWithAck(short, socketio.DefaultNamespace, "authorize", map[string]any{"secret": "not the secret"})
		if err == nil {
			require.JSONEq(t, `{"read":false,"write":false,"write_treatment":false}`, string(reply[0]))
		}
		return err
	}
	// More failed authorizes than the site's GET burst of 3, within the IP's burst of 4
	for range 4 {
		require.NoError(t, authorize("10.0.0.1"))
	}
	require.InDelta(t, 3, testConfig.RateLimiters.SiteGETLimiter(siteID).Tokens(), 0.5, "the site's GET budget is untouched")
	require.Error(t, authorize("10.0.0.1"), "past its burst the IP's authorize is unanswered")
	require.NoError(t, authorize("10.0.0.2"), "another IP has its own budget")
}

// As in Nightscout, an authorize with a token that is no one's is disconnected unanswered, while one with a wrong
// API secret, or none, is answered that it can do nothing. Either way, once a socket has been refused, a token at
// once and otherwise maxAuthorizeAttempts times, its session is closed, rather than held until authorizeWithin.
func TestSocketAuthorizeWithABadTokenDisconnects(t *testing.T) {
	ts := ipSocketServer(t, nil, fmt.Sprint("site", runs.Add(1)), noIdentity)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i, message := range []map[string]any{{"secret": "not the secret"}, {}, {"token": "not-a-real-token"}} {
		client, err := socketio.Dial(ctx, fmt.Sprintf("%s/socket.io/?ip=10.0.0.%d", ts.URL, i), engineio.V4, nil)
		require.NoError(t, err)
		require.NoError(t, client.Connect(ctx, socketio.DefaultNamespace, nil))
		attempts := maxAuthorizeAttempts
		if message["token"] != nil {
			attempts = 1
		}
		for range attempts {
			short, cancelShort := context.WithTimeout(ctx, 500*time.Millisecond)
			reply, err := client.EmitWithAck(short, socketio.DefaultNamespace, "authorize", message)
			cancelShort()
			if message["token"] == nil {
				require.NoError(t, err)
				require.JSONEq(t, `{"read":false,"write":false,"write_treatment":false}`, string(reply[0]))
			} else {
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.False(t, client.Connected(socketio.DefaultNamespace))
			}
		}
		select {
		case <-client.Done():
		case <-time.After(3 * time.Second):
			t.Fatalf("the session of a socket refused with %v stayed open", message)
		}
		client.Close()
	}
}

// A site's sessions are shared by its followers, so one IP, or one IPv6 /64, may hold only
// SocketMaxUnauthorizedSessionsPerIP of them before they are given access, and cannot take every one
func TestSocketOneIPCannotTakeEverySession(t *testing.T) {
	controlDB, identity := siteIdentity(t)
	found := func(_ *config.ServerConfig, _ context.Context, token string) (*controldb.SiteIdentity, error) {
		if token == "token" {
			return identity, nil
		}
		return nil, errors.New("no identity")
	}
	ts := ipSocketServer(t, controlDB, identity.SiteID, found)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	dial := func(ip string) (*socketio.Client, error) {
		client, err := socketio.Dial(ctx, ts.URL+"/socket.io/?ip="+ip, engineio.V4, nil)
		if err == nil {
			t.Cleanup(func() { client.Close() })
		}
		return client, err
	}

	first, err := dial("2001:db8::1")
	require.NoError(t, err)
	_, err = dial("2001:db8::2")
	require.NoError(t, err)
	_, err = dial("2001:db8::3")
	require.ErrorContains(t, err, "429", "a third session from the same /64 is refused")
	_, err = dial("10.0.0.1")
	require.NoError(t, err, "another IP still connects")

	// A session given access no longer counts against its IP
	require.NoError(t, first.Connect(ctx, "/alarm", nil))
	_, err = first.EmitWithAck(ctx, "/alarm", "subscribe", map[string]any{"accessToken": "token"})
	require.NoError(t, err)
	_, err = dial("2001:db8::3")
	require.NoError(t, err)
	_, err = dial("10.0.0.2")
	require.ErrorContains(t, err, "429", "the site's own cap of 4 still holds")
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
