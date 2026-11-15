package api_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// eventsSite is a new site served through the whole server, past its timeouts and middleware, for streaming events
type eventsSite struct {
	t         *testing.T
	config    *config.ServerConfig
	controlDB *db.ControlDB
	site      *db.SiteCreateResult
	router    http.Handler
	client    *http.Client
}

func newEventsSite(t *testing.T) *eventsSite {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	router := server.GetMainHandler(controlDB, config)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	return &eventsSite{t: t, config: config, controlDB: controlDB, site: site, router: router, client: testutils.HostClient(ts)}
}

// request sends a JSON request to the site's path as one of its identities, which has 5 seconds to finish, as an
// event stream it opens is read
func (s *eventsSite) request(identity string, role consts.IdentityRole, method, path, body string) *http.Response {
	t := s.t
	jwt, _, err := s.config.Security.GenerateUserJWTToken(s.site.SiteDB.Subdomain, identity, string(role))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, method, "http://"+s.site.SiteDB.Subdomain+".localhost"+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// stream opens the site's event stream at path as one of its identities
func (s *eventsSite) stream(identity string, role consts.IdentityRole, path string) *bufio.Reader {
	resp := s.request(identity, role, http.MethodGet, path, "")
	require.Equal(s.t, http.StatusOK, resp.StatusCode)
	require.Equal(s.t, "text/event-stream", resp.Header.Get("Content-Type"))
	return bufio.NewReader(resp.Body)
}

// nextEvent is the stream's next event and its data, read with the blank line that ends them
func nextEvent(t *testing.T, stream *bufio.Reader) (string, string) {
	lines := make([]string, 3)
	for i := range lines {
		line, err := stream.ReadString('\n')
		require.NoError(t, err)
		lines[i] = line
	}
	require.Equal(t, "\n", lines[2])
	return lines[0], lines[1]
}

// A store's events stream as changes are written, here a reading written through the Nightscout API
func TestStoreEvents(t *testing.T) {
	s := newEventsSite(t)
	stream := s.stream(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, "/data/glucose/events")

	status, _, _ := testutils.JSONRequest[any](t, s.router, http.MethodPost, "http://"+s.site.SiteDB.Subdomain+".ns.localhost/api/v1/entries?token="+s.site.PwdNSToken, []map[string]any{
		{"type": "sgv", "sgv": 123, "date": time.Now().UnixMilli(), "device": "events test", "direction": "Flat"},
	})
	require.Equal(t, http.StatusOK, status)

	event, data := nextEvent(t, stream)
	require.Equal(t, "event: created\n", event)
	require.True(t, strings.HasPrefix(data, "data: {"), data)
	require.Contains(t, data, `"created_by_device_name":"events test"`)
}

// The summary's events stream when one of the identity's alarms changes, as when it is snoozed, but not another's
func TestSummaryEventsOnAlarmChanges(t *testing.T) {
	s := newEventsSite(t)
	subdomain := s.site.SiteDB.Subdomain
	identities, err := s.controlDB.Queries.AllSiteIdentities(t.Context(), controldb.AllSiteIdentitiesParams{SiteID: subdomain})
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
	require.NoError(t, s.controlDB.Live.Queries.UpsertAlarm(t.Context(), livedb.UpsertAlarmParams{
		SiteID: subdomain, IdentityID: pwd, Type: alarms.UrgentLow, ActiveSince: at, NotifiedAt: at, Alerting: true}))
	stream := s.stream(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, "/data/summary/events")

	// Another identity's alarm changing is not streamed; snoozing its own is, as the summary has its alarms
	s.config.Events.Publish(subdomain, events.ModelChange(subdomain, events.Updated, &livedb.Alarm{SiteID: subdomain, IdentityID: other, Type: alarms.Low}))
	resp := s.request(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, http.MethodDelete, "/data/alarms/urgent_low", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	event, _ := nextEvent(t, stream)
	require.Equal(t, "event: deleted\n", event)
}

// An identity streaming its push registrations' changes is sent its own, not another identity's apps and devices
func TestPushRegistrationEventsStreamToTheirIdentityAlone(t *testing.T) {
	s := newEventsSite(t)
	register := func(identity string, role consts.IdentityRole, app, bundleID, token string) {
		resp := s.request(identity, role, http.MethodPost, "/settings/push-registrations", `{"app": "`+app+`", "provider": "apns", "bundle_id": "`+
			bundleID+`", "production": true, "token": "`+token+`"}`)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}
	stream := s.stream(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, "/settings/push-registrations/events")

	// The phone registering its Loop is not streamed to the read only device; the device registering its LoopFollow is
	register(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, "loop", "com.PHONE12345.loopkit.Loop", strings.Repeat("a1", 32))
	register(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, "loopfollow", "com.KITCHEN123.LoopFollow", strings.Repeat("b2", 32))
	event, data := nextEvent(t, stream)
	require.Equal(t, "event: created\n", event)
	require.Contains(t, data, `"bundle_id":"com.KITCHEN123.LoopFollow"`)
}

// An identity streaming its settings' changes is sent its own, not another identity's, and told the site's changed,
// which it inherits, without being sent them unless it is an admin
func TestSettingsEventsStreamToTheirIdentityAlone(t *testing.T) {
	s := newEventsSite(t)
	stream := s.stream(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, "/settings/identity/events")
	change := func(identity string, role consts.IdentityRole, path, body string) {
		require.Equal(t, http.StatusOK, s.request(identity, role, http.MethodPost, path, body).StatusCode)
	}

	// The admin changing the site's settings is streamed to the read only device as {}; the phone changing its own is
	// not; the device changing its own is, with them
	change(consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin, "/settings/site", `{"timezone": "Pacific/Chatham"}`)
	change(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, "/settings/identity", `{"timezone": "Asia/Tokyo"}`)
	change(consts.FIRST_READ_ONLY_NAME, consts.IdentityRoleDeviceRead, "/settings/identity", `{"timezone": "Europe/Paris"}`)
	event, data := nextEvent(t, stream)
	require.Equal(t, "event: updated\n", event)
	require.Equal(t, "data: {}\n", data)
	event, data = nextEvent(t, stream)
	require.Equal(t, "event: updated\n", event)
	require.Contains(t, data, `"timezone":"Europe/Paris"`)
}

// A day a record moved away from, through the site API, or was deleted from for good, through the Nightscout API, is
// marked stale in the site's DB as it is written, so its summary is built again after a restart too
func TestMovesAndPermanentDeletesMarkTheirDayStale(t *testing.T) {
	s := newEventsSite(t)
	ctx, now := t.Context(), time.Now()
	day := now.AddDate(0, 0, -3)
	q := s.site.SiteDB.Queries()
	stale := func() []int64 {
		starts, err := q.StaleSummaryDays(ctx)
		require.NoError(t, err)
		return starts
	}
	build := func() {
		_, model, err := s.controlDB.GetSite(ctx, s.site.SiteDB.Subdomain)
		require.NoError(t, err)
		siteContext := &httputils.SiteContext{SiteID: s.site.SiteDB.Subdomain, ControlDB: s.controlDB, SiteDB: s.site.SiteDB, SiteModel: model}
		require.NoError(t, sitedata.RebuildSummaryDays(ctx, q, siteContext, []time.Time{day}, now))
		require.NoError(t, q.ClearStaleSummaries(ctx, sitedb.ClearStaleSummariesParams{Before: new(time.Now().UnixMilli())}))
		require.Empty(t, stale())
	}

	resp := s.request(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, http.MethodPost, "/data/insulin", fmt.Sprintf(`{"dose": 4, "occurred_at": %d}`, day.UnixMilli()))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	build()
	resp = s.request(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone, http.MethodPost, "/data/insulin/"+created.Result.ID, fmt.Sprintf(`{"dose": 4, "occurred_at": %d}`, day.AddDate(0, 0, -1).UnixMilli()))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, stale(), 1, "the day it moved from")

	_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "deleted-for-good", Mgdl: 120, ReadingType: "sgv", OccurredAt: day.UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	build()
	status, _, _ := testutils.JSONRequest[any](t, s.router, http.MethodDelete, "http://"+s.site.SiteDB.Subdomain+".ns.localhost/api/v1/entries/deleted-for-good?permanent=true&token="+s.site.PwdNSToken, nil)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, stale(), 1, "the day it was deleted from")
}
