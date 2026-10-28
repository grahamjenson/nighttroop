package alarms

import (
	"slices"
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

// nows are each bus's alarms, one per server
var nows = struct {
	sync.Mutex
	of map[*events.SiteBus]*Now
}{of: map[*events.SiteBus]*Now{}}

// NowFor is the alarms of every alarm event published on the bus, kept from the first time it is asked for, so the
// server asks as it starts
func NowFor(bus *events.SiteBus) *Now {
	nows.Lock()
	defer nows.Unlock()
	if now := nows.of[bus]; now != nil {
		return now
	}
	now := &Now{checked: map[nowKey]Checked{}}
	nows.of[bus] = now
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

// Quiet stops an identity's alarm alerting, as a snooze does, until the checker next sends its alarms
func (n *Now) Quiet(siteID string, identityID int64, t Type) {
	n.mu.Lock()
	defer n.mu.Unlock()
	k := nowKey{siteID, identityID}
	if checked, ok := n.checked[k]; ok {
		checked.Alerting = slices.DeleteFunc(slices.Clone(checked.Alerting), func(alerting Type) bool { return alerting == t })
		n.checked[k] = checked
	}
}
