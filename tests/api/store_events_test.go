package api_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/alarms"
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

	client := testutils.HostClient(ts)
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

// The summary's events stream when one of the identity's alarms changes, as when it is snoozed, but not another's
func TestSummaryEventsOnAlarmChanges(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	router := server.GetMainHandler(controlDB, config)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	subdomain := site.SiteDB.Subdomain
	client := testutils.HostClient(ts)
	jwt, _, err := config.Security.GenerateUserJWTToken(subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	require.NoError(t, err)

	identities, err := controlDB.Queries.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: subdomain})
	require.NoError(t, err)
	var pwd, other int64
	for _, identity := range identities {
		if identity.Name == consts.FIRST_PWD_NAME {
			pwd = identity.ID
		} else {
			other = identity.ID
		}
	}
	at := time.Now().UnixMilli()
	require.NoError(t, controlDB.Live.Queries.UpsertAlarm(t.Context(), livedb.UpsertAlarmParams{
		SiteID: subdomain, IdentityID: pwd, Type: alarms.UrgentLow, ActiveSince: at, NotifiedAt: at, Alerting: true}))

	request := func(method, path string) *http.Response {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)
		req, err := http.NewRequestWithContext(ctx, method, "http://"+subdomain+".localhost"+path, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+jwt)
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { resp.Body.Close() })
		require.Equal(t, http.StatusOK, resp.StatusCode)
		return resp
	}
	stream := bufio.NewReader(request(http.MethodGet, "/data/summary/events").Body)

	// Another identity's alarm changing is not streamed; snoozing its own is, as the summary has its alarms
	config.Events.Publish(subdomain, events.ModelChange(subdomain, events.Updated, &livedb.Alarm{SiteID: subdomain, IdentityID: other, Type: alarms.Low}))
	request(http.MethodDelete, "/data/alarms/urgent_low")
	event, err := stream.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "event: deleted\n", event)
}

// An identity streaming its push registrations' changes is sent its own, not another identity's apps and devices
func TestPushRegistrationEventsStreamToTheirIdentityAlone(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	ts := httptest.NewServer(server.GetMainHandler(controlDB, config))
	t.Cleanup(ts.Close)
	subdomain := site.SiteDB.Subdomain
	client := testutils.HostClient(ts)

	request := func(identity string, role consts.IdentityRole, method, path, body string) *http.Response {
		jwt, _, err := config.Security.GenerateUserJWTToken(subdomain, identity, string(role))
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)
		req, err := http.NewRequestWithContext(ctx, method, "http://"+subdomain+".localhost"+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+jwt)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}
	register := func(identity string, role consts.IdentityRole, bundleID, token string) {
		resp := request(identity, role, http.MethodPost, "/settings/push-registrations", `{"app": "loop", "provider": "apns", "bundle_id": "`+
			bundleID+`", "production": true, "token": "`+token+`"}`)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp := request(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, http.MethodGet, "/settings/push-registrations/events", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	stream := bufio.NewReader(resp.Body)

	// The phone registering its Loop is not streamed to the read only device; the device registering its own is
	register(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, "com.PHONE12345.loopkit.Loop", strings.Repeat("a1", 32))
	register(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, "com.KITCHEN123.loopkit.Loop", strings.Repeat("b2", 32))
	event, err := stream.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "event: created\n", event)
	data, err := stream.ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, data, `"bundle_id":"com.KITCHEN123.loopkit.Loop"`)
}
