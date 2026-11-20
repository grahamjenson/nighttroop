package nscollections

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// siteWithPages is a site with count glucose readings and device reports, five minutes apart, identified by their index
func siteWithPages(t testing.TB, count int64) *httputils.SiteContext {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	siteDB, model, err := controlDB.GetSite(t.Context(), site.SiteDB.Subdomain)
	require.NoError(t, err)
	tx, err := siteDB.SQLiteDB.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	q := siteDB.QueriesWithTx(tx)
	for i := range count {
		id, at := fmt.Sprintf("%024d", i), 1_700_000_000_000+i*300_000
		_, err := q.CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 100, ReadingType: "sgv", OccurredAt: at, CreatedByIdentityName: "test"})
		require.NoError(t, err)
		_, err = q.CreateDeviceReport(t.Context(), sitedb.CreateDeviceReportParams{ID: id, OccurredAt: at, CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
	return &httputils.SiteContext{SiteID: model.ID, ControlDB: controlDB, SiteDB: siteDB, SiteModel: model,
		IdentityModel: &controldb.SiteIdentity{ID: 1, Name: "pwd_phone"}}
}

// byKey reads the models in the order of their identifiers, leaving out one there is not, and a deleted one unless
// deleted ones are asked for, whether its collection reads them in one query or one by one
func TestByKey(t *testing.T) {
	siteContext := siteWithPages(t, 4)
	q := siteContext.SiteDB.Queries()
	_, err := q.SoftDeleteGlucoseReading(t.Context(), sitedb.SoftDeleteGlucoseReadingParams{ID: fmt.Sprintf("%024d", 1)})
	require.NoError(t, err)
	_, err = q.SoftDeleteDeviceReport(t.Context(), sitedb.SoftDeleteDeviceReportParams{ID: fmt.Sprintf("%024d", 1)})
	require.NoError(t, err)
	ids := []string{fmt.Sprintf("%024d", 2), "missing", fmt.Sprintf("%024d", 0), fmt.Sprintf("%024d", 1), fmt.Sprintf("%024d", 3)}
	entries := NewEntriesCollection().(Collection[nightscoutmodels.Entry, sitedb.GlucoseReading])
	statuses := NewDeviceStatusCollection().(Collection[nightscoutmodels.DeviceStatus, sitedb.DeviceReport])
	profiles := NewProfileCollection().(Collection[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle])
	for _, includeDeleted := range []bool{false, true} {
		want := []string{ids[0], ids[2], ids[4]}
		if includeDeleted {
			want = []string{ids[0], ids[2], ids[3], ids[4]}
		}
		readings, err := entries.byKey(q, t.Context(), ids, includeDeleted)
		require.NoError(t, err)
		reports, err := statuses.byKey(q, t.Context(), ids, includeDeleted)
		require.NoError(t, err)
		var readingIDs, reportIDs []string
		for i := range readings {
			readingIDs = append(readingIDs, readings[i].ID)
		}
		for i := range reports {
			reportIDs = append(reportIDs, reports[i].ID)
		}
		require.Equal(t, want, readingIDs)
		require.Equal(t, want, reportIDs)
	}
	// A collection that does not read many in one query reads each by itself
	stores, err := profiles.byKey(q, t.Context(), []string{"missing"}, true)
	require.NoError(t, err)
	require.Empty(t, stores)
}

// A page of 1000 entries and device statuses AAPS reads, sorted by date and from API v3 history, each read by its
// identifier. Run with go test -run XXX -bench PagesByKey ./server/nightscoutapi/nscollections
func BenchmarkPagesByKey(b *testing.B) {
	siteContext := siteWithPages(b, 2000)
	since, ascending := int64(1_700_000_000_000), nsquery.SortOccurredAtAsc
	for name, collection := range map[string]CollectionStore{"entries": NewEntriesCollection(), "devicestatus": NewDeviceStatusCollection()} {
		b.Run(name+"/ascending", func(b *testing.B) {
			for b.Loop() {
				docs, err := collection.Search(b.Context(), siteContext, nsquery.Query{Limit: 1000, DateGT: &since, HasFilters: true, SortOrder: &ascending})
				require.Nil(b, err)
				require.Len(b, docs, 1000)
			}
		})
		b.Run(name+"/history", func(b *testing.B) {
			for b.Loop() {
				docs, err := collection.History(b.Context(), siteContext, 0, nsquery.Query{Limit: 1000})
				require.Nil(b, err)
				require.Len(b, docs, 1000)
			}
		})
	}
}
