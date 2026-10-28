package sitecollections

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

type eventRow struct{ ID string }
type eventModel struct {
	ID string `json:"id"`
}

// eventsCollection is a collection of eventRows, named name for its permissions
type eventsCollection struct{ name string }

func (c eventsCollection) GetName() string { return c.name }
func (eventsCollection) GetRoute() string  { return "things" }
func (eventsCollection) ToModel(row *eventRow, _ *httputils.SiteContext) *eventModel {
	return &eventModel{ID: row.ID}
}
func (eventsCollection) GetRow(context.Context, *siteapiutils.CollectionTx, string, *httputils.SiteContext) (*eventRow, error) {
	return nil, nil
}

// watchAllCollection is an eventsCollection that also watches every other change, as the summary does
type watchAllCollection struct{ eventsCollection }

func (watchAllCollection) Watches(event events.Event, _ *httputils.SiteContext) bool {
	return event.Model != nil
}

// streams are the events streamed, and the control DB whose live DB holds the open streams
type streams struct {
	bus       *events.SiteBus
	controlDB *db.ControlDB
}

func newStreams(t *testing.T) streams {
	_, controlDB := testutils.NewControlDB(t)
	return streams{bus: events.NewSiteBus(), controlDB: controlDB}
}

// eventStream serves the handler for the site as role, and reads its stream
func eventStream(t *testing.T, streams streams, site string, role consts.IdentityRole, handler http.HandlerFunc) (*http.Response, *bufio.Reader) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := httputils.CreateSiteContext(r.Context(), &config.ServerConfig{Events: streams.bus}, streams.controlDB, nil, nil, site, "", httputils.HTTPActionRead, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
		handler(w, r.WithContext(httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{Role: role})))
	}))
	t.Cleanup(server.Close)
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp, bufio.NewReader(resp.Body)
}

// nextEvent is the stream's next event, its lines up to the blank line ending it
func nextEvent(t *testing.T, stream *bufio.Reader) string {
	lines := []string{}
	for {
		line, err := stream.ReadString('\n')
		require.NoError(t, err)
		if line == "\n" {
			return strings.Join(lines, "")
		}
		lines = append(lines, line)
	}
}

func TestEventsHandlerStreamsTheCollectionsChanges(t *testing.T) {
	tests := []struct {
		name       string
		collection CollectionBase[eventRow, eventModel]
		// want is the events streamed for a change to an own row, a change to another model, then another own row
		want []string
	}{
		{
			"its own rows, as their models",
			eventsCollection{name: "glucose_readings"},
			[]string{"event: created\ndata: {\"id\":\"1\"}\n", "event: deleted\ndata: {\"id\":\"2\"}\n"},
		},
		{
			"what it watches, with {} for other models",
			watchAllCollection{eventsCollection{name: "summary"}},
			[]string{"event: created\ndata: {\"id\":\"1\"}\n", "event: updated\ndata: {}\n", "event: deleted\ndata: {\"id\":\"2\"}\n"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			streams := newStreams(t)
			bus := streams.bus
			resp, stream := eventStream(t, streams, "site-a", consts.IdentityRoleUserReadOnly, EventsHandler(test.collection))
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

			// Only the site's own model changes are streamed, not its other events
			bus.Publish("site-b", events.Event{SiteID: "site-b", Action: events.Created, Model: &eventRow{ID: "other site"}})
			bus.Publish("site-a", events.Event{SiteID: "site-a", Name: events.SiteDisabled, Attributes: map[string]any{"by": "admin"}})
			bus.Publish("site-a", events.Event{SiteID: "site-a", Action: events.Created, Model: &eventRow{ID: "1"}})
			bus.Publish("site-a", events.Event{SiteID: "site-a", Action: events.Updated, Model: &struct{}{}})
			bus.Publish("site-a", events.Event{SiteID: "site-a", Action: events.Deleted, Model: &eventRow{ID: "2"}})
			for _, want := range test.want {
				require.Equal(t, want, nextEvent(t, stream))
			}
		})
	}
}

// An identity that may not search a collection cannot stream it, and a site can hold only so many streams open
func TestEventsHandlerLimits(t *testing.T) {
	streams := newStreams(t)
	resp, _ := eventStream(t, streams, "site-a", consts.IdentityRoleUser, EventsHandler[eventRow, eventModel](eventsCollection{name: "site_logs"}))
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	handler := EventsHandler[eventRow, eventModel](eventsCollection{name: "glucose_readings"})
	for range maxEventStreamsPerSite {
		resp, _ := eventStream(t, streams, "site-a", consts.IdentityRoleUser, handler)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
	resp, _ = eventStream(t, streams, "site-a", consts.IdentityRoleUser, handler)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	// A stream that ends is closed, so another can open
	open, _ := eventStream(t, streams, "site-c", consts.IdentityRoleUser, handler)
	for range maxEventStreamsPerSite - 1 {
		eventStream(t, streams, "site-c", consts.IdentityRoleUser, handler)
	}
	open.Body.Close()
	require.Eventually(t, func() bool {
		resp, _ := eventStream(t, streams, "site-c", consts.IdentityRoleUser, handler)
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 50*time.Millisecond)

	// Another site has its own streams
	resp, _ = eventStream(t, streams, "site-b", consts.IdentityRoleUser, handler)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
