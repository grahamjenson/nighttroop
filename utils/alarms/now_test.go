package alarms

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
)

func TestNowKeepsEachIdentitysLastAlarms(t *testing.T) {
	bus := events.NewSiteBus()
	now := WatchNow(bus)
	alarm := func(siteID string, checked Checked) {
		bus.Publish(siteID, events.Event{Name: events.Alarm, SiteID: siteID, Attributes: checked})
	}

	alarm("site-a", Checked{IdentityID: 1, Notifications: []Notification{{Type: UrgentLow, Kind: Started}}, Active: []Type{UrgentLow, Low}, Alerting: []Type{UrgentLow}})
	alarm("site-b", Checked{IdentityID: 1, Active: []Type{NoData}, Alerting: []Type{NoData}})
	require.Equal(t, Checked{IdentityID: 1, Active: []Type{UrgentLow, Low}, Alerting: []Type{UrgentLow}}, now.Of("site-a", 1), "without its notifications")
	require.Equal(t, []Type{NoData}, now.Of("site-b", 1).Alerting, "each site's identities are their own")
	require.Empty(t, now.Of("site-a", 2).Alerting)

	// Other events are not alarms, and an identity with nothing active is forgotten
	bus.Publish("site-a", events.Event{Name: events.SiteDisabled, SiteID: "site-a"})
	require.Equal(t, []Type{UrgentLow}, now.Of("site-a", 1).Alerting)
	alarm("site-a", Checked{IdentityID: 1, Notifications: []Notification{{Type: UrgentLow, Kind: Cleared}}})
	require.Equal(t, Checked{}, now.Of("site-a", 1))
	require.Len(t, now.checked, 1)
}
