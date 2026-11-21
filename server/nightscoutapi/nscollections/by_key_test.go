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
		_, err = q.CreateInsulinDose(t.Context(), sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: at, CreatedByIdentityName: "test"})
		require.NoError(t, err)
		_, err = q.CreateEventGroup(t.Context(), sitedb.CreateEventGroupParams{ID: id, GroupType: "Correction Bolus", InsulinDoseID: &id, OccurredAt: at, CreatedByIdentityName: "test"})
		require.NoError(t, err)
		_, err = q.CreateFoodPreset(t.Context(), sitedb.CreateFoodPresetParams{ID: id, Name: "food", CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
	return &httputils.SiteContext{SiteID: model.ID, ControlDB: controlDB, SiteDB: siteDB, SiteModel: model,
		IdentityModel: &controldb.SiteIdentity{ID: 1, Name: "pwd_phone"}}
}

// keys are the identifiers of the models, in their order, or the error reading them
func keys[DM interface{ GetID() string }](models []*DM, err error) []string {
	if err != nil {
		return []string{err.Error()}
	}
	ids := []string{}
	for _, model := range models {
		ids = append(ids, (*model).GetID())
	}
	return ids
}

// byKey reads the models in the order of their identifiers, leaving out one there is not, and a deleted one unless
// deleted ones are asked for, whether its collection reads them in one query or one by one
func TestByKey(t *testing.T) {
	siteContext := siteWithPages(t, 4)
	q := siteContext.SiteDB.Queries()
	deleted := fmt.Sprintf("%024d", 1)
	_, err := q.SoftDeleteGlucoseReading(t.Context(), sitedb.SoftDeleteGlucoseReadingParams{ID: deleted})
	require.NoError(t, err)
	_, err = q.SoftDeleteDeviceReport(t.Context(), sitedb.SoftDeleteDeviceReportParams{ID: deleted})
	require.NoError(t, err)
	_, err = q.SoftDeleteEventGroup(t.Context(), sitedb.SoftDeleteEventGroupParams{ID: deleted})
	require.NoError(t, err)
	ids := []string{fmt.Sprintf("%024d", 2), "missing", fmt.Sprintf("%024d", 0), deleted, fmt.Sprintf("%024d", 3)}
	entries := NewEntriesCollection().(Collection[nightscoutmodels.Entry, sitedb.GlucoseReading])
	statuses := NewDeviceStatusCollection().(Collection[nightscoutmodels.DeviceStatus, sitedb.DeviceReport])
	treatments := NewTreatmentCollection().(Collection[nightscoutmodels.Treatment, sitedb.EventGroupsBundle])
	foods := NewFoodCollection().(Collection[nightscoutmodels.Food, sitedb.FoodPreset])
	profiles := NewProfileCollection().(Collection[nightscoutmodels.ProfileStore, sitedb.ProfilePresetBundle])
	all := []string{ids[0], ids[2], ids[3], ids[4]}
	for _, includeDeleted := range []bool{false, true} {
		want := []string{ids[0], ids[2], ids[4]}
		if includeDeleted {
			want = all
		}
		require.Equal(t, want, keys(entries.byKey(q, t.Context(), ids, includeDeleted)))
		require.Equal(t, want, keys(statuses.byKey(q, t.Context(), ids, includeDeleted)))
		bundles, err := treatments.byKey(q, t.Context(), ids, includeDeleted)
		require.Equal(t, want, keys(bundles, err))
		for _, bundle := range bundles {
			require.NotNil(t, bundle.InsulinDose, "a treatment's events are read with it")
			require.Equal(t, bundle.EventGroup.ID, bundle.InsulinDose.ID)
		}
		// Foods are not soft deleted
		require.Equal(t, all, keys(foods.byKey(q, t.Context(), ids, includeDeleted)))
	}
	// A collection that does not read many in one query reads each by itself
	stores, err := profiles.byKey(q, t.Context(), []string{"missing"}, true)
	require.NoError(t, err)
	require.Empty(t, stores)
}

// A page of 1000 entries, device statuses and treatments AAPS reads, sorted by date and from API v3 history, each read by its
// identifier. Run with go test -run XXX -bench PagesByKey ./server/nightscoutapi/nscollections
func BenchmarkPagesByKey(b *testing.B) {
	siteContext := siteWithPages(b, 2000)
	since, ascending := int64(1_700_000_000_000), nsquery.SortOccurredAtAsc
	for name, collection := range map[string]CollectionStore{"entries": NewEntriesCollection(), "devicestatus": NewDeviceStatusCollection(),
		"treatments": NewTreatmentCollection()} {
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

// Treatments and foods read a page in one query
var (
	_ CollectionKeysReader[sitedb.EventGroupsBundle] = TreatmentCollection{}
	_ CollectionKeysReader[sitedb.FoodPreset]        = FoodCollection{}
)
