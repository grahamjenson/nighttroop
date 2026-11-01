package dexcommodels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/dexcommodels"
	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

// Each sync fetches the last 15 minutes, so the same reading must not be stored twice
func TestEGVToCreateDedupesRepeatedReadings(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	queries := site.SiteDB.Queries()

	reading := dexcommodels.EGV{Value: 120, SystemTime: time.UnixMilli(1786184674000).UTC(), DisplayTime: time.UnixMilli(1786184674000), RawJson: `{"Value":120}`}
	next := dexcommodels.EGV{Value: 125, SystemTime: reading.SystemTime.Add(5 * time.Minute), DisplayTime: reading.DisplayTime.Add(5 * time.Minute), RawJson: `{"Value":125}`}

	_, err := queries.CreateGlucoseReading(t.Context(), dexcommodels.EGVToCreate(reading, "dexcom_sync"))
	require.NoError(t, err)

	_, err = queries.CreateGlucoseReading(t.Context(), dexcommodels.EGVToCreate(reading, "dexcom_sync"))
	require.True(t, utilerrors.IsSQLDuplicateID(err), "second sync of the same reading should be a duplicate, got %v", err)

	_, err = queries.CreateGlucoseReading(t.Context(), dexcommodels.EGVToCreate(next, "dexcom_sync"))
	require.NoError(t, err)
}
