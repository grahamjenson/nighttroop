package sitedb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A soft delete moves updated_at forward, even in the same millisecond as the write before it, and lastModified
// counts the deleted rows, so a history since a client's last sync has the documents deleted after it
func TestSoftDeleteMovesUpdatedAtForward(t *testing.T) {
	db, err := sitedb.NewDB(context.Background(), "softdelt", testutils.GetTestServerConfig())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	q := db.Queries()

	dose, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "dose", Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	group, err := q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "group", GroupType: "Bolus", InsulinDoseID: &dose.ID, OccurredAt: 1, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	reading, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "reading", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "test"})
	require.NoError(t, err)

	require.NoError(t, sitedb.SoftDeleteEventGroupAndEvents(ctx, q, group.ID, nil))
	_, err = q.SoftDeleteGlucoseReadingsByType(ctx, sitedb.SoftDeleteGlucoseReadingsByTypeParams{ReadingType: "sgv", Limit: 1})
	require.NoError(t, err)

	deletedGroup, err := q.GetEventGroup(ctx, sitedb.GetEventGroupParams{ID: group.ID, IncludeDeleted: true})
	require.NoError(t, err)
	deletedDose, err := q.GetInsulinDose(ctx, sitedb.GetInsulinDoseParams{ID: dose.ID, IncludeDeleted: true})
	require.NoError(t, err)
	deletedReading, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: reading.ID, IncludeDeleted: true})
	require.NoError(t, err)

	for _, c := range []struct {
		name          string
		before, after int64
		deletedAt     *int64
	}{
		{"event group", group.UpdatedAt, deletedGroup.EventGroup.UpdatedAt, deletedGroup.EventGroup.DeletedAt},
		{"insulin dose", dose.UpdatedAt, deletedDose.UpdatedAt, deletedDose.DeletedAt},
		{"glucose reading", reading.UpdatedAt, deletedReading.UpdatedAt, deletedReading.DeletedAt},
	} {
		assert.NotNil(t, c.deletedAt, c.name)
		assert.Greater(t, c.after, c.before, c.name)
	}

	lastModified, err := q.LastModifiedCollections(ctx)
	require.NoError(t, err)
	assert.Equal(t, deletedGroup.EventGroup.UpdatedAt, lastModified.EventGroupUpdatedAt)
	assert.Equal(t, deletedReading.UpdatedAt, lastModified.GlucoseReadingUpdatedAt)
}
