package nsutils

import (
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

// Document is one stored record as the Nightscout API writes it, whatever collection it is from
type Document interface {
	// Identifier is its public ID
	Identifier() string
	// Modified is when the server last changed it, in milliseconds, for Last-Modified and API v3's lastModified
	Modified() int64
	// OccurredAt is when it happened, in milliseconds
	OccurredAt() int64
	// Fields is it as JSON fields: API v3's with the server's fields, or v1's with an _id
	Fields(v3 bool) map[string]any
}

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

func (r Record) Modified() int64 { return r.SrvModified }

func (r Record) OccurredAt() int64 { return r.OccuredAt }

// Fields converts the record into a JSON-friendly map.
//
// When version3 is true, server-managed metadata fields are included.
func (r Record) Fields(version3 bool) map[string]any {
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

// DateString is when the document happened, as Nightscout's dateString
func DateString(doc Document) string {
	return time.UnixMilli(doc.OccurredAt()).UTC().Format(time.RFC3339Nano)
}

// SelectFields projects a document to the requested fields.
func SelectFields(doc Document, fields []string, version3 bool) map[string]any {
	full := doc.Fields(version3)
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
