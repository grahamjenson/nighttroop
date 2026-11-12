package nsutils

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// The database refusing a value is the client's error, saying the column and why, never in the driver's words
func TestGetSQLTranslatedError(t *testing.T) {
	db, err := sitedb.NewDB(context.Background(), "sqlerrs", testutils.GetTestServerConfig())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	q := db.Queries()
	dose, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "dose", Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	_, err = q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "group", GroupType: "Bolus", InsulinDoseID: &dose.ID, OccurredAt: 1, CreatedByIdentityName: "test"})
	require.NoError(t, err)

	tests := []struct {
		name    string
		write   func() error
		message string
	}{
		{"check", func() error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "big", Dose: 8882.5, OccurredAt: 1, CreatedByIdentityName: "test"})
			return err
		}, "Bad or missing body: dose is out of range"},
		{"check in a function", func() error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: strings.Repeat("a", 40), Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
			return err
		}, "Bad or missing body: id is too long"},
		{"unique", func() error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: "dose", Dose: 1, OccurredAt: 1, CreatedByIdentityName: "test"})
			return err
		}, "Bad or missing body: id is already used"},
		{"unique other than the id", func() error {
			_, err := q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "other", GroupType: "Bolus", InsulinDoseID: &dose.ID, OccurredAt: 1, CreatedByIdentityName: "test"})
			return err
		}, "Bad or missing body: insulin_dose_id is already used"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.write()
			require.Error(t, err)
			httpErr := GetSQLTranslatedError(err)
			require.NotNil(t, httpErr, err.Error())
			assert.Equal(t, 400, httpErr.Status)
			assert.Equal(t, tt.message, httpErr.Message)
			assert.NotContains(t, httpErr.Message, "constraint failed")
		})
	}

	assert.Nil(t, GetSQLTranslatedError(errors.New("CHECK constraint failed: dose BETWEEN -500 AND 500")), "not from the database")
}

// A constraint's first column is found in each way the driver words a constraint failing
func TestConstraintColumn(t *testing.T) {
	tests := []struct{ message, column string }{
		{"constraint failed: CHECK constraint failed: dose BETWEEN -500 AND 500 (275)", "dose"},
		{"CHECK constraint failed: carbs BETWEEN -1000 AND 1000", "carbs"},
		{"constraint failed: CHECK constraint failed: length(notes) <= 1024 (275)", "notes"},
		{"constraint failed: CHECK constraint failed: mgdl > 0 AND mgdl <= 1000 (275)", "mgdl"},
		{"constraint failed: NOT NULL constraint failed: insulin_doses.dose (1299)", "dose"},
		{"constraint failed: UNIQUE constraint failed: event_groups.insulin_dose_id, event_groups.id (2067)", "insulin_dose_id"},
		{"constraint failed: FOREIGN KEY constraint failed (787)", "a value"},
		{"", "a value"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.column, constraintColumn(tt.message), tt.message)
	}
}
