package api_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A store's events stream through the whole server, past its timeouts and middleware, as changes are written
func TestStoreEvents(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	router := server.GetMainHandler(controlDB, config)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	subdomain := site.SiteDB.Subdomain

	// Every host goes to the test server, so the site is chosen by the Host header as in production
	addr := ts.Listener.Addr().String()
	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}}}
	jwt, _, err := config.Security.GenerateUserJWTToken(subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+subdomain+".localhost/data/glucose/events", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+jwt)
	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// A reading written through the Nightscout API is streamed to the glucose store's events as it is created
	status, _, _ := testutils.JSONRequest[any](t, router, http.MethodPost, "http://"+subdomain+".ns.localhost/api/v1/entries?token="+site.PwdNSToken, []map[string]any{
		{"type": "sgv", "sgv": 123, "date": time.Now().UnixMilli(), "device": "events test", "direction": "Flat"},
	})
	require.Equal(t, http.StatusOK, status)

	stream := bufio.NewReader(resp.Body)
	event, err := stream.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "event: created\n", event)
	data, err := stream.ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(data, "data: {"), data)
	require.Contains(t, data, `"created_by_device_name":"events test"`)
}
