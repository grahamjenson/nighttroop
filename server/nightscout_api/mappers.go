package nightscout_api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/model"
)

func SGVToModel(sgv *user.Sgv) model.Record {
	if sgv == nil {
		return model.Record{}
	}
	entryModel := model.Entry{
		Identifier: fmt.Sprintf("%d", sgv.ID),
		Type:       "sgv",
		SGV:        sgv.Value,
		Date:       sgv.CreatedAt,
		DateString: dateString(sgv.CreatedAt),
		Device:     fmt.Sprintf("%d", sgv.CreatedByIdentityID),
	}

	return model.Record{
		Collection:  "entries",
		ID:          fmt.Sprintf("%d", sgv.ID),
		Data:        StructToMap(entryModel),
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

func StructToMap(in any) map[string]any {
	b, err := json.Marshal(in)
	if err != nil {
		slog.Error("Failed to marshal ns model to JSON", "error", err)
		return map[string]any{}
	}

	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		slog.Error("Failed to unmarshal ns model from JSON", "error", err)
		return map[string]any{}
	}

	return m
}

func dateString(t int64) string {
	return time.UnixMilli(t).UTC().Format(time.RFC3339Nano)
}
