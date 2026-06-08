package nightscout_api

import (
	"fmt"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
)

func SGVToModel(sgv *user.Sgv) model.Record {
	if sgv == nil {
		return model.Record{}
	}
	data := map[string]any{
		"device": "AAPS-DexcomG6",

		"isValid":    true,
		"sgv":        56,
		"direction":  "FortyFiveDown",
		"type":       "sgv",
		"created_at": "2024-12-09T20:41:45.284Z",
	}
	return model.Record{
		Collection:  "entries",
		ID:          fmt.Sprintf("%d", sgv.ID),
		Data:        data,
		SrvCreated:  sgv.CreatedAt,
		SrvModified: sgv.CreatedAt,
		Subject:     "",
		IsValid:     true,
		DeletedAt:   nil,
	}
}

func CreateRecordSlice[T any](in []T, fn func(T) model.Record) []model.Record {
	out := make([]model.Record, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
