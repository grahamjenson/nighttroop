package nsutils

import (
	"fmt"

	"nighttroop.t1d.nz/utils"
)

// Record is the normalized internal representation used by the storage layer.
type Record struct {
	Collection  string
	ID          string
	Data        map[string]any
	SrvCreated  int64
	SrvModified int64
	Subject     string
	IsValid     bool
	DeletedAt   *int64
}

// Clone returns a deep copy of the record and its payload map.
func (r Record) Clone() Record {
	clone := r
	clone.Data = CloneMap(r.Data)
	return clone
}

// Identifier returns the stable public identifier for the record.
func (r Record) Identifier() string {
	if r.Data == nil {
		return r.ID
	}
	keys := []string{"identifier", "_id"}
	// Just checking these values first
	for _, k := range keys {
		if v, ok := r.Data[k]; ok && v != nil {
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}

	return r.ID
}

// WithData returns a cloned record with replaced payload data.
func (r Record) WithData(data map[string]any) Record {
	clone := r.Clone()
	clone.Data = CloneMap(data)
	return clone
}

// ToMap converts the record into a JSON-friendly map.
//
// When version3 is true, server-managed metadata fields are included.
func (r Record) ToMap(version3 bool) map[string]any {
	out := CloneMap(r.Data)
	id := r.Identifier()
	if version3 {
		delete(out, "_id")
		out["identifier"] = id
		out["srvCreated"] = r.SrvCreated
		out["srvModified"] = r.SrvModified
		out["subject"] = r.Subject
		out["isValid"] = r.IsValid
		if r.DeletedAt != nil {
			out["deletedAt"] = *r.DeletedAt
		}
	} else {
		out["_id"] = id
		delete(out, "identifier")
	}

	return out
}

// SelectFields projects a record to the requested fields.
func (record Record) SelectFields(fields []string, version3 bool) map[string]any {
	full := record.ToMap(version3)
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

		record := Record{
			Collection:  string(collectionName),
			ID:          dbModel.GetID(),
			Data:        mapModel,
			SrvCreated:  dbModel.GetCreatedAt(),
			SrvModified: dbModel.GetUpdatedAt(),
			Subject:     dbModel.GetCreatedByIdentityName(),
			IsValid:     true,
			DeletedAt:   dbModel.GetDeletedAt(),
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
