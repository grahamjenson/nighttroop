package page

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// pagerDuty is a stand in for PagerDuty's Events API, keeping each event it is sent and answering with status
func pagerDuty(t *testing.T, status int) (*httptest.Server, *[]event) {
	events := &[]event{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e event
		require.NoError(t, json.NewDecoder(r.Body).Decode(&e))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		*events = append(*events, e)
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	return server, events
}

// Each level is sent to PagerDuty as its severity, which its service's urgency rule makes low urgency (an email) or
// high (an SMS or a call), with the message as its summary and its dedup key, so the same problem is one incident,
// this server as its source, and what it is about as its details
func TestAdminSendsEachLevelToPagerDuty(t *testing.T) {
	tests := []struct {
		level    Level
		severity string
	}{
		{Email, "warning"},
		{SMS, "error"},
		{Call, "critical"},
	}
	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			server, events := pagerDuty(t, http.StatusAccepted)
			New("routing-key", server.URL, "nt.example.com").Admin(t.Context(), tt.level, "Sites are unavailable",
				"site_ids", []string{"abcdefgh"}, "error", errors.New("no backup"))

			require.Len(t, *events, 1)
			e := (*events)[0]
			require.Equal(t, "routing-key", e.RoutingKey)
			require.Equal(t, "trigger", e.EventAction)
			require.Equal(t, "Sites are unavailable", e.DedupKey)
			require.Equal(t, payload{Summary: "Sites are unavailable", Source: "nt.example.com", Severity: Level(tt.severity),
				CustomDetails: map[string]any{"site_ids": "[abcdefgh]", "error": "no backup"}}, e.Payload)
		})
	}
}

// Without a routing key PagerDuty is not set up, so a page is only logged
func TestAdminWithoutARoutingKeyOnlyLogs(t *testing.T) {
	server, events := pagerDuty(t, http.StatusAccepted)
	New("", server.URL, "nt.example.com").Admin(t.Context(), Call, "Sites are unavailable")
	require.Empty(t, *events)
}

// A page PagerDuty does not take is logged, not retried, and does not stop the caller
func TestSendFailsWhenPagerDutyRefuses(t *testing.T) {
	server, events := pagerDuty(t, http.StatusBadRequest)
	pager := New("bad-key", server.URL, "nt.example.com")
	require.ErrorContains(t, pager.send(t.Context(), event{RoutingKey: "bad-key"}), "400 Bad Request")
	pager.Admin(t.Context(), SMS, "Sites are unavailable")
	require.Len(t, *events, 2)
}
