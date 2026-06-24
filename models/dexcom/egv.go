package dexcom

import (
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
)

// EGV represents a Dexcom glucose reading.
type EGV struct {
	Value       int
	SystemTime  time.Time
	DisplayTime time.Time
	RawJson     string
}

func (e EGV) UTCOffsetInMinutes() int {
	_, offset := e.DisplayTime.Zone()
	return offset / 60
}

func EGVToCreate(e EGV, idPrefix string, createdByIdentityID int64, createdByIdentityName string) user.CreateGlucoseReadingParams {
	return user.CreateGlucoseReadingParams{
		ID:                    utils.NextDateID(idPrefix, e.SystemTime.UnixMilli()),
		OccurredAt:            e.SystemTime.UnixMilli(),
		Value:                 float64(e.Value),
		CreatedByIdentityID:   createdByIdentityID,
		RawJsonData:           &e.RawJson,
		CreatedByIdentityName: createdByIdentityName,
		UtcOffsetMinutes:      int64(e.UTCOffsetInMinutes()),
	}
}
