package nsutils

import (
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

// Record is the normalized internal representation used by the storage layer.
type Record struct {
	Collection  string
	ID          string
	Data        map[string]any
	SrvCreated  int64
	SrvModified int64
	UTCOffset   int64
	Subject     string
	DeletedAt   *int64
	OccuredAt   int64
}

// Identifier returns the stable public identifier for the record.
func (r Record) Identifier() string {
	return r.ID
}

// ToMap converts the record into a JSON-friendly map.
//
// When version3 is true, server-managed metadata fields are included.
func (r Record) ToMap(version3 bool) map[string]any {
	out := utils.DeepCloneMap(r.Data)

	out["created_at"] = time.UnixMilli(r.OccuredAt).UTC().Format(time.RFC3339Nano)
	out["utcOffset"] = r.UTCOffset
	id := r.Identifier()
	if version3 {
		out["identifier"] = id
		out["date"] = r.OccuredAt
		out["srvCreated"] = r.SrvCreated
		out["srvModified"] = r.SrvModified
		out["subject"] = r.Subject
		out["isValid"] = r.DeletedAt == nil
		out["isReadOnly"] = false
		delete(out, "_id")
	} else {
		out["_id"] = id
		delete(out, "identifier")
	}

	return out
}

func (r Record) OccuredAtString() string {
	return time.UnixMilli(r.OccuredAt).UTC().Format(time.RFC3339Nano)
}

// SelectFields projects a record to the requested fields.
func (r Record) SelectFields(fields []string, version3 bool) map[string]any {
	full := r.ToMap(version3)
	if len(fields) == 0 {
		return full
	}
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := full[field]; ok {
			out[field] = value
		}
	}
	if len(fields) > 0 {
		delete(out, "_id")
	}
	return out
}

type ConverterToRecordFN[T any, PT interface {
	*T
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
	GetRawJsonData() *string
}] func(PT, *httputils.SiteContext, bool) Record

func DBModelToRecord[T any, PT interface {
	*T
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
	GetUtcOffsetMinutes() *int64
	GetOccurredAt() int64
	GetRawJsonData() *string
}, S any](collectionName Collection, convertFunc func(PT, *httputils.SiteContext) S) ConverterToRecordFN[T, PT] {
	return func(dbModel PT, siteContext *httputils.SiteContext, raw bool) Record {
		if dbModel == nil {
			return Record{}
		}

		var data map[string]any

		rawJSONData := dbModel.GetRawJsonData()
		if raw && rawJSONData != nil {
			err := json.Unmarshal([]byte(*rawJSONData), &data)
			if err != nil {
				return Record{}
			}
		} else {
			nsModel := convertFunc(dbModel, siteContext)

			nsModelData, err := utils.ToMap(&nsModel)
			if err != nil {
				return Record{}
			}
			data = nsModelData
		}

		utcOffset := int64(0)
		if dbModel.GetUtcOffsetMinutes() != nil {
			utcOffset = *dbModel.GetUtcOffsetMinutes()
		}

		record := Record{
			Collection:  string(collectionName),
			ID:          dbModel.GetID(),
			Data:        data,
			SrvCreated:  dbModel.GetCreatedAt(),
			SrvModified: dbModel.GetUpdatedAt(),
			Subject:     dbModel.GetCreatedByIdentityName(),
			DeletedAt:   dbModel.GetDeletedAt(),
			UTCOffset:   utcOffset,
			OccuredAt:   dbModel.GetOccurredAt(),
		}
		return record
	}
}

func CreateRecordSlice[T any](in []T, fn func(T, *httputils.SiteContext, bool) Record, siteContext *httputils.SiteContext, raw bool) []Record {
	out := make([]Record, len(in))
	for i, v := range in {
		out[i] = fn(v, siteContext, raw)
	}
	return out
}
