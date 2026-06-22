package user

import (
	"time"
)

func (model *GlucoseReading) CreatedAtDate() time.Time {
	t := time.UnixMilli(model.CreatedAt)
	return t
}

func (u *Identity) LastSeenAtDate() *time.Time {
	if u.LastSeenAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastSeenAt)
	return &t
}

func (u *Identity) LastWriteAtDate() *time.Time {
	if u.LastWriteAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastWriteAt)
	return &t
}

func (g *GlucoseReading) ToDBCreateParams() CreateGlucoseReadingParams {
	return CreateGlucoseReadingParams{
		Eid:                   g.Eid,
		OccurredAt:            g.OccurredAt,
		Value:                 g.Value,
		RawJsonData:           g.RawJsonData,
		UtcOffsetMinutes:      g.UtcOffsetMinutes,
		CreatedByIdentityID:   g.CreatedByIdentityID,
		CreatedByIdentityName: g.CreatedByIdentityName,
	}
}
