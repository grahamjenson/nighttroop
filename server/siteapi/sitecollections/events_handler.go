package sitecollections

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// CollectionEvents is a collection whose event stream has more than changes to its own rows, e.g. the summary,
// which is made from many collections' rows, or the alarms, which are an identity's alarm events
type CollectionEvents interface {
	// Watches is whether one of the site's events changes this collection for the identity streaming it
	Watches(event events.Event, siteContext *httputils.SiteContext) bool
}

const (
	// maxEventStreamsPerSite stops one site holding open as many connections as it likes
	maxEventStreamsPerSite = 20
	// eventStreamBuffer is how many changes can wait to be sent; a client too slow to read them is dropped, and
	// reconnects to read the collection again
	eventStreamBuffer = 100
	// eventStreamPing keeps a quiet stream open through proxies, and finds a client that has gone
	eventStreamPing = 25 * time.Second
	// maxEventStreamTime ends a stream, so its client reconnects and is authenticated again
	maxEventStreamTime = time.Hour
)

// EventsHandler streams a collection's changes as server-sent events, each named created, updated or deleted.
// A change to one of the collection's own rows has its model as JSON, as the API answers it; any other change the
// collection watches has {}, and is updated when it is not a model's. A page or API client can then update as changes
// happen, rather than polling.
func EventsHandler[R, M any](collection CollectionBase[R, M]) http.HandlerFunc {
	watches := func(event events.Event, _ *httputils.SiteContext) bool {
		_, own := event.Model.(*R)
		return own
	}
	if c, ok := collection.(CollectionEvents); ok {
		watches = c.Watches
	}

	return func(w http.ResponseWriter, r *http.Request) {
		siteContext := SiteContextFor(w, r, collection.GetName(), siteapiutils.ActionSearch)
		if siteContext == nil {
			return
		}
		// Each open stream is a connection in the live DB, so a site can hold only so many open
		live := siteContext.ControlDB.Live.Queries
		connection, err := live.OpenConnection(r.Context(), livedb.OpenConnectionParams{SiteID: siteContext.SiteID,
			IdentityID: siteContext.IdentityModel.ID, Kind: "event_stream", Collection: collection.GetName(), Max: maxEventStreamsPerSite})
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(60))
			return
		}
		if err != nil {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrInternalServerError)
			return
		}
		// The request's context is done as the stream ends, so it is not the one closing it
		defer live.CloseConnection(context.Background(), livedb.CloseConnectionParams{ID: connection})

		// The bus calls this in the publisher's goroutine, so it must not wait for a slow client
		changes, overflowed, overflow := make(chan events.Event, eventStreamBuffer), make(chan struct{}), sync.Once{}
		// Only the events the collection watches are streamed: the site's others, such as a control identity disabling
		// it, are not its users'
		unsubscribe := siteContext.Config.Events.Subscribe(siteContext.SiteID, func(event events.Event) {
			if !watches(event, siteContext) {
				return
			}
			select {
			case changes <- event:
			default:
				overflow.Do(func() { close(overflowed) })
			}
		})
		defer unsubscribe()

		// The server's write timeout would end the stream
		stream := http.NewResponseController(w)
		_ = stream.SetWriteDeadline(time.Time{})
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		if stream.Flush() != nil {
			return
		}

		ping := time.NewTicker(eventStreamPing)
		defer ping.Stop()
		end := time.After(maxEventStreamTime)
		for {
			var err error
			select {
			case <-r.Context().Done():
				return
			case <-overflowed:
				return
			case <-end:
				return
			case <-ping.C:
				_, err = io.WriteString(w, ": ping\n\n")
			case event := <-changes:
				data := []byte("{}")
				if row, own := event.Model.(*R); own {
					if data, err = json.Marshal(collection.ToModel(row, siteContext)); err != nil {
						return
					}
				}
				_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", cmp.Or(event.Action, events.Updated), data)
			}
			if err != nil || stream.Flush() != nil {
				return
			}
		}
	}
}
