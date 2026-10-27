package sitecollections

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// CollectionEvents is a collection whose event stream has changes to more than its own rows, e.g. the summary,
// which is made from many collections' rows
type CollectionEvents interface {
	// Watches is whether a change to one of the site's models changes this collection
	Watches(event events.Event) bool
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

// eventStreams counts each site's open streams, at most one count per site
var eventStreams = struct {
	sync.Mutex
	open map[string]int
}{open: map[string]int{}}

// EventsHandler streams a collection's changes as server-sent events, each named created, updated or deleted.
// A change to one of the collection's own rows has its model as JSON, as the API answers it; any other change the
// collection watches has {}. A page or API client can then update as changes happen, rather than polling.
func EventsHandler[R, M any](collection CollectionBase[R, M]) http.HandlerFunc {
	watches := func(event events.Event) bool {
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
		if !openEventStream(siteContext.Subdomain) {
			writeError(w, r, siteContext.AcceptType, apiutils.ErrRateLimitExceeded(60))
			return
		}
		defer closeEventStream(siteContext.Subdomain)

		// The bus calls this in the publisher's goroutine, so it must not wait for a slow client
		changes, overflowed, overflow := make(chan events.Event, eventStreamBuffer), make(chan struct{}), sync.Once{}
		// Only model changes are streamed: the site's other events, such as a control identity disabling it, are not its users'
		unsubscribe := siteContext.Config.Events.Subscribe(siteContext.Subdomain, func(event events.Event) {
			if event.Model == nil || !watches(event) {
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
				_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Action, data)
			}
			if err != nil || stream.Flush() != nil {
				return
			}
		}
	}
}

// openEventStream counts a stream the site opens, or is false when it has as many open as it may
func openEventStream(siteID string) bool {
	eventStreams.Lock()
	defer eventStreams.Unlock()
	if eventStreams.open[siteID] >= maxEventStreamsPerSite {
		return false
	}
	eventStreams.open[siteID]++
	return true
}

func closeEventStream(siteID string) {
	eventStreams.Lock()
	defer eventStreams.Unlock()
	eventStreams.open[siteID]--
}
