package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// BenchmarkReadsBySite reads what the control, jobs, live and logs DBs' prepared queries read most often by a site or a
// queue, ordered, as each Users & Devices page, job poll, alarm check and logs page does. Each run binds another site or
// queue, which must not make SQLite prepare the query again (data/sql/README.md).
// Run with go test -run XXX -bench ReadsBySite ./data/db
func BenchmarkReadsBySite(b *testing.B) {
	_, control := testutils.NewControlDB(b)
	ctx := b.Context()
	sites := []string{"benchsitea", "benchsiteb"}
	queues := []string{"benchqueuea", "benchqueueb"}

	b.Run("site's identities", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			_, err := control.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: sites[i%2]})
			require.NoError(b, err)
		}
	})
	b.Run("site's push registrations", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			_, err := control.Queries.SearchPushRegistrations(ctx, controldb.SearchPushRegistrationsParams{SiteID: sites[i%2], Limit: 10})
			require.NoError(b, err)
		}
	})
	b.Run("queue's jobs", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			_, err := control.Jobs.Queries.MarkJobsForConsumer(ctx, jobsdb.MarkJobsForConsumerParams{Queue: queues[i%2], ExecuteAfter: int64(i), Limit: 10})
			require.NoError(b, err)
		}
	})
	b.Run("identity's alarms", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			_, err := control.Live.Queries.IdentityAlarms(ctx, livedb.IdentityAlarmsParams{SiteID: sites[i%2], IdentityID: int64(i % 2)})
			require.NoError(b, err)
		}
	})
	b.Run("site's logs", func(b *testing.B) {
		i := 0
		for b.Loop() {
			i++
			_, err := control.Logs.Queries.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: sites[i%2], Limit: 10})
			require.NoError(b, err)
		}
	})
	b.Run("connector's events", func(b *testing.B) {
		connector := "dexcom"
		i := 0
		for b.Loop() {
			i++
			_, err := control.Logs.Queries.RecentConnectorEvents(ctx, logsdb.RecentConnectorEventsParams{SiteID: sites[i%2], ConnectorType: &connector, Limit: 10})
			require.NoError(b, err)
		}
	})
}
