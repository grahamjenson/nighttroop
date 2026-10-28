package alarms

import (
	"sync"

	"nighttroop.t1d.nz/data/events"
)

// Now is each identity's alarms as the checker last sent them, kept from the server's alarm events, so a page can
// show them as it opens. The checker and the pages are one server, so they start, and forget, together.
type Now struct {
	mu      sync.Mutex
	checked map[nowKey]Checked
}

type nowKey struct {
	siteID     string
	identityID int64
}

// WatchNow keeps the alarms of every alarm event published on the bus
func WatchNow(bus *events.SiteBus) *Now {
	now := &Now{checked: map[nowKey]Checked{}}
	bus.SubscribeAll(func(event events.Event) {
		checked, ok := event.Attributes.(Checked)
		if !ok || event.Name != events.Alarm {
			return
		}
		now.mu.Lock()
		defer now.mu.Unlock()
		k := nowKey{event.SiteID, checked.IdentityID}
		if len(checked.Active) == 0 {
			delete(now.checked, k)
			return
		}
		now.checked[k] = Checked{IdentityID: checked.IdentityID, Active: checked.Active, Alerting: checked.Alerting}
	})
	return now
}

// Of is an identity's active and alerting alarms
func (n *Now) Of(siteID string, identityID int64) Checked {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.checked[nowKey{siteID, identityID}]
}
