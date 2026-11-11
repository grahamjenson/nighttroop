package events

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type GlucoseReading struct{}

type EventGroupsBundle struct{}

func TestModelChangeIsNamedForTheModelAndAction(t *testing.T) {
	tests := []struct {
		model  any
		action Action
		name   Name
	}{
		{&GlucoseReading{}, Created, "glucose_reading.created"},
		{GlucoseReading{}, Updated, "glucose_reading.updated"},
		{&EventGroupsBundle{}, Deleted, "event_groups_bundle.deleted"},
		{nil, Created, "model.created"},
	}
	for _, test := range tests {
		event := ModelChange("site-a", test.action, test.model)
		require.Equal(t, test.name, event.Name)
		require.Equal(t, "site-a", event.SiteID)
		require.Equal(t, test.action, event.Action)
		require.False(t, event.IsPersisted(), "a model change is kept in the site's DB, not as an event")
	}
}

func TestOnlyPersistedEventsAreKept(t *testing.T) {
	for _, name := range Persisted {
		require.True(t, Event{Name: name}.IsPersisted(), name)
	}
	require.Contains(t, PersistedNames(), "connector_backfilled")
	require.False(t, Event{Name: "something_else"}.IsPersisted())
}
