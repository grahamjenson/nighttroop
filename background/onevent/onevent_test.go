package onevent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/liteq"
	"nighttroop.t1d.nz/background/onevent"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestListenQueuesAJobForEachMatchingSite(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, bus, queue := context.TODO(), events.NewSiteBus(), liteq.New(controlDB)
	ran := []onevent.Fired{}
	job := onevent.Job{
		Queue:  "test-on-event",
		Filter: func(event events.Event) bool { return event.Name == events.SiteTimezoneChanged },
		Run: func(ctx context.Context, fired onevent.Fired) error {
			ran = append(ran, fired)
			return nil
		},
	}
	consumers := onevent.Listen(bus, queue, job)
	require.Len(t, consumers, 1)

	// An event that does not match queues nothing
	bus.Publish("site-a", events.Event{Name: events.SiteCreated, SiteID: "site-a"})
	require.NoError(t, consumers[0].RunQueued(ctx))
	require.Empty(t, ran)

	// Matching events run the job once a site, however many there were before it ran
	bus.Publish("site-a", events.Event{Name: events.SiteTimezoneChanged, SiteID: "site-a"})
	bus.Publish("site-a", events.Event{Name: events.SiteTimezoneChanged, SiteID: "site-a"})
	bus.Publish("site-b", events.Event{Name: events.SiteTimezoneChanged, SiteID: "site-b", Batch: "b"})
	require.NoError(t, consumers[0].RunQueued(ctx))
	require.ElementsMatch(t, []onevent.Fired{
		{Name: events.SiteTimezoneChanged, SiteID: "site-a"},
		{Name: events.SiteTimezoneChanged, SiteID: "site-b", Batch: "b"},
	}, ran)

	// Once it has run, the next matching event runs it again
	bus.Publish("site-a", events.Event{Name: events.SiteTimezoneChanged, SiteID: "site-a"})
	require.NoError(t, consumers[0].RunQueued(ctx))
	require.Len(t, ran, 3)
}
