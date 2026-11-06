package sitecollections

import (
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
// which is made from many collections' rows
type CollectionEvents interface {
	// Watches is whether one of the site's events changes this collection for the identity streaming it
	Watches(event events.Event, siteContext *httputils.SiteContext) bool
}

// CollectionEventsHidden is a collection whose stream is told of some changes to its own rows it may not read, as a
// non-admin's settings are of the site's, which they inherit; those are sent as {}
type CollectionEventsHidden interface {
	// Hides is whether a change to one of the collection's own rows is sent as {} to the identity streaming it
	Hides(event events.Event, siteContext *httputils.SiteContext) bool
}

const (
	// maxEventStreamsPerSite stops one site holding open as many connections as it likes
	maxEventStreamsPerSite = 20
	// eventStreamBuffer is how many changes can wait to be sent; a client too slow to read them is dropped, and
	// reconnects to read the collection again
	eventStreamBuffer = 100
	// maxEventStreamTime ends a stream, so its client reconnects and is authenticated again
	maxEventStreamTime = time.Hour
)

// eventStreamPing keeps a quiet stream open through proxies, finds a client that has gone, and checks the stream's
// identity has not been revoked
var eventStreamPing = 25 * time.Second

// EventsHandler streams a collection's changes as server-sent events, each named created, updated or deleted.
// A change to one of the collection's own rows has its model as JSON, as the API answers it; any other change the
// collection watches has {}. A page or API client can then update as changes happen, rather than polling.
func EventsHandler[R, M any](collection CollectionBase[R, M]) http.HandlerFunc {
	watches := func(event events.Event, _ *httputils.SiteContext) bool {
		_, own := event.Model.(*R)
		return own
	}
	if c, ok := collection.(CollectionEvents); ok {
		watches = c.Watches
	}
	hides := func(events.Event, *httputils.SiteContext) bool { return false }
	if c, ok := collection.(CollectionEventsHidden); ok {
		hides = c.Hides
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
			// A row an identity owns, as its alarms, is streamed to it alone
			owned, isOwned := event.Model.(interface{ GetIdentityID() int64 })
			if !watches(event, siteContext) || (isOwned && owned.GetIdentityID() != siteContext.IdentityModel.ID) {
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
				// A stream is only authenticated as it opens, so it ends once its identity is revoked
				if httputils.IdentityRevoked(r.Context(), siteContext.ControlDB, siteContext.IdentityModel) {
					return
				}
				_, err = io.WriteString(w, ": ping\n\n")
			case event := <-changes:
				data := []byte("{}")
				if row, own := event.Model.(*R); own && !hides(event, siteContext) {
					if data, err = json.Marshal(collection.ToModel(row, siteContext)); err != nil {
						return
					}
				}
				_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Action, data)
			}
			if err != nil || stream.Flush() != nil {
				return
			}
		}
	}
}
