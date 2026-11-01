package libremodels_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/libremodels"
	utilerrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

func reading(t *testing.T, raw string) libremodels.Reading {
	var r libremodels.Reading
	require.NoError(t, json.Unmarshal([]byte(raw), &r))
	return r
}

func TestReadingToCreate(t *testing.T) {
	r := reading(t, `{"FactoryTimestamp":"12/30/2022 4:47:40 PM","Timestamp":"12/31/2022 6:32:40 AM","ValueInMgPerDl":115,"TrendArrow":3}`)
	params := libremodels.ReadingToCreate(r, "libre_sync")
	require.Equal(t, "libre-1672418860", params.ID)
	require.Equal(t, time.Date(2022, 12, 30, 16, 47, 40, 0, time.UTC).UnixMilli(), params.OccurredAt)
	require.Equal(t, 115.0, params.Mgdl)
	require.Equal(t, "sgv", params.ReadingType)
	require.Equal(t, "libre_sync", params.CreatedByIdentityName)
	require.Equal(t, int64(13*60+45), *params.UtcOffsetMinutes)
	require.Equal(t, "Flat", *params.Direction)
	require.Contains(t, *params.RawJsonData, `"TrendArrow":3`)
}

func TestReadingDirections(t *testing.T) {
	tests := []struct {
		trend string
		want  string
	}{
		{`,"TrendArrow":1`, "SingleDown"},
		{`,"TrendArrow":2`, "FortyFiveDown"},
		{`,"TrendArrow":3`, "Flat"},
		{`,"TrendArrow":4`, "FortyFiveUp"},
		{`,"TrendArrow":5`, "SingleUp"},
		{`,"TrendArrow":0`, ""},
		{`,"TrendArrow":6`, ""},
		{``, ""},
	}
	for _, tt := range tests {
		t.Run(tt.trend, func(t *testing.T) {
			r := reading(t, `{"FactoryTimestamp":"1/2/2026 3:04:05 PM","Timestamp":"1/2/2026 3:04:05 PM","ValueInMgPerDl":100`+tt.trend+`}`)
			direction := libremodels.ReadingToCreate(r, "libre_sync").Direction
			if tt.want == "" {
				require.Nil(t, direction)
				return
			}
			require.Equal(t, tt.want, *direction)
		})
	}
}

func TestReadingRefusesBadValues(t *testing.T) {
	tests := map[string]string{
		"over 1000 mg/dL":       `{"FactoryTimestamp":"1/2/2026 3:04:05 PM","Timestamp":"1/2/2026 3:04:05 PM","ValueInMgPerDl":1001}`,
		"zero":                  `{"FactoryTimestamp":"1/2/2026 3:04:05 PM","Timestamp":"1/2/2026 3:04:05 PM","ValueInMgPerDl":0}`,
		"bad factory timestamp": `{"FactoryTimestamp":"2026-01-02T15:04:05Z","Timestamp":"1/2/2026 3:04:05 PM","ValueInMgPerDl":100}`,
		"bad timestamp":         `{"FactoryTimestamp":"1/2/2026 3:04:05 PM","ValueInMgPerDl":100}`,
		"null":                  `null`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var r libremodels.Reading
			require.Error(t, json.Unmarshal([]byte(raw), &r))
		})
	}
}

// Each sync reads the graph again, so the same reading must not be stored twice
func TestReadingToCreateDedupesRepeatedReadings(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	queries := site.SiteDB.Queries()

	first := reading(t, `{"FactoryTimestamp":"12/30/2022 4:47:40 PM","Timestamp":"12/30/2022 5:47:40 PM","ValueInMgPerDl":115}`)
	next := reading(t, `{"FactoryTimestamp":"12/30/2022 5:02:40 PM","Timestamp":"12/30/2022 6:02:40 PM","ValueInMgPerDl":120}`)

	_, err := queries.CreateGlucoseReading(t.Context(), libremodels.ReadingToCreate(first, "libre_sync"))
	require.NoError(t, err)
	_, err = queries.CreateGlucoseReading(t.Context(), libremodels.ReadingToCreate(first, "libre_sync"))
	require.True(t, utilerrors.IsSQLDuplicateID(err), "second sync of the same reading should be a duplicate, got %v", err)
	_, err = queries.CreateGlucoseReading(t.Context(), libremodels.ReadingToCreate(next, "libre_sync"))
	require.NoError(t, err)
}
