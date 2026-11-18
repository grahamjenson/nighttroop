package controlapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/controlapi/controlcollections"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// An admin's test page goes to PagerDuty at the level they pick, as a new incident naming them; without PagerDuty set
// up, at a level there is not, or as a read only identity, it is refused, and one PagerDuty refuses says why
func TestTestPage(t *testing.T) {
	var sent []map[string]any
	status := http.StatusAccepted
	pagerDuty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&event))
		sent = append(sent, event)
		w.WriteHeader(status)
	}))
	defer pagerDuty.Close()
	routingKey, eventsURL := testConfig.PagerDutyRoutingKey, testConfig.PagerDutyEventsURL
	t.Cleanup(func() { testConfig.PagerDutyRoutingKey, testConfig.PagerDutyEventsURL = routingKey, eventsURL })
	testConfig.PagerDutyEventsURL = pagerDuty.URL

	controlDB, err := db.NewControlDB(t.Context(), testConfig)
	require.NoError(t, err)
	defer controlDB.Close()
	testPage := func(role consts.ControlRole, level string) *httptest.ResponseRecorder {
		ctx := httputils.CreateControlContext(context.Background(), controlDB, "", httputils.HTTPActionWrite)
		ctx = httputils.SetAdminControlContext(ctx, &controldb.ControlIdentity{Name: "erin", Role: role}, config.ServerStatus{})
		r := httptest.NewRequest(http.MethodPost, "/admins/test-page", strings.NewReader(url.Values{"level": {level}}.Encode())).WithContext(ctx)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		controlcollections.AdminOnly(http.HandlerFunc(TestPageHandler)).ServeHTTP(w, r)
		return w
	}

	w := testPage(consts.ControlRoleAdmin, "call")
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "PagerDuty is not set up")

	testConfig.PagerDutyRoutingKey = "routing-key"
	require.Equal(t, http.StatusForbidden, testPage(consts.ControlRoleRead, "call").Code)
	require.Equal(t, http.StatusBadRequest, testPage(consts.ControlRoleAdmin, "fax").Code)
	require.Empty(t, sent)

	for level, severity := range map[string]string{"email": "warning", "sms": "error", "call": "critical"} {
		require.Equal(t, http.StatusOK, testPage(consts.ControlRoleAdmin, level).Code, level)
		payload := sent[len(sent)-1]["payload"].(map[string]any)
		require.Equal(t, severity, payload["severity"])
		require.Contains(t, payload["summary"], "Test page from the control site, sent by erin at ")
	}

	status = http.StatusBadRequest
	w = testPage(consts.ControlRoleAdmin, "sms")
	require.Equal(t, http.StatusBadGateway, w.Code)
	require.Contains(t, w.Body.String(), "400 Bad Request")
}
