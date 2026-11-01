package nscollections

import (
	"maps"
	"time"

	"nighttroop.t1d.nz/utils"
)

// stored is a collection's Nightscout model, or the JSON it was sent as when that is asked for, with the server's
// fields from its row. It is only made into JSON fields when it is written.
type stored[NS any] struct {
	meta
	model NS
	raw   map[string]any
}

// meta are the fields the server keeps for every document
type meta struct {
	id                      string
	srvCreated, srvModified int64
	occurredAt, utcOffset   int64
	subject                 string
	deletedAt               *int64
}

func (m meta) Identifier() string { return m.id }
func (m meta) Modified() int64    { return m.srvModified }
func (m meta) OccurredAt() int64  { return m.occurredAt }

// row is a database model as a document's server fields, and the JSON it was sent as
func row(dbModel SharedDBFields) (meta, *string) {
	m := meta{id: dbModel.GetID(), srvCreated: dbModel.GetCreatedAt(), srvModified: dbModel.GetUpdatedAt(),
		occurredAt: dbModel.GetOccurredAt(), subject: dbModel.GetCreatedByIdentityName(), deletedAt: dbModel.GetDeletedAt()}
	if offset := dbModel.GetUtcOffsetMinutes(); offset != nil {
		m.utcOffset = *offset
	}
	return m, dbModel.GetRawJsonData()
}

func (s stored[NS]) Fields(v3 bool) map[string]any {
	out := maps.Clone(s.raw)
	if out == nil {
		out, _ = utils.ToMap(&s.model)
	}
	if out == nil {
		out = map[string]any{}
	}

	out["created_at"] = time.UnixMilli(s.occurredAt).UTC().Format(time.RFC3339Nano)
	out["utcOffset"] = s.utcOffset
	if v3 {
		out["identifier"] = s.id
		out["date"] = s.occurredAt
		out["srvCreated"] = s.srvCreated
		out["srvModified"] = s.srvModified
		out["subject"] = s.subject
		out["isValid"] = s.deletedAt == nil
		out["isReadOnly"] = false
		delete(out, "_id")
	} else {
		out["_id"] = s.id
		delete(out, "identifier")
	}
	return out
}
