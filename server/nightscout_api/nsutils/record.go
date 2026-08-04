package nsutils

import (
	"time"

	"nighttroop.t1d.nz/utils"
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
	} else {
		out["_id"] = id
	}

	return out
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
}] func(PT) Record

func DBModelToRecord[T any, PT interface {
	*T
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
	GetUtcOffsetMinutes() *int64
	GetOccurredAt() int64
}, S any](collectionName Collection, convertFunc func(PT) S) ConverterToRecordFN[T, PT] {
	return func(dbModel PT) Record {
		if dbModel == nil {
			return Record{}
		}

		nsModel := convertFunc(dbModel)

		mapModel, err := utils.ToMap(&nsModel)
		if err != nil {
			return Record{}
		}
		utcOffset := int64(0)
		if dbModel.GetUtcOffsetMinutes() != nil {
			utcOffset = *dbModel.GetUtcOffsetMinutes()
		}

		record := Record{
			Collection:  string(collectionName),
			ID:          dbModel.GetID(),
			Data:        mapModel,
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

func CreateRecordSlice[T any](in []T, fn func(T) Record) []Record {
	out := make([]Record, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
